package device

import (
	"gateway/net"
	"gateway/zigbee"
	"testing"
)

func TestDeviceStateChange(t *testing.T) {
	var d Device

	// first test, fail expected, trying to state change a nil device

	d = Device{}

	if d.DeviceStateChange() == nil {
		// Test has failed because DeviceStateChange ought to balk at an empty device
		t.Errorf("DeviceStateChange has published a change upon an empty device")
	}

	// Further tests require passable data within the device structure
	// Numbers given below came from older test data
	d.NodeID = zigbee.NodeID{zigbee.UInt16{uint16(0xBA5B)}}
	eui64, _ := net.NewEUI64("2F91EC429F3FDB91")
	d.Endpoint.EUI64 = zigbee.EUI64(eui64)
	d.State = 0x01
}
