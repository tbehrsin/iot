package net

import (
	"api"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"github.com/behrsin/go-v8"
)

type Device interface {
	Gateway() Gateway
	GetEUI64() EUI64
	GetModel() string
	GetManufacturer() string
	Match(*v8.Value) (bool, error)
}

var currentDevices = make(map[*v8.Context]*DeviceProxy)

type DeviceProxy struct {
	network    *Network
	gateway    *GatewayProxy
	device     Device
	controller *Controller
	value      *v8.Value
}

func NewDeviceProxy(n *Network, g *GatewayProxy, d Device) *DeviceProxy {
	p := &DeviceProxy{
		network: n,
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

	if _, err := controller.Call(nil, d.value); err != nil {
		log.Println(err)
	}

	if d.controller == nil {
		d.value = nil
	} else {
		d.Holder().Publish()
	}
}

func (d *DeviceProxy) onLeaveNetwork() {
	if d.controller == nil {
		return
	} else if _, err := d.controller.value.CallMethod("onLeave"); err != nil {
		log.Println(err)
	} else {
		log.Println("onLeaveNetwork", err)
	}
}

func (d *DeviceProxy) onUpdate() {
	if d.controller == nil {
		return
	} else if _, err := d.controller.value.CallMethod("onUpdate"); err != nil {
		log.Println("onUpdate", err)
	}
}

func (d *DeviceProxy) String() string {
	return fmt.Sprintf("[eui64:%s model:\"%s\" manufacturer:\"%s\" Device]", d.device.GetEUI64(), d.device.GetModel(), d.device.GetManufacturer())
}

func (d *DeviceProxy) GetName() string {
	defaultName := fmt.Sprintf("%s %s", d.device.GetManufacturer(), d.device.GetModel())

	if d.controller == nil {
		return defaultName
	} else {
		return d.controller.Name
	}
}

func (d *DeviceProxy) SetName(name string) {
	defaultName := fmt.Sprintf("%s %s", d.device.GetManufacturer(), d.device.GetModel())

	if d.controller == nil {
		return
	} else if name != "" {
		d.controller.Name = name
	} else {
		d.controller.Name = defaultName
	}
}

func (d *DeviceProxy) GetState() map[string]interface{} {
	if d.controller == nil {
		return nil
	}

	return d.controller.state
}

func (d *DeviceProxy) SetState(state map[string]interface{}) error {
	if d.controller == nil {
		return fmt.Errorf("no controller for setState")
	}

	d.controller.SetState(state, false)

	return nil
}

func (d *DeviceProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	backend := d.controller.Application().Backend()
	index := d.controller.Index

	publicPath := d.controller.Application().Package().Public()
	path := filepath.Join("/", publicPath, r.URL.Path)
	indexPath := filepath.Join("/", publicPath, index)

	if b, err := backend.ReadFile(path); err == nil {
		w.Write(b)
	} else if b, err := backend.ReadFile(indexPath); err != nil {
		w.Write(b)
	} else {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 Not Found"))
	}
}

func (d *DeviceProxy) V8FuncToString(in v8.FunctionArgs) (*v8.Value, error) {
	return in.Context.Create(d.String())
}

func (d *DeviceProxy) V8GetDevice(in v8.GetterArgs) (*v8.Value, error) {
	return in.Context.Create(d.device)
}

func (d *DeviceProxy) V8FuncMatch(in v8.FunctionArgs) (*v8.Value, error) {
	if matches, err := d.device.Match(in.Arg(0)); err != nil {
		return nil, err
	} else if matches {
		return in.Context.True(), nil
	} else {
		return in.Context.False(), nil
	}
}

func (d *DeviceProxy) V8FuncSubscribe(in v8.FunctionArgs) (*v8.Value, error) {
	if d.controller != nil {
		return nil, fmt.Errorf("subscription already exists for device %s", d.EUI64())
	}

	currentDevices[in.Context] = d
	defer delete(currentDevices, in.Context)

	if controller, err := in.Arg(0).New(); err != nil {
		return nil, err
	} else {
		d.controller = controller.Receiver(api.ControllerType).Interface().(*Controller)
	}
	return nil, nil
}

type deviceHolder struct {
	deviceProxy  *DeviceProxy
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Manufacturer string                 `json:"manufacturer"`
	Model        string                 `json:"model"`
	State        map[string]interface{} `json:"state"`
}

func (d *DeviceProxy) Holder() *deviceHolder {
	return &deviceHolder{
		deviceProxy:  d,
		ID:           d.EUI64().String(),
		Name:         d.GetName(),
		Manufacturer: d.device.GetManufacturer(),
		Model:        d.device.GetModel(),
		State:        d.GetState(),
	}
}

func (d *deviceHolder) Publish() error {
	if data, err := json.Marshal(d); err != nil {
		return err
	} else if token := d.deviceProxy.network.mqtt.Publish(fmt.Sprintf("iot/%s/notify", d.deviceProxy.EUI64()), 2, false, data); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	return nil
}
