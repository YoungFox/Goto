package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/rpc"
)

var (
	listenAddr = flag.String("http", ":8080", "http listen address")
	dataFile   = flag.String("file", "store.json", "data store file name")
	hostname   = flag.String("host", "localhost:8080", "http host name")
	rpcEnabled = flag.Bool("rpc", false, "enable RPC server")
	masterAddr = flag.String("master", "", "RPC master address")
)

// AddForm 页面HTML
const AddForm = `
	<form method="POST" action="/add">
	URL: <input type="text" name="url">
		<input type="submit" value="Add">
	</form>
`

type Store interface {
	Put(url, key *string) error
	Get(key, url *string) error
}

var store Store

func main() {
	flag.Parse()
	if *masterAddr != "" {
		store = NewProxyStore(*masterAddr)
	} else {
		store = NewURLStore(*dataFile)
	}

	if *rpcEnabled {
		rpc.RegisterName("Store", store)
		rpc.HandleHTTP()
	}

	http.HandleFunc("/", Redirect)
	http.HandleFunc("/add", Add)
	http.ListenAndServe(*listenAddr, nil)
}

// Redirect 短连接重定向
func Redirect(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path[1:]
	var url string
	if err := store.Get(&key, &url); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, url, http.StatusFound)
}

// Add 新增长连接
func Add(w http.ResponseWriter, r *http.Request) {
	url := r.FormValue("url")
	if url == "" {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, AddForm)
		return
	}

	var key string

	if err := store.Put(&url, &key); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "http://%s/%s", *hostname, key)
}
