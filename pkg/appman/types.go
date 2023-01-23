package appman

import (
	"encoding/json"
	"net/http"
)

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

type SetDomainRequest struct {
	Domain string
	Port   string
}
type SetDomainResponse struct {
	Error string
}

func (r *SetDomainResponse) Write(w http.ResponseWriter) {
	b, err := json.Marshal(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(b)
}
