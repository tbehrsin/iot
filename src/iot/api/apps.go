package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (api *API) HandleAppsCLRUD(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// Create
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

}
