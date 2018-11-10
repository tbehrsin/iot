package ble

import (
	"fmt"
	"log"

	"github.com/paypal/gatt"
)

func Start() {
	var s *JSONService
	d, err := gatt.NewDevice()
	if err != nil {
		log.Fatalf("Failed to open device, err: %s", err)
	}

	// Register optional handlers.
	d.Handle(
		gatt.CentralConnected(func(c gatt.Central) {
			fmt.Println("Connect: ", c.ID(), c.MTU())
			s.Centrals[c] = &JSONConnection{
				Transactions:  make([]JSONTransaction, 256),
				NotifyChannel: make(chan struct{}),
			}
		}),
		gatt.CentralDisconnected(func(c gatt.Central) {
			fmt.Println("Disconnect: ", c.ID())
		}),
	)

	// A mandatory handler for monitoring device state.
	onStateChanged := func(d gatt.Device, state gatt.State) {
		fmt.Printf("State: %s\n", s)
		switch state {
		case gatt.StatePoweredOn:
			d.AddService(NewGapService("z3js"))
			d.AddService(NewGattService())

			s = NewJSONService()
			d.AddService(s.Service)

			// Advertise device name and service's UUIDs.
			d.AdvertiseNameAndServices("z3js", []gatt.UUID{s.Service.UUID()})

			// Advertise as an OpenBeacon iBeacon
			//d.AdvertiseIBeacon(gatt.MustParseUUID("AA6062F098CA42118EC4193EB73CCEB6"), 1, 2, -59)

		default:
		}
	}

	d.Init(onStateChanged)
	select {}
}
