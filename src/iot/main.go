package main

import (
	"github.com/gorilla/websocket"
	proxy "github.com/koding/websocketproxy"
	"iot/api"
	"iot/plugin"
	"log"
	"net/http"
	"net/url"
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

	for i := 0; i < 4; i++ {
		if plugin, err := plugin.NewPlugin("src/iot/plugin/hello-world.js"); err != nil {
			log.Fatal(err)
		} else {
			defer plugin.Terminate()
		}
	}

	if plugin, err := plugin.NewPlugin("src/iot/plugin/api.js"); err != nil {
		log.Fatal(err)
	} else {
		defer plugin.Terminate()
	}

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

	if _, err := api.NewAPI(); err != nil {
		log.Fatal(err)
	}

	log.Fatal(http.ListenAndServe(":80", nil))
}
