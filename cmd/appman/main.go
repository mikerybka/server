package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/library-development/go-web"
	"github.com/mikerybka/server/pkg/appman"
)

func main() {
	flag.Parse()
	action := flag.Arg(0)
	switch action {
	case "add-app":
		port, err := addApp(flag.Arg(1))
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Println("Port:", port)
	case "set-domain":
		err := setDomain(flag.Arg(1), flag.Arg(2))
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
	}
}

func addApp(appID string) (string, error) {
	fmt.Println("Adding app:", appID)
	path := web.ParsePath("/" + appID)
	json.NewEncoder(os.Stdout).Encode(path)
	if !(path.First() == "github.com") {
		return "", fmt.Errorf("appID must start with github.com/")
	}
	b, err := json.Marshal(appman.AddAppRequest{
		Repo: struct {
			Name  string
			Owner string
		}{
			Owner: path[1],
			Name:  path[2],
		},
		Path: path[3:].String(),
	})
	if err != nil {
		panic(err)
	}
	resp, err := http.Post("http://localhost:54321/add-app", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var r appman.AddAppResponse
	json.NewDecoder(resp.Body).Decode(&r)
	if r.Error != "" {
		return "", fmt.Errorf(r.Error)
	}
	return r.Port, nil
}

func setDomain(domain string, port string) error {
	b, err := json.Marshal(appman.SetDomainRequest{Domain: domain, Port: port})
	if err != nil {
		panic(err)
	}
	resp, err := http.Post("http://localhost:54321/set-domain", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var r appman.SetDomainResponse
	json.NewDecoder(resp.Body).Decode(&r)
	if r.Error != "" {
		return fmt.Errorf(r.Error)
	}
	return nil
}
