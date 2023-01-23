package appman

import (
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/library-development/go-systemd"
)

type Manager struct {
	ConfigFile string
	NextPort   int
	configLock sync.Mutex
	portLock   sync.Mutex
}

func (m *Manager) Config() (map[string]string, error) {
	domains := make(map[string]string)
	b, err := os.ReadFile(m.ConfigFile)
	if err != nil {
		return domains, err
	}
	err = json.Unmarshal(b, &domains)
	if err != nil {
		return domains, err
	}
	return domains, nil
}

func (m *Manager) AddApp(w http.ResponseWriter, r *http.Request) {
	var req AddAppRequest
	var res AddAppResponse
	// Parse the request.
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		res.Error = err.Error()
		res.Write(w)
		return
	}
	// Install the latest version
	appID := "github.com/" + req.Repo.Owner + "/" + req.Repo.Name + req.Path
	cmd := exec.Command("go", "install", appID+"@latest")
	cmd.Env = append(os.Environ(), "GOPROXY=direct")
	out, err := cmd.CombinedOutput()
	if err != nil {
		res.Error = err.Error() + ": " + string(out)
		res.Write(w)
		return
	}
	// Find a free port.
	port := m.findPort()
	// Create a systemd service for the app.
	err = systemd.AddService("app-"+port, appID+" on port "+port, "/root/go/bin/"+filepath.Base(appID)+" "+port)
	if err != nil {
		res.Error = err.Error()
		res.Write(w)
		return
	}
	// Return the port to the client.
	res.Port = port
	res.Write(w)
}

func (m *Manager) SetDomain(w http.ResponseWriter, r *http.Request) {
	// Parse input.
	var req SetDomainRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Aquire the lock.
	m.configLock.Lock()
	defer m.configLock.Unlock()

	// Check if the domain already exists.
	domains, err := m.Config()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, ok := domains[req.Domain]
	if ok {
		http.Error(w, "domain already exists", http.StatusBadRequest)
		return
	}

	// Update the config file.
	domains[req.Domain] = req.Port
	b, err := json.Marshal(domains)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = os.WriteFile(m.ConfigFile, b, 0644)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (m *Manager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		m.Root(w, r)
	case "/add-app":
		m.AddApp(w, r)
	case "/set-domain":
		m.SetDomain(w, r)
	}
}

func (m *Manager) Root(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, m.ConfigFile)
}

func (m *Manager) findPort() string {
	m.portLock.Lock()
	defer m.portLock.Unlock()
	port := m.NextPort
	m.NextPort++
	return strconv.Itoa(port)
}
