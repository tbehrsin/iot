package zigbee

import (
	"encoding/binary"
	"fmt"
	"gateway/events"
	"gateway/net"
	"strings"
	"sync"
	"time"

	"github.com/behrsin/go-v8"
)

type Device struct {
	events.Hub

	gateway *Gateway
	mutex   *sync.Mutex
	onNet   bool
	seq     uint8

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

func (d *Device) IsValid() bool {
	return d.Model != "" && d.Manufacturer != ""
}

func (d *Device) merge(m *DeviceMessage) {
	// d.mutex.Lock()
	// defer d.mutex.Unlock()

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

	var fetchManufacturer func()
	fetchManufacturer = func() {
		if d.Manufacturer == "" {
			for _, endpoint := range d.FindEndpointsForCluster(Cluster{ClusterType{}, ClusterID{UInt16{0x0000}}}) {
				d.gateway.commands <- CommandListMessage{[]Command{
					Command{"zcl global direction 0", 0},
					Command{"zcl global read 0x0000 0x0004", 0},
					Command{fmt.Sprintf("send %s 0x01 %s", d.NodeID, endpoint.Endpoint), 0},
				}}

				go func(endpoint DeviceEndpoint) {
					channel := d.Once(fmt.Sprintf("attr:%d:0:4", endpoint.Endpoint))

					e := <-channel.Receive()

					if d.Manufacturer != "" {
						return
					} else {
						d.Manufacturer = e.Args[0].([]interface{})[0].(string)

						if d.IsValid() {
							d.Emit("merge")
						}
					}
				}(endpoint)
			}
		}
	}
	go fetchManufacturer()

	var fetchModel func()
	fetchModel = func() {
		if d.Model == "" {
			for _, endpoint := range d.FindEndpointsForCluster(Cluster{ClusterType{}, ClusterID{UInt16{0x0000}}}) {
				d.gateway.commands <- CommandListMessage{[]Command{
					Command{"zcl global direction 0", 0},
					Command{"zcl global read 0x0000 0x0005", 0},
					Command{fmt.Sprintf("send %s 0x01 %s", d.NodeID, endpoint.Endpoint), 0},
				}}

				go func(endpoint DeviceEndpoint) {
					channel := d.Once(fmt.Sprintf("attr:%d:0:5", endpoint.Endpoint))

					e := <-channel.Receive()

					if d.Model != "" {
						return
					} else {
						d.Model = e.Args[0].([]interface{})[0].(string)

						if d.IsValid() {
							d.Emit("merge")
						}
					}
				}(endpoint)
			}
		}
	}
	go fetchModel()

}

func (d *Device) V8FuncRead(in v8.FunctionArgs) (*v8.Value, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	ep := Endpoint(in.Arg(0).Float64())
	clusterId := ClusterID{V8Uint16(in.Arg(1))}

	attributes := make([]uint16, len(in.Args)-2)
	for i, arg := range in.Args[2:] {
		attributes[i] = V8Uint16(arg).Value
	}

	list := make([]string, len(attributes))
	buf := make([]byte, 3+2*len(attributes))
	buf[0] = 0x00
	buf[1] = d.seq
	d.seq++

	for i, attribute := range attributes {
		binary.LittleEndian.PutUint16(buf[3+i*2:], attribute)
		list[i] = fmt.Sprintf("%d", attribute)
	}

	if r, err := in.Context.NewResolver(); err != nil {
		return nil, err
	} else {
		channel := d.Once(fmt.Sprintf("attr:%d:%d:%s", ep, clusterId.Value, strings.Join(list, ":")))

		go func() {
			select {
			case <-time.After(30 * time.Second):
				if err := r.Reject(fmt.Errorf("read timeout")); err != nil {
					panic(err)
				}
			case e := <-channel.Receive():
				if err := r.Resolve(e.Args[0]); err != nil {
					panic(err)
				}
			}
		}()

		d.gateway.commands <- CommandListMessage{[]Command{
			Command{fmt.Sprintf("raw %s %s", clusterId, CommandData(buf).bracketString()), 0},
			Command{fmt.Sprintf("send %s 0x01 %s", d.NodeID, ep), 0},
		}}

		return r.Promise(), nil
	}
}

func (d *Device) V8FuncSend(in v8.FunctionArgs) (*v8.Value, error) {
	clusterId := ClusterID{V8Uint16(in.Arg(0))}
	commandId := CommandID{V8Uint8(in.Arg(1))}
	ep := Endpoint(in.Arg(2).Float64())
	data := append([]byte{0x01, 0x00, byte(commandId.Value)}, in.Arg(3).Bytes()...)

	d.gateway.commands <- CommandListMessage{[]Command{
		Command{fmt.Sprintf("raw %s %s", clusterId, CommandData(data).bracketString()), 0},
		Command{fmt.Sprintf("send %s 0x01 %s", d.NodeID, ep), 0},
	}}

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

func (d *Device) FindEndpointsForCluster(cluster Cluster) []DeviceEndpoint {
	out := []DeviceEndpoint{}
	for _, endpoint := range d.Endpoints {
		if endpoint.Match(cluster) {
			out = append(out, endpoint)
		}
	}
	return out
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
