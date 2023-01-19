package main

import (
	"fmt"
	"net/http"

	"github.com/library-development/go-web"
	"github.com/mikerybka/go-libdev"
	"github.com/mikerybka/go-mikerybkadev"
	"github.com/mikerybka/go-schemacafe"
)

var apps = map[string]http.Handler{
	"lib.dev":       &libdev.App{},
	"mikerybka.dev": &mikerybkadev.App{},
	"schema.cafe":   &schemacafe.App{},
}

func main() {
	p := web.Platform{
		LetsEncryptEmail: "merybka@gmail.com",
		CertDir:          "/root/certs",
		Apps:             apps,
	}
	err := p.Start()
	if err != nil {
		fmt.Println(err)
	}
}
