package main

import (
	"fmt"
	"net/http"

	"lib.dev/libdev"
	"lib.dev/web"
)

func main() {
	libdevApp := libdev.App{}
	p := web.Platform{
		LetsEncryptEmail: "merybka@gmail.com",
		CertDir:          "/root/certs",
		Apps: map[string]http.Handler{
			"lib.dev": &libdevApp,
		},
	}
	err := p.Start()
	if err != nil {
		fmt.Println(err)
	}
}
