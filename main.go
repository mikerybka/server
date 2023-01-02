package main

import (
	"flag"
	"log"

	"lib.dev/web"
)

func main() {
	flag.Parse()
	platform := &web.Platform{
		DataDir:          flag.Arg(0),
		LetsEncryptEmail: flag.Arg(1),
	}
	err := platform.Start()
	if err != nil {
		log.Fatal(err)
	}
}
