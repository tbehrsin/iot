package device

import (
	"testing"
)

func TestDeviceStateChange(t *testing.T) {
	var d Device

	d = Device{}

	// first test, fail expected, trying to state change a nil device

	if d.DeviceStateChange() == nil {
		// Test has failed because DeviceStateChange ought to balk at an empty device
		t.Error("DeviceStateChange has published a change upon an empty device")
	}

	// Further tests require passable data within the device structure
	// Numbers given below came from older test data
	// The old device appears short for Testing.
	// May be better to 'borrow' newdevice/delete to create one for Testing

	// d.NodeID = zigbee.NodeID{zigbee.UInt16{uint16(0xBA5B)}}
	// eui64, _ := net.NewEUI64("2F91EC429F3FDB91")
	// d.Endpoint.EUI64 = zigbee.EUI64(eui64)
	var table Table
	var r struct {
		Gateway *Gateway
	}

	//set table to something sensible here?
	//borrow from commands and repl

	table = r.Gateway.GetDeviceTable()

	// d, err := table.NewDevice()
	defer d.Delete()

	d.State = 0x01

	if d.DeviceStateChange() != nil {
		t.Error("DeviceStateChange erroring whilst publishing a state of 0x01")
	}
	d.State = 0x00

	if d.DeviceStateChange() != nil {
		t.Error("DeviceStateChange erroring whilst publishing a state of 0x00")
	}
}
