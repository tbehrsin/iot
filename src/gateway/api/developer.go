package api

import (
	"api"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"gateway/errors"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/grandcat/zeroconf"
)

type DeveloperMode struct {
	api        *API
	Enabled    bool
	httpServer *http.Server
	mdnsServer *zeroconf.Server
}

func (d *DeveloperMode) Start() (err error) {
	if d.Enabled {
		return nil
	}
	d.Enabled = true

	mutex := &sync.Mutex{}

	if gw, err := d.api.Gateway(); err != nil {
		return err
	} else if gw == nil || len(gw.Certificates) == 0 || gw.PrivateKey == nil {
		return fmt.Errorf("gateway has no certificates")
	} else {
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
		d.httpServer = &http.Server{
			Addr:         ":0",
			Handler:      d.createRouter(),
			TLSConfig:    config,
			TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
		}

		var listener net.Listener
		mutex.Lock()
		go func() {
			l, err := net.Listen("tcp", d.httpServer.Addr)
			if err == nil {
				listener = l
				mutex.Unlock()

				tlsListener := tls.NewListener(tcpKeepAliveListener{l.(*net.TCPListener)}, config)
				d.httpServer.Serve(tlsListener)
			} else {
				log.Println(err)
			}
			d.Stop()
		}()

		go func() {
			mutex.Lock()
			mutex.Unlock()

			if _, p, err := net.SplitHostPort(listener.Addr().String()); err != nil {
				log.Println(err)
				d.Stop()
				return
			} else if port, err := strconv.Atoi(p); err != nil {
				log.Println(err)
				d.Stop()
				return
			} else if d.mdnsServer, err = zeroconf.Register(gw.ID, "_iot-gateway._tcp", "local.", port, nil, nil); err != nil {
				d.Stop()
				return
			}
		}()
	}

	return nil
}

func (d *DeveloperMode) Stop() (err error) {
	if !d.Enabled {
		return
	}
	d.Enabled = false

	d.mdnsServer.Shutdown()
	d.httpServer.Close()

	return nil
}

type DeveloperModeMessage struct {
	Enabled bool `json:"enabled"`
}

func (d *DeveloperMode) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		response := DeveloperModeMessage{d.Enabled}
		errors.APIJSON(w, response)
	} else if r.Method == http.MethodPost {
		var request DeveloperModeMessage

		if b, err := ioutil.ReadAll(r.Body); err != nil {
			errors.NewBadRequest(err).Println().Write(w)
			return
		} else {
			log.Println(string(b))
			if err := json.Unmarshal(b, &request); err != nil {
				errors.NewBadRequest(err).Println().Write(w)
				return
			}
		}

		if request.Enabled {
			d.Start()
		} else {
			d.Stop()
		}

		errors.APIJSON(w, request)
	}
}

func (d *DeveloperMode) createRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/api/v1/auth/", d.api.CreateCLITokenHandler).Methods(http.MethodPost)
	return r
}

func (d *DeveloperMode) CreateRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/developer", d.HandleWebSocket)
	return r
}

func (d *DeveloperMode) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("developer upgrade:", err)
		return
	}

	var app api.Application

	p := &api.ServerProtocol{}

	write := make(chan []byte)
	read := make(chan []byte)

	defer func() {
		if app != nil {
			defer log.Println("DeveloperMode: terminating app")
			app.Terminate()
		}
	}()
	defer c.Close()
	defer close(write)
	defer close(read)

	defer log.Println("DeveloperMode: closing reader")
	go func() {
		for {
			message, more := <-write

			if more {
				c.WriteMessage(websocket.BinaryMessage, message)
			} else {
				break
			}
		}
		defer log.Println("DeveloperMode: closing writer")
	}()

	if err := p.Run(d, read, write); err != nil {
		log.Println("error running server protocol:", err)
	}

	go func() {
		if a, err := d.api.Registry.Load(p); err != nil {
			log.Println("error loading client application:", err)
		} else {
			app = a
		}
	}()

	for {
		if mt, message, err := c.ReadMessage(); err != nil {
			return
		} else if mt == websocket.BinaryMessage {
			read <- message
		} else if mt == websocket.CloseMessage {
			return
		}
	}

}

func (d *DeveloperMode) ProtocolError(err error) {
	log.Println("developer protocol error:", err)
}
