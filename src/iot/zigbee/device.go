package zigbee

import (
	"iot/net"
	"sync"
	"time"
)

type device struct {
	gateway *Gateway    `json:"-"`
	mutex   *sync.Mutex `json:"-"`

	EUI64      net.EUI64        `v8:"eui64"`
	NodeID     nodeID           `v8:"nodeId"`
	State      state            `v8:"state"`
	DeviceType deviceType       `v8:"type"`
	LastSeen   time.Time        `v8:"lastSeen"`
	Endpoints  []deviceEndpoint `v8:"endpoints"`

	Model        string `v8:"model"`
	Manufacturer string `v8:"manufacturer"`
}

func NewDevice(g *Gateway, m *deviceMessage) *device {
	d := &device{
		gateway:   g,
		mutex:     &sync.Mutex{},
		Endpoints: []deviceEndpoint{},
	}
	d.merge(m)
	return d
}

func (d *device) GetEUI64() net.EUI64 {
	return d.EUI64
}

func (d *device) Gateway() net.Gateway {
	return d.gateway
}

func (d *device) GetModel() string {
	return d.Model
}

func (d *device) GetManufacturer() string {
	return d.Manufacturer
}

func (d *device) merge(m *deviceMessage) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.EUI64 = net.EUI64(m.Endpoint.EUI64)
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
