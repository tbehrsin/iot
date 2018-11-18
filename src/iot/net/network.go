package net

import (
	"log"
	"sync"
)

type Network struct {
	gateways *sync.Map
}

func New() *Network {
	return &Network{
		gateways: &sync.Map{},
	}
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
