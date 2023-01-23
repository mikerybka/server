package main

import (
	"net/http"

	"github.com/mikerybka/server/pkg/appman"
)

// The server's application manager.

const configfile = "/etc/reverseproxy/config.json"

func main() {
	http.ListenAndServe(":54321", &appman.Manager{
		ConfigFile: configfile,
		NextPort:   19801,
	})
}
