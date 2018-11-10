package main

import (
	"context"
	"crypto/tls"
	"db"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"log"
	"net/http"
	"strings"
)

type gatewayContextType int

const GatewayContextKey gatewayContextType = 0

func GetGatewayFromToken(token string) (*db.Gateway, error) {
	// extract the gateway id from the claims
	tokenParts := strings.Split(token, ".")
	var claims map[string]interface{}

	if len(tokenParts) != 3 {
		return nil, fmt.Errorf("invalid bearer token: expected jwt")
	} else if claimsString, err := base64.StdEncoding.DecodeString(tokenParts[1]); err != nil {
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
	if !strings.HasPrefix(strings.ToLower(r.Header["Authorization"][0]), "bearer ") {
		APIErrorWithStatus(w, fmt.Errorf("must provide bearer token"), http.StatusUnauthorized)
		return
	}

	tokenString := strings.Trim(strings.SplitN(r.Header["Authorization"][0], " ", 2)[1], " ")

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
		APIErrorWithStatus(w, fmt.Errorf("404 Not Found\nUPnP Port Forwarding Disabled"), http.StatusNotFound)
		return
	} else {
		url := *r.URL
		url.Scheme = "https"
		url.User = nil
		url.Host = fmt.Sprintf("%s.z3js.net:%d", gw.ID, gw.Port)

		w.Header()["Location"] = []string{url.String()}
		w.WriteHeader(http.StatusTemporaryRedirect)
		w.Write([]byte{})
	}
})

var serverHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains")

	if len(r.TLS.PeerCertificates) >= 1 {
		id := strings.TrimSuffix(r.TLS.PeerCertificates[0].Subject.CommonName, ".z3js.net")
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

	if _, ok := r.Header["Authorization"]; !ok {
		publicHandler.ServeHTTP(w, r)
		return
	}

	redirectHandler.ServeHTTP(w, r)
})

var insecureHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains")

	url := *r.URL
	url.Scheme = "https"
	url.User = nil

	w.Header()["Location"] = []string{url.String()}
	w.WriteHeader(http.StatusPermanentRedirect)
	w.Write([]byte{})
})

func main() {
	ca := flag.Bool("ca", false, "create new certificate authority")

	flag.Parse()

	if *ca {
		if err := CreateCA(); err != nil {
			log.Println(err)
		}
	} else {
		go func() {
			log.Fatal(http.ListenAndServe(":http", insecureHandler))
		}()

		if ca, err := GetCA(); err != nil {
			log.Fatal(err)
		} else if cert, err := ca.CreateServerCertificate("z3js.net"); err != nil {
			log.Fatal(err)
		} else if pool, err := ca.CertPool(); err != nil {
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

			s := &http.Server{
				Addr:      ":https",
				TLSConfig: config,
				Handler:   serverHandler,
			}

			log.Fatal(s.ListenAndServeTLS("", ""))
		}
	}
}
