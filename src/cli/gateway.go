package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"github.com/grandcat/zeroconf"
	"golang.org/x/net/publicsuffix"
)

type Gateway struct {
	client   *http.Client
	hostName string
	Port     int
	ID       string
	Token    string
}

func FindGateway(id string) *Gateway {
	// Discover all services on the network (e.g. _workstation._tcp)
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		log.Fatalln("Failed to initialize resolver:", err.Error())
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	var gateway *zeroconf.ServiceEntry
	entries := make(chan *zeroconf.ServiceEntry)
	go func(results <-chan *zeroconf.ServiceEntry) {
		for entry := range results {
			if id != "" && entry.Instance == id {
				gateway = entry
				cancel()
				return
			} else if id == "" {
				gateway = entry
				cancel()
				return
			}
		}
	}(entries)

	err = resolver.Browse(ctx, "_iot-gateway._tcp", "local.", entries)
	if err != nil {
		log.Fatalln("Failed to browse:", err.Error())
	}

	<-ctx.Done()

	if gateway == nil {
		return nil
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			ServerName: fmt.Sprintf("%s.iot.behrsin.com", gateway.Instance),
		},
	}

	if jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List}); err != nil {
		panic(err)
	} else {
		client := &http.Client{Transport: tr, Jar: jar}

		return &Gateway{
			client:   client,
			hostName: gateway.HostName,
			Port:     gateway.Port,
			ID:       gateway.Instance,
		}
	}
}

type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e APIError) Error() string {
	return e.Message
}

func (g *Gateway) parseHTTPResponse(res *http.Response, response interface{}) error {
	var r map[string]interface{}

	defer res.Body.Close()
	if b, err := ioutil.ReadAll(res.Body); err != nil {
		return err
	} else {
		if len(b) == 0 {
			return nil
		}

		if err := json.Unmarshal(b, &r); err != nil {
			return err
		} else {
			if errorObject, ok := r["error"]; ok {
				b, _ := json.Marshal(errorObject)
				var apiError APIError
				if err := json.Unmarshal(b, &apiError); err != nil {
					return err
				} else {
					return apiError
				}
			} else if response != nil {
				if bodyObject, ok := r["body"]; ok {
					b, _ := json.Marshal(bodyObject)
					if err := json.Unmarshal(b, response); err != nil {
						return err
					} else {
						return nil

					}
				} else {
					return fmt.Errorf("failed to parse response")
				}
			}
		}
	}
	return nil
}

func (g *Gateway) httpRequest(method string, path string, request interface{}, response interface{}) error {
	if body, err := json.Marshal(request); err != nil {
		return err
	} else if req, err := http.NewRequest(method, fmt.Sprintf("https://%s:%d%s", g.hostName, g.Port, path), bytes.NewBuffer(body)); err != nil {
		return err
	} else {
		if g.Token != "" {
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", g.Token))
		}
		req.Header.Set("Content-Type", "application/json")
		if res, err := g.client.Do(req); err != nil {
			return err
		} else if err := g.parseHTTPResponse(res, response); err != nil {
			return err
		}
	}
	return nil
}

func (g *Gateway) Get(path string, response interface{}) error {
	return g.httpRequest(http.MethodGet, path, nil, response)
}

func (g *Gateway) Post(path string, request interface{}, response interface{}) error {
	return g.httpRequest(http.MethodPost, path, request, response)
}

func (g *Gateway) Put(path string, request interface{}, response interface{}) error {
	return g.httpRequest(http.MethodPut, path, request, response)
}

func (g *Gateway) WebSocket() (*websocket.Conn, error) {
	u := url.URL{Scheme: "wss", Host: g.hostName, Path: "/developer"}

	var dialer = &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
		TLSClientConfig: &tls.Config{
			ServerName: fmt.Sprintf("%s.iot.behrsin.com", g.ID),
		},
	}

	header := http.Header{}
	if g.Token != "" {
		header.Set("X-Authorization", fmt.Sprintf("Bearer %s", g.Token))
	}

	c, _, err := dialer.Dial(u.String(), header)
	return c, err
}
