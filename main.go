package main

import (
	"flag"
	"fmt"
	"net/http"
)

var (
	listenAddr = flag.String("http", ":8888", "http listen address")
	dataFile   = flag.String("file", "store.json", "data store file name")
	hostname   = flag.String("host", "localhost:8888", "http host name")
)

// AddForm 页面HTML
const AddForm = `
	<form method="POST" action="/add">
	URL: <input type="text" name="url">
		<input type="submit" value="Add">
	</form>
`

var store *URLStore

func main() {
	flag.Parse()
	store = NewURLStore(*dataFile)

	http.HandleFunc("/", Redirect)
	http.HandleFunc("/add", Add)
	http.ListenAndServe(*listenAddr, nil)
}

// Redirect 短连接重定向
func Redirect(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path[1:]
	url := store.Get(key)

	if url == "" {
		http.NotFound(w, r)
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

	key := store.Put(url)
	fmt.Fprintf(w, "http://%s/%s", *hostname, key)
}
