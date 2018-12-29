package main

import (
	"context"
	"crypto/tls"
	"db"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/tcpproxy"
	"github.com/hashicorp/yamux"
)

type gatewayContextType int

const GatewayContextKey gatewayContextType = 0

func GetGatewayFromToken(token string) (*db.Gateway, error) {
	// extract the gateway id from the claims
	tokenParts := strings.Split(token, ".")
	var claims map[string]interface{}

	if len(tokenParts) != 3 {
		return nil, fmt.Errorf("invalid bearer token: expected jwt")
	} else if claimsString, err := base64.RawURLEncoding.DecodeString(tokenParts[1]); err != nil {
		return nil, fmt.Errorf("invalid bearer token: invalid jwt, %+v", err)
	} else if err := json.Unmarshal([]byte(claimsString), &claims); err != nil {
		return nil, err
	} else if id, ok := claims["gateway"]; !ok {
		return nil, fmt.Errorf("claims does not contain gateway")
	} else if idString, ok := id.(string); !ok {
		return nil, fmt.Errorf("claims does not contain gateway id")
	} else if gw, err := db.GetGateway(idString); err != nil {
		return nil, err
	} else if gw == nil {
		return nil, fmt.Errorf("unable to find gateway matching claims")
	} else {
		return gw, nil
	}
}

var publicHandler = api()
var privateHandler = private()

var redirectHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	var ah []string
	if header, ok := r.Header["Authorization"]; ok {
		ah = header
	} else if header, ok := r.Header["X-Authorization"]; ok {
		ah = header
	} else if _, ok := r.Header["Sec-Websocket-Protocol"]; ok {
		if r.Header["Sec-Websocket-Protocol"][0] == "mqtt" {
			ah = []string{fmt.Sprintf("Bearer %s", strings.TrimPrefix(r.URL.Path, "/"))}
		}
	}

	if ah == nil {
		APIErrorWithStatus(w, fmt.Errorf("must provide bearer token"), http.StatusUnauthorized)
		return
	}

	if !strings.HasPrefix(strings.ToLower(ah[0]), "bearer ") {
		APIErrorWithStatus(w, fmt.Errorf("must provide bearer token"), http.StatusUnauthorized)
		return
	}

	tokenString := strings.Trim(strings.SplitN(ah[0], " ", 2)[1], " ")

	if gw, err := GetGatewayFromToken(tokenString); err != nil {
		APIErrorWithStatus(w, fmt.Errorf("invalid bearer token: %+v", err), http.StatusUnauthorized)
		return
	} else if publicKey, err := gw.UnmarshalPublicKey(); err != nil {
		APIErrorWithStatus(w, fmt.Errorf("invalid public key for gateway: %+v", err), http.StatusUnauthorized)
		return
	} else if publicKey == nil {
		APIErrorWithStatus(w, fmt.Errorf("no public key for gateway"), http.StatusUnauthorized)
		return
	} else if token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return publicKey, nil
	}); err != nil {
		APIErrorWithStatus(w, fmt.Errorf("invalid bearer token: %+v", err), http.StatusUnauthorized)
		return
	} else if _, ok := token.Claims.(jwt.MapClaims); !ok || !token.Valid {
		APIErrorWithStatus(w, fmt.Errorf("invalid bearer token"), http.StatusUnauthorized)
		return
	} else if gw.Port == 0 {
		APIErrorWithStatus(w, fmt.Errorf("404 Not Found\nPort Forwarding Disabled"), http.StatusNotFound)
		return
	} else {
		url := *r.URL
		url.Scheme = "https"
		url.User = nil
		url.Host = fmt.Sprintf("%s.%s:%d", gw.ID, strings.TrimSuffix(db.Domain, "."), gw.Port)

		w.Header()["Location"] = []string{url.String()}

		if r.Method == http.MethodOptions {
			w.Header()["Access-Control-Allow-Methods"] = []string{"HEAD, POST, GET, OPTIONS, PUT, PATCH, DELETE"}
			w.Header()["Access-Control-Allow-Headers"] = []string{"Accept, Content-Type, Content-Length, Accept-Encoding, X-Authorization"}
			w.WriteHeader(http.StatusOK)
			return
		}

		w.WriteHeader(http.StatusTemporaryRedirect)

		w.Write([]byte{})
	}
})

var serverHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Strict-Transport-Security", "max-age=63072000")

	w.Header().Set("Access-Control-Expose-Headers", "Location")
	// if r.Header.Get("Origin") != "" {
	// 	w.Header()["Access-Control-Allow-Origin"] = []string{r.Header.Get("Origin")}
	// } else {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	// }
	w.Header().Set("Vary", "Origin")

	if r.Method == http.MethodOptions && r.Header.Get("X-Authorization") == "" {
		w.Header()["Access-Control-Allow-Methods"] = []string{"HEAD, POST, GET, OPTIONS, PUT, PATCH, DELETE"}
		w.Header()["Access-Control-Allow-Headers"] = []string{"Accept, Content-Type, Content-Length, Accept-Encoding, X-Authorization"}
		w.WriteHeader(http.StatusOK)
		return
	}

	if len(r.TLS.PeerCertificates) >= 1 {
		id := strings.TrimSuffix(r.TLS.PeerCertificates[0].Subject.CommonName, fmt.Sprintf(".%s", strings.TrimSuffix(db.Domain, ".")))
		log.Println(id)
		if gw, err := db.GetGateway(id); err != nil {
			APIErrorWithStatus(w, fmt.Errorf("invalid peer certificate for gateway: %+v", err), http.StatusUnauthorized)
			return
		} else {
			ctx := context.WithValue(r.Context(), GatewayContextKey, gw)
			privateHandler.ServeHTTP(w, r.WithContext(ctx))
			return
		}
	}

	if _, ok := r.Header["Authorization"]; ok {
		redirectHandler.ServeHTTP(w, r)
		return
	} else if _, ok := r.Header["X-Authorization"]; ok {
		redirectHandler.ServeHTTP(w, r)
		return
	} else if _, ok := r.Header["Sec-Websocket-Protocol"]; ok {
		if r.Header["Sec-Websocket-Protocol"][0] == "mqtt" {
			redirectHandler.ServeHTTP(w, r)
			return
		}
	}

	publicHandler.ServeHTTP(w, r)
})

var insecureHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	if r.Host == fmt.Sprintf("ca.%s", strings.TrimSuffix(db.Domain, ".")) || r.Host == fmt.Sprintf("ca.%s:80", strings.TrimSuffix(db.Domain, ".")) {
		ServeCACertificate(w, r)
		return
	}

	w.Header().Add("Strict-Transport-Security", "max-age=63072000")

	url := *r.URL
	url.Scheme = "https"
	url.Host = strings.TrimSuffix(db.Domain, ".")
	url.User = nil

	w.Header()["Location"] = []string{url.String()}
	w.WriteHeader(http.StatusPermanentRedirect)
	w.Write([]byte{})
})

var externalIP string

func main() {
	if err := db.Initialize(); err != nil {
		panic(err)
	}

	ca := flag.Bool("ca", false, "create new certificate authority")
	local := flag.Bool("local", false, "use localhost as external IP")
	gce := flag.Bool("gce", false, "use Google Compute Engine metadata to obtain external IP")

	flag.Parse()

	if *ca {
		if err := db.CreateCA(); err != nil {
			log.Println(err)
		}
	} else {

		if *local {
			externalIP = "127.0.0.1"
		} else if *gce {
			if eip, err := GetGCEExternalIP(); err != nil {
				panic(err)
			} else {
				externalIP = eip
			}
		} else {
			log.Println("Cannot find external IP")
			flag.Usage()
			os.Exit(1)
		}

		go func() {
			insecurePort := "http"
			if os.Getenv("INSECURE_PORT") != "" {
				insecurePort = os.Getenv("INSECURE_PORT")
			}

			log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", insecurePort), insecureHandler))
		}()

		if ca, err := db.GetCA(); err != nil {
			log.Fatal(err)
		} else if cert, err := ca.CreateServerCertificate(strings.TrimSuffix(db.Domain, ".")); err != nil {
			log.Fatal(err)
		} else if pool, err := ca.CertPool(); err != nil {
			log.Fatal(err)
		} else if gatewayListener, err := createGatewayListener(ca); err != nil {
			log.Fatal(err)
		} else {
			config := &tls.Config{
				MinVersion:       tls.VersionTLS12,
				CurvePreferences: []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},

				NextProtos:   []string{"http/1.1"},
				ClientCAs:    pool,
				ClientAuth:   tls.VerifyClientCertIfGiven,
				Certificates: []tls.Certificate{*cert},
			}

			port := "https"
			if os.Getenv("PORT") != "" {
				port = os.Getenv("PORT")
			}
			port = fmt.Sprintf(":%s", port)
			domain := strings.TrimSuffix(db.Domain, ".")

			var p tcpproxy.Proxy
			listener := &tcpproxy.TargetListener{}
			proxyListener := &tcpproxy.TargetListener{}
			p.AddSNIRoute(port, domain, listener)
			p.AddSNIRoute(port, fmt.Sprintf("proxy.%s", domain), gatewayListener)
			suffix := fmt.Sprintf(".%s", domain)
			p.AddSNIMatchRoute(port, func(ctx context.Context, hostname string) bool {
				return strings.HasSuffix(hostname, suffix) && !strings.ContainsAny(strings.TrimSuffix(hostname, suffix), ".")
			}, proxyListener)
			p.Start()

			go func() {
				log.Fatal(ServeProxy(proxyListener))
			}()

			s := &http.Server{
				Addr:      fmt.Sprintf(":%s", port),
				TLSConfig: config,
				Handler:   serverHandler,
			}

			log.Fatal(s.ServeTLS(listener, "", ""))
		}
	}
}

func GetGCEExternalIP() (string, error) {
	client := &http.Client{}

	if req, err := http.NewRequest(http.MethodGet, "http://metadata/computeMetadata/v1/instance/network-interfaces/0/access-configs/0/external-ip", nil); err != nil {
		return "", err
	} else {
		req.Header.Set("Metadata-Flavor", "Google")

		if res, err := client.Do(req); err != nil {
			return "", err
		} else {
			defer res.Body.Close()
			if b, err := ioutil.ReadAll(res.Body); err != nil {
				return "", err
			} else {
				return string(b), nil
			}
		}
	}
}

var sessions = map[string]*yamux.Session{}

func createGatewayListener(ca *db.CA) (*tcpproxy.TargetListener, error) {
	if cert, err := ca.CreateServerCertificate(fmt.Sprintf("proxy.%s", strings.TrimSuffix(db.Domain, "."))); err != nil {
		return nil, err
	} else if pool, err := ca.CertPool(); err != nil {
		return nil, err
	} else {
		config := &tls.Config{
			MinVersion:       tls.VersionTLS12,
			CurvePreferences: []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},

			NextProtos:   []string{"http/1.1"},
			ClientCAs:    pool,
			ClientAuth:   tls.RequireAndVerifyClientCert,
			Certificates: []tls.Certificate{*cert},
		}

		listener := &tcpproxy.TargetListener{}

		go func() {
			log.Fatal(ServeGateway(listener, config))
		}()

		return listener, nil
	}
}

func ServeProxy(listener net.Listener) error {
	suffix := fmt.Sprintf(".%s", strings.TrimSuffix(db.Domain, "."))

	for {
		conn, _ := listener.Accept()

		t := tcpproxy.UnderlyingConn(conn).(*net.TCPConn)
		t.SetKeepAlive(true)
		t.SetKeepAlivePeriod(time.Duration(3) * time.Minute)

		go func(conn *tcpproxy.Conn) {
			id := strings.TrimSuffix(conn.HostName, suffix)

			if session, ok := sessions[id]; !ok {
				log.Println("cannot find gateway", id)
				ServeRemoteProxy(id, conn)
			} else if stream, err := session.Open(); err != nil {
				log.Println("cannot open session", id)
				delete(sessions, id)
				ServeRemoteProxy(id, conn)
			} else {
				defer conn.Close()
				defer stream.Close()
				log.Println("piping stream")
				pipe(stream, conn)
			}
		}(conn.(*tcpproxy.Conn))
	}
}

func ServeRemoteProxy(id string, conn net.Conn) {
	if gw, err := db.GetGateway(id); err != nil {
		log.Println("cannot find gateway", id, "closing connection")
		conn.Close()
		return
	} else if gw.Address != externalIP {
		tcpproxy.To(fmt.Sprintf("%s:%d", gw.Address, gw.Port)).HandleConn(conn)
	} else {
		log.Println("loopback detected, gateway configured to point to this server but no gateway connection", id)
	}
}

func ServeGateway(listener net.Listener, config *tls.Config) error {
	suffix := fmt.Sprintf(".%s", strings.TrimSuffix(db.Domain, "."))
	for {
		c, _ := listener.Accept()

		go func(conn net.Conn) {
			tlsConn := tls.Server(conn, config)
			tlsConn.Read([]byte{})
			if session, err := yamux.Client(tlsConn, nil); err != nil {
				log.Println("cannot create client session", err)
				tlsConn.Close()
			} else if state := tlsConn.ConnectionState(); len(state.PeerCertificates) >= 1 {
				id := strings.TrimSuffix(state.PeerCertificates[0].Subject.CommonName, suffix)
				if gw, err := db.GetGateway(id); err != nil {
					log.Println("closing connection", gw)
					conn.Close()
					return
				} else {
					if s, ok := sessions[id]; ok {
						log.Println("closing pre-exiting session for gateway", id)
						s.Close()
						delete(sessions, id)
					}
					gw.Address = externalIP
					gw.Port = 443
					gw.IsOrigin = true
					if err := gw.Update(); err != nil {
						log.Println("unable to update gateway", err)
						conn.Close()
					} else {
						log.Printf("client session active and available on %s:443", externalIP)
						sessions[id] = session
					}
				}
			} else {
				log.Println("closing connection as no valid peer certificates")
				conn.Close()
			}
		}(c)
	}
}

// https://www.stavros.io/posts/proxying-two-connections-go/
func chanFromConn(conn net.Conn) chan []byte {
	c := make(chan []byte)

	go func() {
		b := make([]byte, 1024)

		for {
			n, err := conn.Read(b)
			if n > 0 {
				res := make([]byte, n)
				// Copy the buffer so it doesn't get changed while read by the recipient.
				copy(res, b[:n])
				c <- res
			}
			if err != nil {
				c <- nil
				break
			}
		}
	}()

	return c
}

// https://www.stavros.io/posts/proxying-two-connections-go/
func pipe(conn1 net.Conn, conn2 net.Conn) {
	chan1 := chanFromConn(conn1)
	chan2 := chanFromConn(conn2)
	defer log.Println("closing pipe")
	for {
		select {
		case b1 := <-chan1:
			if b1 == nil {
				return
			} else {
				conn2.Write(b1)
			}
		case b2 := <-chan2:
			if b2 == nil {
				return
			} else {
				conn1.Write(b2)
			}
		}
	}
}
