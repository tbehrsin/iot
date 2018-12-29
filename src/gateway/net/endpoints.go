package net

import (
	"encoding/json"
	"fmt"
	"gateway/errors"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

func (n *Network) CreateRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/api/v1/devices/{eui64:[0-9A-F]+}/public/", n.deviceFileServer).Methods("POST")
	r.PathPrefix("/api/v1/devices/{eui64:[0-9A-F]+}/public/").Methods("GET").HandlerFunc(n.deviceFileServer)
	r.HandleFunc("/api/v1/devices/{eui64:[0-9A-F]+}/", n.getDeviceHandler).Methods("GET")
	r.HandleFunc("/api/v1/devices/{eui64:[0-9A-F]+}/", n.updateDeviceHandler).Methods("PATCH")
	r.HandleFunc("/api/v1/devices/{eui64:[0-9A-F]+}/", n.deleteDeviceHandler).Methods("DELETE")
	r.HandleFunc("/api/v1/devices/", n.listDevicesHandler).Methods("GET")
	r.HandleFunc("/api/v1/devices/", n.createDeviceHandler).Methods("POST")
	return r
}

func (n *Network) listDevicesHandler(w http.ResponseWriter, r *http.Request) {
	devices := make([]deviceHolder, 0, 10)

	n.gateways.Range(func(k, v interface{}) bool {
		gateway := v.(*GatewayProxy)

		gateway.devices.Range(func(k, v interface{}) bool {
			deviceProxy := v.(*DeviceProxy)
			holder := deviceProxy.Holder()
			devices = append(devices, *holder)
			return true
		})

		return true
	})

	errors.APIJSON(w, devices)
}

func (n *Network) createDeviceHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("createDeviceHandler")
}

func (n *Network) getDeviceHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	if eui64, err := NewEUI64(vars["eui64"]); err != nil {
		errors.NewNotFound(err).Write(w)
	} else {
		written := false
		n.gateways.Range(func(k, v interface{}) bool {
			gateway := v.(*GatewayProxy)

			if d, ok := gateway.devices.Load(eui64); !ok {
				return true
			} else {
				deviceProxy := d.(*DeviceProxy)
				holder := deviceProxy.Holder()
				errors.APIJSON(w, holder)
				written = true
				return false
			}
		})

		if !written {
			errors.NewNotFound(err).Write(w)
		}
	}
}

func (n *Network) updateDeviceHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var obj map[string]interface{}
	defer r.Body.Close()

	if eui64, err := NewEUI64(vars["eui64"]); err != nil {
		errors.NewNotFound(err).Write(w)
	} else {
		written := false
		n.gateways.Range(func(k, v interface{}) bool {
			gateway := v.(*GatewayProxy)

			if d, ok := gateway.devices.Load(eui64); !ok {
				return true
			} else {
				deviceProxy := d.(*DeviceProxy)

				if b, err := ioutil.ReadAll(r.Body); err != nil {
					errors.NewBadRequest(err).Println().Write(w)
				} else if err := json.Unmarshal(b, &obj); err != nil {
					log.Println(string(b))
					errors.NewBadRequest(err).Println().Write(w)
				} else {
					if s, ok := obj["state"]; ok {
						deviceProxy.SetState(s.(map[string]interface{}))
					}
					// if n, ok := obj["name"]; ok {
					//   dp.SetName(n)
					// }

				}

				written = true
				return false
			}
		})
		if !written {
			errors.APIJSON(w, nil)
		}
	}
}

func (n *Network) deleteDeviceHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("deleteDeviceHandler")
}

func (n *Network) deviceFileServer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	if r.Method == http.MethodPost {
		cookie := &http.Cookie{
			Name: "token",
			Path: strings.TrimSuffix(r.URL.Path, "public/"),
			// Domain: r.URL.Host,
			MaxAge: 3600,
		}

		if ah, ok := r.Header["X-Authorization"]; ok {
			if s := strings.TrimPrefix(ah[0], "Bearer "); s != "" {
				cookie.Value = s
			}
		} else if s := r.PostFormValue("__authToken"); s != "" {
			cookie.Value = s
		}

		if cookie.Value != "" {
			http.SetCookie(w, cookie)
		}

		http.Redirect(w, r, "index.html", http.StatusFound)
		return
	}

	if eui64, err := NewEUI64(vars["eui64"]); err != nil {
		errors.NewNotFound(err).Println().Write(w)
	} else {
		served := false
		n.gateways.Range(func(k, v interface{}) bool {
			gateway := v.(*GatewayProxy)

			if d, ok := gateway.devices.Load(eui64); !ok {
				return true
			} else {
				deviceProxy := d.(*DeviceProxy)

				if deviceProxy.controller == nil {
					return false
				}

				http.StripPrefix(fmt.Sprintf("/api/v1/devices/%s/public/", eui64), deviceProxy).ServeHTTP(w, r)
				served = true
				return false
			}
		})

		if !served {
			errors.NewNotFound().Write(w)
		}
	}
}
