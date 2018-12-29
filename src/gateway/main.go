package main

import (
	"gateway/api"
	"gateway/ble"
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
	defer api.Stop()
	// go api.StartInspector()

	// go func() {
	// 	if _, err := api.Registry.LoadFromName("test"); err != nil {
	// 		log.Println(err)
	// 	}
	// }()

	go func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		<-signals
		api.Stop()
		os.Exit(0)
	}()

	ble.Start()
}