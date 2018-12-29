package net

import (
	"encoding/json"
	"log"
	"strings"
	"sync"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Network struct {
	gateways *sync.Map
	mqtt     mqtt.Client
}

func New(client mqtt.Client) *Network {
	n := &Network{
		gateways: &sync.Map{},
		mqtt:     client,
	}

	go func() {
		if token := client.Subscribe("iot/+/publish", 2, n.onPublish); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
	}()

	return n
}

func (n *Network) AddGateway(g Gateway) {
	proxy := NewGatewayProxy(n, g)
	n.gateways.Store(g.Protocol(), proxy)
	go func() {
		if err := g.Start(n); err != nil {
			log.Println(err)
		}
	}()
}

func (n *Network) Gateways() []*GatewayProxy {
	gateways := []*GatewayProxy{}
	n.gateways.Range(func(k, v interface{}) bool {
		gateways = append(gateways, v.(*GatewayProxy))
		return true
	})
	return gateways
}

func (n *Network) AddDevice(d Device) {
	if g, ok := n.gateways.Load(d.Gateway().Protocol()); ok {
		g.(*GatewayProxy).AddDevice(d)
	}
}

func (n *Network) RemoveDevice(d Device) {
	if g, ok := n.gateways.Load(d.Gateway().Protocol()); ok {
		g.(*GatewayProxy).RemoveDevice(d)
	}
}

func (n *Network) UpdateDevice(d Device) {
	if g, ok := n.gateways.Load(d.Gateway().Protocol()); ok {
		g.(*GatewayProxy).UpdateDevice(d)
	}
}

func (n *Network) FindDevice(eui64 EUI64) *DeviceProxy {
	var device *DeviceProxy
	n.gateways.Range(func(k, v interface{}) bool {
		gateway := v.(*GatewayProxy)

		if d, ok := gateway.devices.Load(eui64); !ok {
			return true
		} else {
			device = d.(*DeviceProxy)
			return false
		}
	})
	return device
}

func (n *Network) onPublish(client mqtt.Client, message mqtt.Message) {
	go func() {
		var holder deviceHolder

		if e, err := NewEUI64(strings.TrimPrefix(strings.TrimSuffix(message.Topic(), "/publish"), "iot/")); err != nil {
			log.Println(err)
		} else if err := json.Unmarshal(message.Payload(), &holder); err != nil {
			log.Println(err)
		} else if d := n.FindDevice(e); d != nil {
			d.SetState(holder.State)
		}
	}()
}
