package main

import (
	"db"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

func APIError(w http.ResponseWriter, err error) {
	log.Println(err)
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(fmt.Sprintf("%+v", err)))
}

func APIErrorWithStatus(w http.ResponseWriter, err error, statusCode int) {
	w.WriteHeader(statusCode)
	w.Write([]byte(fmt.Sprintf("%+v", err)))
}

func APIJSON(w http.ResponseWriter, body interface{}) {
	if data, err := json.Marshal(body); err != nil {
		APIError(w, err)
		return
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}

func api() http.Handler {
	r := mux.NewRouter()
	r.KeepContext = true
	r.HandleFunc("/ca", ServeCACertificate).Methods("GET")
	r.HandleFunc("/api/v1/gateway/{id}/certificate/", CreateCertificate).Methods("POST")
	r.HandleFunc("/api/v1/gateway/", CreateGateway).Methods("POST")
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 Not Found"))
	})
	return r
}

func private() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/api/v1/gateway/", UpdateGateway).Methods("PUT")
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 Not Found"))
	})
	return r
}

type CreateGatewayRequest struct {
	Port uint16 `json:"port"`
}

type CreateGatewayResponse struct {
	ID   string `json:"id"`
	FQDN string `json:"fqdn"`
}

func CreateGateway(w http.ResponseWriter, r *http.Request) {
	var request CreateGatewayRequest

	if body, err := ioutil.ReadAll(r.Body); err != nil {
		APIError(w, err)
		return
	} else if err := json.Unmarshal([]byte(body), &request); err != nil {
		APIError(w, err)
		return
	}

	addr := strings.Split(r.RemoteAddr, ":")[0]

	if gw, err := db.CreateGateway(addr, request.Port); err != nil {
		APIError(w, err)
		return
	} else {
		r := CreateGatewayResponse{
			ID:   gw.ID,
			FQDN: gw.FQDNWithoutDot(),
		}
		APIJSON(w, r)
		return
	}
}

func UpdateGateway(w http.ResponseWriter, r *http.Request) {
	var request CreateGatewayRequest

	if body, err := ioutil.ReadAll(r.Body); err != nil {
		APIError(w, err)
		return
	} else if err := json.Unmarshal([]byte(body), &request); err != nil {
		APIError(w, err)
		return
	}

	gw := r.Context().Value(GatewayContextKey).(*db.Gateway)
	gw.Port = request.Port

	if err := gw.Update(); err != nil {
		APIError(w, err)
		return
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte{})
		return
	}
}
