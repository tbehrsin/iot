package net

import (
	"sync"

	"github.com/behrsin/go-v8"
)

type GatewayController struct {
	Context  *v8.Context
	Instance *v8.Value
}

type Gateway interface {
	Protocol() string
	Start(n *Network) error
	Stop() error
}

type GatewayProxy struct {
	network *Network
	gateway Gateway

	devices     *sync.Map
	controllers *sync.Map

	Controller func(v8.FunctionArgs) (*Controller, error) `v8:"Controller"`
}

func NewGatewayProxy(n *Network, g Gateway) *GatewayProxy {
	p := &GatewayProxy{
		network:     n,
		gateway:     g,
		devices:     &sync.Map{},
		controllers: &sync.Map{},
		Controller:  NewController,
	}
	return p
}

func (p *GatewayProxy) Protocol() string {
	return p.gateway.Protocol()
}

func (p *GatewayProxy) AddDevice(d Device) {
	dp := NewDeviceProxy(p.network, p, d)
	p.devices.Store(d.GetEUI64(), dp)
	p.onJoinNetwork(dp)
}

func (p *GatewayProxy) RemoveDevice(d Device) {
	if dp, ok := p.devices.Load(d.GetEUI64()); ok {
		dp.(*DeviceProxy).onLeaveNetwork()
		p.devices.Delete(d.GetEUI64())
	}
}

func (p *GatewayProxy) UpdateDevice(d Device) {
	if dp, ok := p.devices.Load(d.GetEUI64()); ok {
		if dp.(*DeviceProxy).controller == nil {
			p.onJoinNetwork(dp.(*DeviceProxy))
		} else {
			dp.(*DeviceProxy).onUpdate()
		}
	}
}

func (p *GatewayProxy) onJoinNetwork(d *DeviceProxy) {
	p.controllers.Range(func(k, v interface{}) bool {
		context := k.(*v8.Context)
		controller := v.(*v8.Value)

		d.onSubscribe(context, controller)
		return d.controller != nil
	})

	if d.controller == nil {
		d.Holder().Publish()
	}
}

func (p *GatewayProxy) V8FuncSubscribe(in v8.FunctionArgs) (*v8.Value, error) {
	controller := in.Arg(0)
	p.controllers.Store(in.Context, controller)

	in.Context.GetIsolate().AddShutdownHook(func(i *v8.Isolate) {
		p.controllers.Delete(in.Context)
	})

	p.devices.Range(func(k, v interface{}) bool {
		d := v.(*DeviceProxy)
		if d.controller == nil {
			d.onSubscribe(in.Context, controller)
		}
		return true
	})
	return nil, nil
}
