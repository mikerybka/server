package main

import (
	"flag"
	"net/http"
)

func main() {
	flag.Parse()
	port := flag.Arg(0)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(r.Host + r.URL.Path))
	})
	http.ListenAndServe(":"+port, nil)
}
