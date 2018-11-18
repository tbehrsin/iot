package main

import (
	"iot/api"
	"iot/ble"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("dist/assets/"))))

	http.HandleFunc("/renderer.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "dist/renderer.js")
	})

	api, _ := api.NewAPI()
	defer api.DB.Close()
	defer api.Mapping.Close()
	defer api.Stop()

	go func() {
		if _, err := api.Registry.Load("test"); err != nil {
			log.Println(err)
		}
	}()

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
