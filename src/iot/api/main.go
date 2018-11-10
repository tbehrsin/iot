package api

import (
	"crypto/tls"
	"fmt"
	"github.com/boltdb/bolt"
	"iot/apps"
	"iot/ble"
	"iot/upnp"
	"log"
	"net"
	"net/http"
	"sync"
)

const Server = "https://z3js.net"

type API struct {
	Registry    *apps.Registry
	DB          *bolt.DB
	Server      *http.Server
	Mapping     *upnp.Mapping
	serverMutex *sync.Mutex
	httpClient  *http.Client
}

func (api *API) Error(w http.ResponseWriter, err error, code int) {
	http.Error(w, err.Error(), code)
}

func NewAPI() (*API, error) {
	api := &API{}

	if db, err := bolt.Open("/data/z3js.db", 0600, nil); err != nil {
		log.Fatal(err)
	} else {
		api.DB = db

		db.Update(func(tx *bolt.Tx) error {
			if _, err := tx.CreateBucketIfNotExists([]byte("Auth")); err != nil {
				return fmt.Errorf("create bucket: %s", err)
			}

			return nil
		})
	}

	if r, err := apps.NewRegistry(); err != nil {
		log.Fatal(err)
	} else {
		api.Registry = r
	}

	http.HandleFunc("/api/v1/apps", api.HandleAppsCLRUD)
	http.Handle("/api/v1/", api.Registry)

	ble.HandleFunc("auth/GET_PIN_CODE_SEED", api.GetPINCodeSeed)
	ble.HandleFunc("auth/SET_PIN_CODE", api.SetPINCode)
	ble.HandleFunc("auth/VERIFY_PIN_CODE", api.VerifyPINCode)
	ble.HandleFunc("gateway/CREATE_GATEWAY", api.CreateGateway)

	if err := api.Start(); err != nil {
		log.Println(fmt.Errorf("cannot start server: %+v", err))
	}

	return api, nil
}

func (api *API) Start() error {
	if api.Server != nil {
		return fmt.Errorf("server already started")
	}

	if gw, err := api.Gateway(); err != nil {
		return err
	} else if gw == nil || len(gw.Certificates) == 0 || gw.PrivateKey == nil {
		return fmt.Errorf("gateway has no certificates")
	} else {
		api.Mapping = upnp.NewMapping(443, gw.Port, "z3js (https)", func(m *upnp.Mapping) {
			if m.ExternalIP != "" && m.PublicPort != 0 {
				log.Printf("upnp: mapped https to %s:%d\n", m.ExternalIP, m.PublicPort)

				gw.Port = m.PublicPort
				if err := gw.Update(); err != nil {
					log.Printf("upnp: could not save gateway: %+v\n", err)
				} else if err := api.UpdateGateway(); err != nil {
					log.Printf("upnp: could not update gateway: %+v\n", err)
				}
			} else {
				log.Printf("upnp: unmapped https\n")
			}
		})

		config := &tls.Config{
			MinVersion:               tls.VersionTLS12,
			CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
			PreferServerCipherSuites: true,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			},
			NextProtos:   []string{"http/1.1"},
			Certificates: []tls.Certificate{gw.MarshalTLSCertificate()},
		}
		api.Server = &http.Server{
			Addr:         ":https",
			Handler:      http.HandlerFunc(api.middleware),
			TLSConfig:    config,
			TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
		}
		go func() {
			api.serverMutex = &sync.Mutex{}
			api.serverMutex.Lock()
			l, err := net.Listen("tcp", api.Server.Addr)
			if err == nil {
				tlsListener := tls.NewListener(tcpKeepAliveListener{l.(*net.TCPListener)}, config)
				api.Server.Serve(tlsListener)
			}
			api.Server = nil
			api.serverMutex.Unlock()
		}()
	}
	return nil
}

func (api *API) Stop() error {
	if api.Server == nil {
		return fmt.Errorf("server not running")
	}

	if api.Mapping != nil {
		api.Mapping.Close()
	}

	err := api.Server.Close()
	api.serverMutex.Lock()
	api.serverMutex.Unlock()
	api.serverMutex = nil
	return err
}

func (api *API) middleware(w http.ResponseWriter, r *http.Request) {
	if r = api.VerifyTokenMiddleware(w, r); r == nil {
		return
	}

	http.DefaultServeMux.ServeHTTP(w, r)
}
