package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "setup" {
		err := setup()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		return
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	})
	http.ListenAndServe(":2222", nil)
}

var systemdServiceFile = []byte(`[Unit]
Description=Server
After=network.target

[Service]
Type=simple
ExecStart=/bin/server

[Install]
WantedBy=multi-user.target
`)

func setup() error {
	switch runtime.GOOS {
	case "linux":
		if usesSystemd() {
			return setupSystemd()
		} else {
			return fmt.Errorf("system not supported")
		}
	default:
		return fmt.Errorf("system not supported")
	}
}

func setupSystemd() error {
	err := os.WriteFile("/etc/systemd/system/server.service", systemdServiceFile, os.ModePerm)
	if err != nil {
		return fmt.Errorf("writing systemd unit file: %s", err)
	}

	cmd := exec.Command("systemctl", "daemon-reload")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("reloading systemd daemon: %s", err)
	}

	cmd = exec.Command("systemctl", "enable", "server")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("enabling systemd unit: %s", err)
	}

	cmd = exec.Command("reboot")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("rebooting: %s", err)
	}

	return nil
}

func usesSystemd() bool {
	data, err := os.ReadFile("/proc/1/comm")
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(data)) == "systemd"
}
