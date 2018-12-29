package api

import (
	"encoding/json"
	"fmt"
	"gateway/errors"
	"net/http"
)

func (api *API) HandleAppsCLRUD(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		api.HandleAppsAdd(w, r)
	}
}

func (api *API) HandleAppsAdd(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		errors.NewBadRequest("empty request body").Write(w)
		return
	}

	var d map[string]string
	err := json.NewDecoder(r.Body).Decode(&d)
	if err != nil {
		errors.NewBadRequest(err).Write(w)
		return
	}

	for name, version := range d {
		fmt.Printf("%s: %s\n", name, version)
		if _, err := api.Registry.Add(name, version); err != nil {
			errors.NewInternalServerError(err).Println().Write(w)
		}
	}
}
