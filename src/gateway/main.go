package main

import (
	"flag"
	"gateway/api"
	"gateway/ble"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	netpprof "net/http/pprof"

	v8 "github.com/behrsin/go-v8"
	"github.com/gorilla/mux"
)

var profile = flag.Bool("profile", false, "serve profile information on port :6060")

func main() {
	flag.Parse()

	// if *cpuprofile != "" {
	// 	if f, err := os.Create(*cpuprofile); err != nil {
	// 		log.Fatal("could not create CPU profile: ", err)
	// 	} else if err := pprof.StartCPUProfile(f); err != nil {
	// 		log.Fatal("could not start CPU profile: ", err)
	// 	}
	// 	defer pprof.StopCPUProfile()
	// }

	// http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("dist/assets/"))))
	//
	// http.HandleFunc("/renderer.js", func(w http.ResponseWriter, r *http.Request) {
	// 	http.ServeFile(w, r, "dist/renderer.js")
	// })

	if *profile {
		go func() {
			r := mux.NewRouter()
			r.PathPrefix("/debug/pprof/").HandlerFunc(netpprof.Index)
			r.Handle("/debug/v8/", v8.TracerHandler())
			log.Println(http.ListenAndServe("localhost:6060", r))
		}()

		v8.StartTracer(v8.SimpleTracer)
		v8.EnableAllocationStackTraces()

		// go func() {
		// 	for {
		// 		v8.DumpTracer(os.Stdout, false)
		//
		// 		// var m runtime.MemStats
		// 		// runtime.ReadMemStats(&m)
		// 		// // For info on each, see: https://golang.org/pkg/runtime/#MemStats
		// 		// fmt.Printf("Alloc = %v", m.Alloc)
		// 		// fmt.Printf("\tTotalAlloc = %v", m.TotalAlloc)
		// 		// fmt.Printf("\tSys = %v", m.Sys)
		// 		// fmt.Printf("\tNumGC = %v\n", m.NumGC)
		//
		// 		time.Sleep(1 * time.Second)
		// 	}
		// }()
	}

	api, _ := api.NewAPI()
	defer api.Stop()
	// go api.StartInspector()

	// go func() {
	// 	if _, err := api.Registry.LoadFromName("test"); err != nil {
	// 		log.Println(err)
	// 	}
	// }()

	go func() {
		ble.Start()
	}()

	go func() {
		for {
			signals := make(chan os.Signal, 1)
			signal.Notify(signals, syscall.SIGHUP)
			<-signals
			v8.DumpTracer(os.Stdout, true)
		}
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-signals
	api.Stop()

	// if *memprofile != "" {
	// 	var f *os.File
	// 	var err error
	// 	if f, err = os.Create(*memprofile); err != nil {
	// 		log.Fatal("could not create memory profile: ", err)
	// 	}
	// 	runtime.GC()
	// 	if err = pprof.WriteHeapProfile(f); err != nil {
	// 		log.Fatal("could not write memory profile: ", err)
	// 	}
	// 	f.Close()
	// }
}
