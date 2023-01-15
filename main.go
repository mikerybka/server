package main

import (
	"fmt"
	"net/http"

	"lib.dev/libdev"
	"lib.dev/web"
)

var apps = map[string]http.Handler{
	"lib.dev": &libdev.App{},
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
