package api

import (
	"net/http"
)

type API struct {
}

func (api *API) Error(w http.ResponseWriter, err error, code int) {
	http.Error(w, err.Error(), 400)
}

func NewAPI() (*API, error) {
	api := &API{}

	http.HandleFunc("/api/v1/apps", api.HandleAppsCLRUD)

	return api, nil
}
