package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func (api *API) HandleAppsCLRUD(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		api.HandleAppsAdd(w, r)
	}
}

func (api *API) HandleAppsAdd(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		api.Error(w, fmt.Errorf("empty request body"), 400)
		return
	}

	var d map[string]string
	err := json.NewDecoder(r.Body).Decode(&d)
	if err != nil {
		api.Error(w, err, 400)
		return
	}

	for name, version := range d {
		fmt.Printf("%s: %s\n", name, version)
		if _, err := api.Registry.Add(name, version); err != nil {
			log.Fatal(err)
		}
	}
}
