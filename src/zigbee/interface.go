package main

import (
	"gateway/net"
	"gateway/zigbee"
	"time"
	"zigbee/device"
)

type Gateway struct {
	gateway   *zigbee.Gateway
	running   bool
	networkUp bool
	devices   *device.Table
}

func (gw *Gateway) Start() {
	if gw.running == true {
		return
	}

	gw.running = true

	go gw.SendHeartbeats(5 * time.Second)
	go gw.OnPublishState(zigbee.EUI64(gw.gateway.EUI64()), zigbee.PublishStateMessage{})
}

func (gw *Gateway) Stop() {
	gw.running = false
}

func (gw *Gateway) GetDeviceTable() *device.Table {
	if gw.devices == nil {
		var err error
		if gw.devices, err = device.LoadTable("zigbee-gateway.db", gw.gateway); err != nil {
			panic(err)
		}
	}
	return gw.devices
}

func (gw *Gateway) OnHeartbeat(eui64 zigbee.EUI64, message zigbee.HeartbeatMessage)                 {}
func (gw *Gateway) OnDeviceList(eui64 zigbee.EUI64, message zigbee.DeviceListMessage)               {}
func (gw *Gateway) OnRelayList(eui64 zigbee.EUI64, message zigbee.RelayListMessage)                 {}
func (gw *Gateway) OnSettings(eui64 zigbee.EUI64, message zigbee.SettingsMessage)                   {}
func (gw *Gateway) OnDeviceJoined(eui64 zigbee.EUI64, message zigbee.DeviceMessage)                 {}
func (gw *Gateway) OnDeviceLeft(eui64 zigbee.EUI64, message zigbee.DeviceLeftMessage)               {}
func (gw *Gateway) OnDeviceStateChange(eui64 zigbee.EUI64, message zigbee.DeviceStateChangeMessage) {}
func (gw *Gateway) OnOTAEvent(eui64 zigbee.EUI64, message zigbee.OTAEventMessage)                   {}
func (gw *Gateway) OnExecuted(eui64 zigbee.EUI64, message zigbee.ExecutedMessage)                   {}
func (gw *Gateway) OnZCLResponse(eui64 zigbee.EUI64, message zigbee.ZCLResponseMessage)             {}
func (gw *Gateway) OnZDOResponse(eui64 zigbee.EUI64, message zigbee.ZDOResponseMessage)             {}
func (gw *Gateway) OnAPSResponse(eui64 zigbee.EUI64, message zigbee.APSResponseMessage)             {}

func (gw *Gateway) OnCommandList(eui64 zigbee.EUI64, message zigbee.CommandListMessage) {

}

func (gw *Gateway) OnPublishState(eui64 zigbee.EUI64, message zigbee.PublishStateMessage) {
	e, _ := net.NewEUI64("A1B2C3D4E5F6A7B8")

	gw.gateway.Publish(zigbee.DeviceList, zigbee.DeviceListMessage{Devices: []zigbee.DeviceMessage{
		zigbee.DeviceMessage{
			NodeID:               zigbee.NodeID{UInt16: zigbee.UInt16{Value: 0x9C3D}},
			State:                zigbee.State(zigbee.StateJoined),
			Type:                 zigbee.DeviceType{UInt16: zigbee.UInt16{Value: 0x0000}},
			TimeSinceLastMessage: 0x00000010,
			Endpoint: zigbee.DeviceEndpoint{
				DeviceEndpointInfo: zigbee.DeviceEndpointInfo{
					EUI64:    zigbee.EUI64(e),
					Endpoint: zigbee.Endpoint(0x01),
				},
				Clusters: []zigbee.Cluster{
					zigbee.Cluster{Type: zigbee.ClusterType{Out: true}, ID: zigbee.ClusterID{UInt16: zigbee.UInt16{Value: 0x0000}}},
					zigbee.Cluster{Type: zigbee.ClusterType{Out: false}, ID: zigbee.ClusterID{UInt16: zigbee.UInt16{Value: 0x0001}}},
				},
			},
		},
		zigbee.DeviceMessage{
			NodeID:               zigbee.NodeID{UInt16: zigbee.UInt16{Value: 0x9C3D}},
			State:                zigbee.State(zigbee.StateJoined),
			Type:                 zigbee.DeviceType{UInt16: zigbee.UInt16{Value: 0x0000}},
			TimeSinceLastMessage: 0x00000010,
			Endpoint: zigbee.DeviceEndpoint{
				DeviceEndpointInfo: zigbee.DeviceEndpointInfo{
					EUI64:    zigbee.EUI64(e),
					Endpoint: zigbee.Endpoint(0x02),
				},
				Clusters: []zigbee.Cluster{
					zigbee.Cluster{Type: zigbee.ClusterType{Out: true}, ID: zigbee.ClusterID{UInt16: zigbee.UInt16{Value: 0x0005}}},
					zigbee.Cluster{Type: zigbee.ClusterType{Out: true}, ID: zigbee.ClusterID{UInt16: zigbee.UInt16{Value: 0x0006}}},
				},
			},
		},
	}})
	gw.gateway.Publish(zigbee.RelayList, zigbee.RelayListMessage{})
	gw.gateway.Publish(zigbee.Settings, zigbee.SettingsMessage{
		NCPStackVersion: "iot-zigbee-1.0",
		NetworkUp:       gw.networkUp,
	})
}

func (gw *Gateway) OnUpdateSettings(eui64 zigbee.EUI64, message zigbee.UpdateSettingsMessage) {

}

func (gw *Gateway) SendHeartbeats(d time.Duration) {
	for _ = range time.Tick(d) {
		if !gw.running {
			break
		}

		gw.gateway.Publish(zigbee.Heartbeat, zigbee.HeartbeatMessage{NetworkUp: gw.networkUp})
	}
}
