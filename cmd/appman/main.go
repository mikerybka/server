package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/library-development/go-golang"
	"github.com/library-development/go-systemd"
)

// The server's application manager.

const configfile = "/etc/reverseproxy/config.json"

func main() {
	w, _ := golang.SetupWorkdir("/src")
	http.ListenAndServe(":54321", &Manager{
		ConfigFile: configfile,
		GoWorkdir:  w,
		NextPort:   19801,
	})
}

type Manager struct {
	ConfigFile string
	GoWorkdir  golang.Workdir
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

type AddAppRequest struct {
	Repo struct {
		Name  string
		Owner string
	}
	Path string
}
type AddAppResponse struct {
	Port  string
	Error string
}

func (r *AddAppResponse) Write(w http.ResponseWriter) {
	b, err := json.Marshal(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(b)
}

func (m *Manager) AddApp(w http.ResponseWriter, r *http.Request) {
	// Parse the request.
	var req AddAppRequest
	var res AddAppResponse
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		res.Error = err.Error()
		res.Write(w)
		return
	}
	// Pull the latest code for the app.
	err = m.GoWorkdir.Pull(req.Repo.Owner, req.Repo.Name)
	if err != nil {
		res.Error = err.Error()
		res.Write(w)
		return
	}
	// Find a free port.
	port := m.findPort()
	// Build the app and place the binary in a file named after the port.
	appID := "github.com/" + req.Repo.Owner + "/" + req.Repo.Name
	bin := "/apps/" + port
	err = m.GoWorkdir.Build(appID, bin)
	if err != nil {
		res.Error = err.Error()
		res.Write(w)
		return
	}
	// Create a systemd service for the app.
	err = systemd.AddService("app-"+port, appID, bin+" "+port)
	if err != nil {
		res.Error = err.Error()
		res.Write(w)
		// Clean up the binary and the systemd service.
		os.Remove(bin)
		return
	}
	// Return the port to the client.
	res.Port = port
	res.Write(w)
}

type SetDomainRequest struct {
	Domain string
	Port   string
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
