package api

import (
	"crypto/tls"
	"fmt"
	"gateway/apps"
	"gateway/ble"
	"gateway/errors"
	iotnet "gateway/net"
	"gateway/upnp"
	"gateway/zigbee"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/boltdb/bolt"
	jwt "github.com/dgrijalva/jwt-go"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/hashicorp/yamux"
	proxy "github.com/koding/websocketproxy"
	"github.com/rs/xid"
)

const Server = "https://iot.behrsin.com"

type API struct {
	Network           *iotnet.Network
	Registry          *apps.Registry
	DB                *bolt.DB
	server            *http.Server
	session           net.Listener
	Mapping           *upnp.Mapping
	developerMode     *DeveloperMode
	developerRouter   http.Handler
	serverMutex       *sync.Mutex
	httpClient        *http.Client
	mqttClient        mqtt.Client
	inspectorMessages chan []byte
	running           bool
}

func NewAPI() (*API, error) {
	api := &API{}
	api.developerMode = &DeveloperMode{api: api}
	api.developerRouter = api.developerMode.CreateRouter()

	dbfile := "/data/z3js.db"

	if os.Getenv("DATABASE_FILE") != "" {
		dbfile = os.Getenv("DATABASE_FILE")
	}

	if db, err := bolt.Open(dbfile, 0600, nil); err != nil {
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

	broker := "tcp://localhost:1883"
	if os.Getenv("MQTT_URL") != "" {
		broker = os.Getenv("MQTT_URL")
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetAutoReconnect(true)
	opts.SetClientID(fmt.Sprintf("iot-gateway-%s", xid.New()))
	mqtt.CRITICAL = log.New(os.Stdout, "CRITICAL ", 0)
	mqtt.ERROR = log.New(os.Stdout, "ERROR ", 0)

	api.mqttClient = mqtt.NewClient(opts)

	api.mqttConnect()

	network := iotnet.New(api.mqttClient)
	network.AddGateway(zigbee.NewGateway(api.mqttClient))
	api.Network = network

	if r, err := apps.NewRegistry(api.Network); err != nil {
		log.Fatal(err)
	} else {
		api.Registry = r
	}

	router := mux.NewRouter()
	router.Handle("/api/v1/developer/", api.developerMode).Methods(http.MethodGet, http.MethodPost)
	router.HandleFunc("/api/v1/auth/", api.CreateEmailTokenHandler).Methods(http.MethodPost)
	router.HandleFunc("/api/v1/auth/code", api.CreateAuthCode).Methods(http.MethodPost)
	router.PathPrefix("/api/v1/devices/").Handler(network.CreateRouter())
	router.HandleFunc("/api/v1/apps", api.HandleAppsCLRUD)
	// router.PathPrefix("/api/v1/").Handler(api.Registry)
	http.Handle("/api/v1/", router)

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
	if api.running {
		return fmt.Errorf("server already started")
	}
	api.running = true

	if gw, err := api.Gateway(); err != nil {
		api.running = false
		return err
	} else if gw == nil || len(gw.Certificates) == 0 || gw.PrivateKey == nil {
		api.running = false
		return fmt.Errorf("gateway has no certificates")
	} else {
		// api.Mapping = upnp.NewMapping(443, gw.Port, "z3js (https)", func(m *upnp.Mapping) {
		// 	if m.ExternalIP != "" && m.PublicPort != 0 {
		// 		log.Printf("upnp: mapped https://%s:%d to %s:%d\n", gw.FQDN, m.PublicPort, m.ExternalIP, m.PublicPort)
		//
		// 		gw.Port = m.PublicPort
		// 		if err := gw.Update(); err != nil {
		// 			log.Printf("upnp: could not save gateway: %+v\n", err)
		// 		} else if err := api.UpdateGateway(); err != nil {
		// 			log.Printf("upnp: could not update gateway: %+v\n", err)
		// 		}
		// 	} else {
		// 		log.Printf("upnp: unmapped https\n")
		// 	}
		// })

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

		go func() {
			api.server = &http.Server{
				Addr:         ":https",
				Handler:      http.HandlerFunc(api.middleware),
				TLSConfig:    config,
				TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
			}

			l, err := net.Listen("tcp", api.server.Addr)
			if err == nil {
				tlsListener := tls.NewListener(tcpKeepAliveListener{l.(*net.TCPListener)}, config)
				api.server.Serve(tlsListener)
			} else {
				api.server = nil
			}
		}()

		go func() {
			clientConfig := &tls.Config{
				Certificates: []tls.Certificate{gw.MarshalTLSCertificate()},
			}
			clientConfig.BuildNameToCertificate()

			if conn, err := tls.Dial("tcp", "proxy.iot.behrsin.com:443", clientConfig); err != nil {
				if api.running {
					api.running = false
					if api.server != nil {
						api.server.Close()
					}
					time.Sleep(1 * time.Second)
					api.Start()
				}
			} else {
				address := conn.LocalAddr().(*net.TCPAddr).IP
				if err := api.SetLocalAddress(address.String()); err != nil {
					log.Println(err)
				}

				if session, err := yamux.Server(conn, nil); err != nil {
					conn.Close()
					if api.running {
						api.running = false
						if api.server != nil {
							api.server.Close()
						}
						time.Sleep(1 * time.Second)
						api.Start()
					}
				} else {
					//log.Println(api.Server.Serve(&TLSListener{session, config}))
					//tlsListener := tls.NewListener(session, config)
					for {
						if stream, err := session.Accept(); err != nil {
							break
						} else {
							go func(stream net.Conn) {
								if client, err := net.Dial("tcp", "localhost:443"); err != nil {
									log.Println(err)
								} else {
									defer stream.Close()
									defer client.Close()
									pipe(stream, client)
								}
							}(stream)
						}
					}

					if api.running {
						api.running = false
						if api.server != nil {
							api.server.Close()
						}
						time.Sleep(1 * time.Second)
						api.Start()
					}
				}
			}
		}()

		api.developerMode.Start()
	}
	return nil
}

func (api *API) Stop() (err error) {
	if !api.running {
		return fmt.Errorf("server not running")
	}
	api.running = false

	if api.Mapping != nil {
		api.Mapping.Close()
	}

	if api.server != nil {
		err = api.server.Close()
		api.server = nil
	}

	if api.DB != nil {
		api.DB.Close()
	}

	return
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func (api *API) middleware(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Origin") != "" && r.Header.Get("Origin") != "null" {
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	} else {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	}
	w.Header().Set("Vary", "Origin")

	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, PATCH, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-Authorization, Origin")
		w.WriteHeader(200)
		return
	}

	if r = api.VerifyTokenMiddleware(w, r); r == nil {
		return
	}

	claims := r.Context().Value(JWTClaimsContextKey).(jwt.MapClaims)
	if claims["aud"] == "app" {
		if websocket.IsWebSocketUpgrade(r) {
			wsurl := "ws://127.0.0.1:9001/"
			if os.Getenv("MQTT_WS_URL") != "" {
				wsurl = os.Getenv("MQTT_WS_URL")
			}
			if u, err := url.Parse(wsurl); err != nil {
				log.Fatal(err)
			} else {
				backend := func(r *http.Request) *url.URL {
					t := *u
					return &t
				}
				p := &proxy.WebsocketProxy{Backend: backend}
				p.Upgrader = &upgrader
				p.ServeHTTP(w, r)
			}
			return
		}

		http.DefaultServeMux.ServeHTTP(w, r)
	} else if claims["aud"] == "developer" {
		api.developerRouter.ServeHTTP(w, r)
	} else if _, ok := tokenConversions[claims["aud"].(string)]; ok {
		router := mux.NewRouter()
		router.HandleFunc("/api/v1/auth/", api.ConvertToken).Methods(http.MethodPut)
		router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			errors.NewUnauthorized().Write(w)
		})
		router.ServeHTTP(w, r)
	} else {
		errors.NewUnauthorized().Write(w)
	}
}

func (a *API) mqttConnect() {
	retry := time.NewTicker(5 * time.Second)
RetryLoop:
	for {
		select {
		case <-retry.C:
			if token := a.mqttClient.Connect(); token.Wait() && token.Error() != nil {

			} else {
				retry.Stop()
				break RetryLoop
			}
		}
	}
}
