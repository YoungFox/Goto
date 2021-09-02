package main

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"sync"
)

// URLStore 结构体
type URLStore struct {
	urls map[string]string
	mu   sync.RWMutex
	save chan record
}

// record 存储结构体
type record struct {
	Key, URL string
}

var saveQueueLength = 1000

// NewURLStore 工厂函数
func NewURLStore(filename string) *URLStore {
	s := &URLStore{
		urls: make(map[string]string),
		save: make(chan record, saveQueueLength)}

	if err := s.load(filename); err != nil {
		log.Println("Error loading URLStore", err)
	}

	go s.saveLoop(filename)
	return s
}

func (s *URLStore) saveLoop(filename string) {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal("URLStore:", err)
	}
	defer f.Close()

	e := json.NewEncoder(f)

	for {
		r := <-s.save
		if err := e.Encode(r); err != nil {
			log.Println("URLStore:", err)
		}
	}
}

// Get 获取长链接方法
func (s *URLStore) Get(key string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.urls[key]
}

// Set 设置长链接方法
func (s *URLStore) Set(key, url string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, present := s.urls[key]; present {
		return false
	}

	s.urls[key] = url

	return true
}

// Count 链接个数
func (s *URLStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.urls)
}

// Put 执行存储
func (s *URLStore) Put(url string) string {
	for {
		key := genKey(s.Count())
		if s.Set(key, url) {
			s.save <- record{key, url}
			return key
		}
	}
	// 不可执行到这
	panic("shouldn’t get here")
}

func (s *URLStore) load(filename string) error {
	f, err := os.Open(filename)

	if err != nil {
		log.Println("Error opening URLStore", err)
		return err
	}

	defer f.Close()

	d := json.NewDecoder(f)
	for err == nil {
		var r record
		if err = d.Decode(&r); err == nil {
			s.Set(r.Key, r.URL)
		}
	}

	if err == io.EOF {
		return nil
	}

	log.Println("Error decoding URLStore:", err)
	return err
}
