package main

import (
	"net/http"

	"github.com/library-development/go-golang"
	"github.com/mikerybka/server/pkg/appman"
)

// The server's application manager.

const configfile = "/etc/reverseproxy/config.json"

func main() {
	w, _ := golang.SetupWorkdir("/src")
	http.ListenAndServe(":54321", &appman.Manager{
		ConfigFile: configfile,
		GoWorkdir:  w,
		NextPort:   19801,
	})
}
