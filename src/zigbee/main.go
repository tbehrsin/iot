package main

import (
	"gateway/net"
	"gateway/zigbee"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var eui64 net.EUI64
	var err error

	if eui64, err = net.NewEUI64("A5C789101A5DC16A"); err != nil {
		log.Fatal(err)
	}

	gateway := &Gateway{gateway: zigbee.NewGateway()}
	gateway.gateway.SetEUI64(zigbee.EUI64(eui64))
	gateway.gateway.SetInterface(gateway)

	gateway.gateway.Start(nil)
	go gateway.Start()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	gateway.Stop()
}
