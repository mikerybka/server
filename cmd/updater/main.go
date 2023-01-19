package main

import (
	"fmt"
	"os/exec"
	"time"
)

func main() {
	for {
		time.Sleep(4 * time.Hour)
		err := update()
		if err != nil {
			fmt.Println(err)
		}
	}
}

func update() error {
	cmd := exec.Command("go", "install", "github.com/mikerybka/server/cmd/server@latest")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error: %s, output: %s", err, out)
	}
	cmd = exec.Command("systemctl", "restart", "server")
	out, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error: %s, output: %s", err, out)
	}
	return nil
}
