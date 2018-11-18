package net

import (
	"fmt"
	"log"

	"github.com/tbehrsin/v8"
)

type Device interface {
	Gateway() Gateway
	GetEUI64() EUI64
	GetModel() string
	GetManufacturer() string
}

var currentDevices = make(map[*v8.Context]*DeviceProxy)

type DeviceProxy struct {
	gateway    *GatewayProxy
	device     Device
	controller *v8.Value
	value      *v8.Value
}

func NewDeviceProxy(g *GatewayProxy, d Device) *DeviceProxy {
	p := &DeviceProxy{
		gateway: g,
		device:  d,
	}
	return p
}

func (d *DeviceProxy) EUI64() EUI64 {
	return d.device.GetEUI64()
}

func (d *DeviceProxy) onSubscribe(context *v8.Context, controller *v8.Value) {
	if dv, err := context.Create(d); err != nil {
		log.Println(err)
		return
	} else {
		d.value = dv
	}

	if _, err := controller.Call(context.Global(), d.value); err != nil {
		log.Println(err)
	}

	if d.controller == nil {
		d.value = nil
	}
}

func (d *DeviceProxy) onLeaveNetwork() {
	if d.controller == nil {
		return
	} else if m, err := d.controller.Get("onLeave"); err != nil {
		log.Println(err)
	} else if _, err := m.Call(d.controller); err != nil {
		log.Println(err)
	}
}

func (d *DeviceProxy) onUpdate() {
	if d.controller == nil {
		return
	} else if m, err := d.controller.Get("onUpdate"); err != nil {
		log.Println(err)
	} else if _, err := m.Call(d.controller); err != nil {
		log.Println(err)
	}
}

func (d *DeviceProxy) String() string {
	return fmt.Sprintf("[eui64:%s model:\"%s\" manufacturer:\"%s\" Device]", d.device.GetEUI64(), d.device.GetModel(), d.device.GetManufacturer())
}

func (d *DeviceProxy) V8FuncToString(in v8.CallbackArgs) (*v8.Value, error) {
	return in.Context.Create(d.String())
}

func (d *DeviceProxy) V8GetProps(in v8.CallbackArgs) (*v8.Value, error) {
	//var props map[string]interface{}
	/*if buf, err := json.Marshal(d.device); err != nil {
		return nil, err
	} else if err := json.Unmarshal(buf, &props); err != nil {
		return nil, err
	} else {*/
	return in.Context.Create(d.device)
	//}
}

func (d *DeviceProxy) V8FuncSubscribe(in v8.CallbackArgs) (*v8.Value, error) {
	if d.controller != nil {
		return nil, fmt.Errorf("subscription already exists for device %s", d.EUI64())
	}

	currentDevices[in.Context] = d
	defer delete(currentDevices, in.Context)

	if controller, err := in.Arg(0).New(); err != nil {
		return nil, err
	} else {
		d.controller = controller
	}
	return nil, nil
}

type DeviceShadow struct {
	device *DeviceProxy
}

func NewDeviceShadow(in v8.CallbackArgs) (*v8.Value, error) {
	if device, ok := currentDevices[in.Context]; !ok {
		return nil, fmt.Errorf("not a constructor")
	} else {
		shadow := &DeviceShadow{device}
		if jso, err := in.Context.Create(shadow); err != nil {
			return nil, err
		} else {
			return jso, nil
		}
	}
}

func (d *DeviceShadow) V8FuncToString(in v8.CallbackArgs) (*v8.Value, error) {
	return d.device.V8FuncToString(in)
}

func (d *DeviceShadow) V8GetProps(in v8.CallbackArgs) (*v8.Value, error) {
	return d.device.V8GetProps(in)
}

// returns a promise
func (d *DeviceShadow) V8FuncSend(in v8.CallbackArgs) (*v8.Value, error) {
	return nil, nil
}

// returns a promise
func (d *DeviceShadow) V8FuncRead(in v8.CallbackArgs) (*v8.Value, error) {
	return nil, nil
}

func (d *DeviceShadow) V8FuncAddEventListener(in v8.CallbackArgs) (*v8.Value, error) {
	return nil, nil
}

func (d *DeviceShadow) V8FuncRemoveEventListener(in v8.CallbackArgs) (*v8.Value, error) {
	return nil, nil
}
