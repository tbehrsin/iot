package main

import (
	"db"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"

	"github.com/gorilla/mux"
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
	r.HandleFunc("/api/v1/gateway/local", SetLocalAddress).Methods("PUT")
	r.HandleFunc("/api/v1/gateway/", UpdateGateway).Methods("PUT")
	r.HandleFunc("/api/v1/auth/", CreateEmailToken).Methods("POST")
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 Not Found"))
	})
	return r
}

func ServeCACertificate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet || r.URL.Path != "/ca.crt" {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 Not Found"))
		return
	}

	if ca, err := db.GetCA(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("503 Internal Server Error"))
	} else {
		w.Header()["Content-Type"] = []string{"application/x-x509-ca-cert"}
		w.WriteHeader(http.StatusOK)
		pem.Encode(w, &pem.Block{Type: "CERTIFICATE", Bytes: ca.Certificate.Raw})
	}
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

	addr, _, _ := net.SplitHostPort(r.RemoteAddr)

	if gw, err := db.CreateGateway(addr, int(request.Port)); err != nil {
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
	gw.Address, _, _ = net.SplitHostPort(r.RemoteAddr)
	gw.Port = int(request.Port)

	if err := gw.Update(); err != nil {
		APIError(w, err)
		return
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte{})
		return
	}
}

type SetLocalAddressRequest struct {
	LocalAddress string `json:"localAddress"`
}

type SetLocalAddressResponse struct {
}

func SetLocalAddress(w http.ResponseWriter, r *http.Request) {
	var request SetLocalAddressRequest

	if body, err := ioutil.ReadAll(r.Body); err != nil {
		APIError(w, err)
		return
	} else if err := json.Unmarshal([]byte(body), &request); err != nil {
		APIError(w, err)
		return
	}

	gw := r.Context().Value(GatewayContextKey).(*db.Gateway)
	gw.LocalAddress = request.LocalAddress

	if err := gw.Update(); err != nil {
		APIError(w, err)
		return
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte{})
		return
	}
}
