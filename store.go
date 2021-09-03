package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/rpc"
	"os"
	"sync"
)

// URLStore 结构体
type URLStore struct {
	urls map[string]string
	mu   sync.RWMutex
	save chan record
}

type ProxyStore struct {
	urls   *URLStore
	client *rpc.Client
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
	}

	if filename != "" {
		s.save = make(chan record, saveQueueLength)
		if err := s.load(filename); err != nil {
			log.Println("Error loading URLStore:", err)
		}
		go s.saveLoop(filename)
	}
	return s
}

func NewProxyStore(addr string) *ProxyStore {
	client, err := rpc.DialHTTP("tcp", addr)

	if err != nil {
		log.Printf("Error constructing ProxyStore", err)
	}

	return &ProxyStore{urls: NewURLStore(""), client: client}
}

func (s *ProxyStore) Get(key, url *string) error {
	if err := s.urls.Get(key, url); err == nil {
		return nil
	}
	if err := s.client.Call("Store.Get", key, url); err != nil {
		return err
	}
	s.urls.Set(key, url)
	return nil
}

func (s *ProxyStore) Put(url, key *string) error {
	if err := s.client.Call("Store.Put", url, key); err != nil {
		return err
	}

	s.urls.Set(key, url)
	return nil
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
func (s *URLStore) Get(key, url *string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if u, ok := s.urls[*key]; ok {
		*url = u
		return nil
	}

	return errors.New("key not found")
}

// Set 设置长链接方法
func (s *URLStore) Set(key, url *string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, present := s.urls[*key]; present {
		return errors.New("key already exists")
	}

	s.urls[*key] = *url

	return nil
}

// Count 链接个数
func (s *URLStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.urls)
}

// Put 执行存储
func (s *URLStore) Put(url, key *string) error {
	for {
		*key = genKey(s.Count())
		if err := s.Set(key, url); err == nil {
			break
		}
	}
	if s.save != nil {
		s.save <- record{*key, *url}
	}
	return nil
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
			s.Set(&r.Key, &r.URL)
		}
	}

	if err == io.EOF {
		return nil
	}

	log.Println("Error decoding URLStore:", err)
	return err
}
