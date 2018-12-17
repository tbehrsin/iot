package ble

import (
	"encoding/json"
	"io/ioutil"
	"iot/errors"
	"log"
	"net/http"
	"os"

	"github.com/paypal/gatt"
)

func Start() {
	if os.Getenv("BLUETOOTH_EMULATION") == "" {
		var s *JSONService
		log.SetOutput(ioutil.Discard)
		d, err := gatt.NewDevice()
		log.SetOutput(os.Stderr)
		if err != nil {
			log.Fatalf("failed to open device, err: %s", err)
		}

		d.Handle(
			gatt.CentralConnected(func(c gatt.Central) {
				log.Println("connect: ", c.ID(), c.MTU())
				s.Centrals[c] = &JSONConnection{
					Transactions:  make([]JSONTransaction, 256),
					NotifyChannel: make(chan struct{}),
				}
			}),
			gatt.CentralDisconnected(func(c gatt.Central) {
				log.Println("disconnect: ", c.ID())
			}),
		)

		onStateChanged := func(d gatt.Device, state gatt.State) {
			switch state {
			case gatt.StatePoweredOn:
				log.SetOutput(ioutil.Discard)
				d.AddService(NewGapService("z3js"))
				d.AddService(NewGattService())
				log.SetOutput(os.Stderr)

				s = NewJSONService()
				log.SetOutput(ioutil.Discard)
				d.AddService(s.Service)
				log.SetOutput(os.Stderr)

				d.AdvertiseNameAndServices("z3js", []gatt.UUID{s.Service.UUID()})
			default:
			}
		}

		d.Init(onStateChanged)

	} else {
		http.ListenAndServe(":http", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := ioutil.ReadAll(r.Body)

			var request JSONRequest
			if err := json.Unmarshal(body, &request); err != nil {
				log.Println(err)
				return
			}

			if handler, ok := JSONHandlers[request.Type]; !ok {
				errors.NewInternalServerError("unknown message type \"%s\"", request.Type).Println().Write(w)
				return
			} else if response, err := handler(request.Payload); err != nil {
				errors.NewInternalServerError(err).Println().Write(w)
				return
			} else {
				out := make(map[string]interface{})
				out["body"] = response
				// marshal response to json
				if d, err := json.Marshal(out); err != nil {
					errors.NewInternalServerError("error encountered serializing response to json").Println().Write(w)
				} else {
					w.Write(d)
				}
			}
		}))
	}

	select {}
}
