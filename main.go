package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/mikerybka/git"
	"github.com/mikerybka/golang"
	"github.com/mikerybka/secretdb"
	"github.com/mikerybka/util"
)

var reboot = false

func main() {
	// Handle `server setup` command
	if len(os.Args) > 1 && os.Args[1] == "setup" {
		err := setup()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		return
	}

	// Full system update (may restart)
	recordError(updateSystem())

	// Reboot at midnight
	go func() {
		util.WaitUntil(0, 0, 0)
		reboot = true
		time.Sleep(10 * time.Second) // wait for any requests to finish
		cmd := exec.Command("reboot", "now")
		cmd.Run()
	}()

	// Run system
	backendURL, err := url.Parse("http://100.115.153.54:3333")
	if err != nil {
		panic(err)
	}
	secrets := secretdb.NewClient("http://100.115.153.54:2222/secrets")
	radioPassword := util.Must(secrets.Read("radio.mikerybka.com/password"))
	broadcaster := util.NewStreamBroadcaster(func() (io.ReadCloser, error) {
		req, err := http.NewRequest("GET", "http://100.115.153.54:3333/live.mp3", nil)
		if err != nil {
			return nil, err
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		if res.StatusCode != 200 {
			return nil, fmt.Errorf("%s", res.Status)
		}
		return res.Body, nil
	}).HTTPHandler("audio/mpeg")
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if reboot {
			return
		}
		mux := http.NewServeMux()
		mux.HandleFunc("mikerybka.com/{$}", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "text/html")
			w.Write([]byte(`<pre><a href="https://radio.mikerybka.com">Radio</a>
	</pre>`))
		})
		mux.Handle("GET /live.mp3", util.PasswordProtected(radioPassword, http.HandlerFunc(broadcaster)))
		mux.Handle("/", httputil.NewSingleHostReverseProxy(backendURL))
		mux.ServeHTTP(w, r)
	})
	certDir := filepath.Join(workdir(), "certs")
	err = os.MkdirAll(certDir, os.ModePerm)
	if err != nil {
		fmt.Println(err)
		return
	}
	email, _ := secrets.Read("email")
	allowHost := func(host string) bool {
		return len(strings.Split(host, ".")) <= 4
	}
	err = notifyAdmin(upAndRunningMessage())
	recordError(err)
	err = util.ServeHTTPS(h, email, certDir, allowHost)
	if err != nil {
		fmt.Println(err)
	}
}

func upAndRunningMessage() string {
	msgs := []string{
		"Up and running!",
		"Good to go!",
		"Deployment a success!",
		"Another good one!",
		"Boom!",
		"Bam!",
		"Skiddy bop!",
	}
	msg, err := util.RandomElement(msgs)
	if err != nil {
		panic(err)
	}
	return msg
}

func notifyAdmin(msg string) error {
	req, err := http.NewRequest(http.MethodPost, "http://100.115.153.54:2222/alert", bytes.NewReader([]byte(msg)))
	if err != nil {
		panic(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		b, _ := io.ReadAll(res.Body)
		return fmt.Errorf("%s: %s", res.Status, b)
	}
	return nil
}

func updateSystem() error {
	// Run: apt update
	fmt.Println("Updating package lists")
	cmd := exec.Command("apt", "update")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("apt update: %s", out)
	}

	// Run: apt upgrade -y
	fmt.Println("Upgrading packages")
	cmd = exec.Command("apt", "upgrade", "-y")
	out, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("apt upgrade: %s", out)
	}

	// Update Go
	fmt.Println("Updating Go")
	rebuild, err := golang.InstallOrUpdateGo()
	if err != nil {
		return err
	}

	// Setup workspace
	fmt.Println("Configuring Go workspace")
	srcDir := filepath.Join(workdir(), "src")
	err = os.MkdirAll(srcDir, os.ModePerm)
	if err != nil {
		return err
	}
	w := &golang.Workspace{
		Dir: srcDir,
	}
	err = w.Init()
	if err != nil {
		return err
	}

	// Update binaries
	libraries := []string{
		"util",
		"brass",
		"english",
		"git",
		"golang",
		"secretdb",
	}
	binaries := []string{
		"server",
	}
	for _, lib := range libraries {
		pkg := fmt.Sprintf("github.com/mikerybka/%s", lib)
		fmt.Println("Updating", pkg)
		dir := filepath.Join(w.Dir, pkg)
		url, err := golang.PkgGitURL(pkg)
		if err != nil {
			panic(err)
		}
		changed, err := git.CloneOrPull(dir, url)
		if err != nil {
			return err
		}
		if changed {
			rebuild = true
		}
		cmd = exec.Command("go", "work", "use", ".")
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("%s: %s", cmd, out)
		}
	}
	for _, bin := range binaries {
		pkg := fmt.Sprintf("github.com/mikerybka/%s", bin)
		fmt.Println("Updating", pkg)
		dir := filepath.Join(w.Dir, pkg)
		url, err := golang.PkgGitURL(pkg)
		if err != nil {
			panic(err)
		}
		changed, err := git.CloneOrPull(dir, url)
		if err != nil {
			return err
		}
		cmd = exec.Command("go", "work", "use", ".")
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("%s: %s", cmd, out)
		}
		if changed || rebuild {
			fmt.Println("Building", bin)
			err = build(bin)
			if err != nil {
				return err
			}
		} else {
			fmt.Println("No changes")
		}
	}

	if rebuild {
		fmt.Println("Rebooting")
		cmd = exec.Command("reboot")
		out, err = cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("reboot: %s", out)
		}
	}

	return nil
}

func binPath(bin string) string {
	return filepath.Join("/bin", bin)
}

func build(bin string) error {
	workspace := &golang.Workspace{
		Dir: filepath.Join(workdir(), "src"),
	}
	pkg := fmt.Sprintf("github.com/mikerybka/%s", bin)
	return workspace.Build(pkg, binPath(bin))
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
Environment=PATH="/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/games:/usr/local/games:/snap/bin"
Environment=GOMODCACHE="/root/.local/share/server/go/mod"
Environment=GOCACHE="/root/.local/share/server/go/cache"

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
