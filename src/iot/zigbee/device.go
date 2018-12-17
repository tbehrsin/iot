package zigbee

import (
	"fmt"
	"iot/events"
	"iot/net"
	"log"
	"sync"
	"time"

	"github.com/behrsin/go-v8"
)

type Device struct {
	gateway         *Gateway
	mutex           *sync.Mutex
	attributeEvents events.EventEmitter

	EUI64      EUI64            `v8:"eui64"`
	NodeID     NodeID           `v8:"nodeId"`
	State      State            `v8:"state"`
	DeviceType DeviceType       `v8:"type"`
	LastSeen   time.Time        `v8:"lastSeen"`
	Endpoints  []DeviceEndpoint `v8:"endpoints"`

	Model        string `v8:"model"`
	Manufacturer string `v8:"manufacturer"`
}

func NewDevice(g *Gateway, m *DeviceMessage) *Device {
	d := &Device{
		gateway:   g,
		mutex:     &sync.Mutex{},
		Endpoints: []DeviceEndpoint{},
	}
	d.merge(m)
	return d
}

func (d *Device) GetEUI64() net.EUI64 {
	return net.EUI64(d.EUI64)
}

func (d *Device) Gateway() net.Gateway {
	return d.gateway
}

func (d *Device) GetModel() string {
	return d.Model
}

func (d *Device) GetManufacturer() string {
	return d.Manufacturer
}

func (d *Device) merge(m *DeviceMessage) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.EUI64 = m.Endpoint.EUI64
	d.NodeID = m.NodeID
	d.State = m.State
	d.DeviceType = m.Type
	d.LastSeen = time.Now().Add(-time.Duration(m.TimeSinceLastMessage) * time.Second)

	ok := false
	for i, e := range d.Endpoints {
		if e.EUI64 == m.Endpoint.EUI64 && e.Endpoint == m.Endpoint.Endpoint {
			d.Endpoints[i] = m.Endpoint
			ok = true
			break
		}
	}
	if !ok {
		d.Endpoints = append(d.Endpoints, m.Endpoint)
	}
}

func (d *Device) V8FuncRead(in v8.FunctionArgs) (*v8.Value, error) {
	clusterId := ClusterID{V8Uint16(in.Arg(0))}
	attributeId := AttributeID{V8Uint16(in.Arg(1))}
	ep := Endpoint(in.Arg(2).Float64())

	d.gateway.commands <- CommandListMessage{[]Command{
		Command{fmt.Sprintf("zcl global read %s %s", clusterId, attributeId), 0},
		Command{fmt.Sprintf("plugin device-table send %s %s", EUI64(d.EUI64).bracketString(), ep), 0},
	}}

	if p, err := in.Context.NewPromise(); err != nil {
		return nil, err
	} else {
		// start a goroutined to listen for the attribute being read
		go func() {
			d.attributeEvents.AddOnceListener(fmt.Sprintf("%s:%s:%s", ep, clusterId, attributeId), func(args ...interface{}) {
				err := args[0].(error)
				if err != nil {
					if errorObject, err2 := in.Context.Create(err); err2 != nil {
						panic(err2)
					} else {
						p.Reject.Call(nil, errorObject)
					}
				} else if value, err := in.Context.Create(args[1]); err != nil {
					if errorObject, err2 := in.Context.Create(err); err2 != nil {
						panic(err2)
					} else {
						p.Reject.Call(nil, errorObject)
					}
				} else {
					p.Resolve.Call(nil, value)
				}
			})
		}()

		return p.Value, nil
	}
}

func (d *Device) V8FuncSend(in v8.FunctionArgs) (*v8.Value, error) {
	clusterId := ClusterID{V8Uint16(in.Arg(0))}
	commandId := CommandID{V8Uint8(in.Arg(1))}
	ep := Endpoint(in.Arg(2).Float64())
	log.Println(in.Arg(3))
	log.Println(in.Arg(3).Bytes())
	data := append([]byte{0x01, 0x00, byte(commandId.Value)}, in.Arg(3).Bytes()...)

	d.gateway.commands <- CommandListMessage{[]Command{
		Command{fmt.Sprintf("raw %s %s", clusterId, CommandData(data).bracketString()), 0},
		Command{fmt.Sprintf("send %s 0x01 %s", d.NodeID, ep), 0},
	}}
	// d.gateway.commands <- fmt.Sprintf("send %s %s %s", d.NodeID, ep, ep)

	return nil, nil
}

// FindEndpoint finds the provided endpoint, if it does not exist then return nil
func (d *Device) FindEndpoint(endpointId Endpoint) *DeviceEndpoint {
	for _, endpoint := range d.Endpoints {
		if endpoint.Endpoint == endpointId {
			return &endpoint
		}
	}
	return nil
}

func (d *Device) Match(query *v8.Value) (bool, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if vendpoints, err := query.Get("endpoints"); err != nil {
		return false, err
	} else if vlength, err := vendpoints.Get("length"); err != nil {
		return false, err
	} else if length := int(vlength.Int64()); length > 0 {
		for i := 0; i < length; i++ {
			var endpointId Endpoint
			var vendpoint *v8.Value

			if vendpoint, err = vendpoints.GetIndex(i); err != nil {
				return false, err
			} else if vendpointId, err := vendpoint.Get("id"); err != nil {
				return false, err
			} else {
				endpointId = Endpoint(vendpointId.Int64())
			}

			if endpoint := d.FindEndpoint(endpointId); endpoint == nil {
				return false, nil
			} else {
				if vclusters, err := vendpoint.Get("clusters"); err != nil {
					return false, err
				} else if b, err := endpoint.V8MatchAll(vclusters); err != nil {
					return false, err
				} else if !b {
					return false, nil
				}
			}
		}
	}

	return true, nil
}
