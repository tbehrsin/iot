package main

import (
	"github.com/gorilla/websocket"
	proxy "github.com/koding/websocketproxy"
	"iot/api"
	"iot/ble"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func main() {
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("dist/assets/"))))

	http.HandleFunc("/renderer.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "dist/renderer.js")
	})

	if u, err := url.Parse("ws://localhost:9001/"); err != nil {
		log.Fatal(err)
	} else {
		p := proxy.NewProxy(u)
		p.Upgrader = &upgrader
		http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
			if websocket.IsWebSocketUpgrade(r) {
				p.ServeHTTP(w, r)
			} else {
				http.ServeFile(w, r, "dist/index.html")
			}
		})
	}

	api, _ := api.NewAPI()
	defer api.DB.Close()
	defer api.Mapping.Close()
	defer api.Stop()

	go api.Registry.Load("test")

	go func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		<-signals
		api.Mapping.Close()
		api.DB.Close()
		api.Stop()
		os.Exit(0)
	}()

	ble.Start()
}
