package main

import (
	"fmt"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mikerybka/constants"
	"github.com/mikerybka/secretdb"
	"github.com/mikerybka/util"
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

	recordError(updateSystem())

	// Run system
	backendURL, err := url.Parse(fmt.Sprintf("http://%s:3000", constants.BackendIP))
	if err != nil {
		panic(err)
	}
	handler := httputil.NewSingleHostReverseProxy(backendURL)
	certDir := filepath.Join(workdir(), "certs")
	secrets := secretdb.NewClient(constants.BackendIP, "mike", "1212")
	email, _ := secrets.Email()
	allowHost := func(host string) bool {
		return len(strings.Split(host, ".")) <= 4
	}
	err = util.ServeHTTPS(handler, email, certDir, allowHost)
	if err != nil {
		fmt.Println(err)
	}
}

func updateSystem() error {
	// Run: apt update
	cmd := exec.Command("apt", "update")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("apt update: %s", out)
	}

	// Run: apt upgrade -y
	cmd = exec.Command("apt", "upgrade", "-y")
	out, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("apt upgrade: %s", out)
	}

	// Update Go
	//TODO

	// Update binaries
	//TODO

	return nil
}

func workdir() string {
	return filepath.Join(util.HomeDir(), ".local/share/server")
}

func recordError(err error) {
	if err != nil {
		path := filepath.Join(util.HomeDir(), "errors", util.UnixNanoTimestamp())
		err = util.WriteFile(path, []byte(err.Error()))
		if err != nil {
			panic(err)
		}
	}
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
