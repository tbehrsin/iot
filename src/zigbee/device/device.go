package device

import (
	"encoding/json"
	"fmt"
	"gateway/zigbee"
	"math/rand"
	"time"

	"github.com/boltdb/bolt"
)

func init() {
	// Ensure Rand.Intn has some ability to generate different sequences
	rand.Seed(time.Now().UTC().UnixNano())
}

const DeviceBucket = "Device"

type Device struct {
	zigbee.DeviceMessage
	table      *Table
	advertised bool
}

// NewDevice creates a new device object
// Since the device does not currently exist, there is
// no point attempting to find it's NodeID within the bolt
// database. Rather, the NodeID needs to be derived from the
// table given to it as a parameter.
// After that, it should function similar to update.
func (t *Table) NewDevice() (*Device, error) {
	d := &Device{
		table: t,
	}
	// Initializing dbresult to start false(a space_ot_found).
	// This loop is intended to exit when the randomly
	// generated NodeId finds a space in the database
	// i.e. the get function returns nil for a key that
	// does not exist
	db := d.table.db
	id := fmt.Sprintf("%s", d.NodeID)

	dbresult := false
	for dbresult == false {
		d.NodeID = zigbee.NodeID{zigbee.UInt16{uint16(rand.Intn(0x10000))}}
		d.Type = zigbee.DeviceType{zigbee.UInt16{1}}
		d.TimeSinceLastMessage = 1
		d.Endpoint.Clusters = []zigbee.Cluster{
			zigbee.Cluster{zigbee.ClusterType{}, zigbee.ClusterID{zigbee.UInt16{0x0000}}},
			zigbee.Cluster{zigbee.ClusterType{true}, zigbee.ClusterID{zigbee.UInt16{0x0006}}},
			zigbee.Cluster{zigbee.ClusterType{true}, zigbee.ClusterID{zigbee.UInt16{0x0008}}},
			zigbee.Cluster{zigbee.ClusterType{true}, zigbee.ClusterID{zigbee.UInt16{0x0300}}},
		}

		// check that this new NodeID is not already present
		db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(DeviceBucket))
			v := b.Get([]byte(id))
			if v == nil {
				dbresult = true
			}
			return nil
		})
	}

	return d, nil
}

// Update finds a device record within the BoltDB
// (using NodeID as the key)
// and replaces it with a db.put command.
func (d *Device) Update() (*Device, error) {
	db := d.table.db
	id := fmt.Sprintf("%s", d.NodeID)

	d.table.devices.Store(id, d)
	if buf, err := json.Marshal(d); err != nil {
		return nil, err
	} else if err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(DeviceBucket))
		return b.Put([]byte(id), buf)
	}); err != nil {
		return nil, err
	} else {
		if err := d.advertise(); err != nil {
			return nil, err
		} else {
			return d, nil
		}
	}
}

func (d *Device) String() string {
	return fmt.Sprintf("{%s %s %s %d %s %s}", d.NodeID, d.State, d.Type, d.TimeSinceLastMessage, d.Endpoint.EUI64, d.Endpoint.Endpoint)
}

// Delete finds a device record within the BoltDB
// (using NodeID as the key)
func (d *Device) Delete() error {
	gw := d.table.gateway
	db := d.table.db
	id := fmt.Sprintf("%s", d.NodeID)

	d.table.devices.Delete(id)
	gw.Publish(zigbee.DeviceLeft, &zigbee.DeviceLeftMessage{
		EUI64: d.Endpoint.EUI64,
	})

	if err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(DeviceBucket))
		return b.Delete([]byte(id))
	}); err != nil {
		return err
	} else {
		return nil
	}
}
func (d *Device) advertise() error {
	gw := d.table.gateway
	d.advertised = true
	return gw.Publish(zigbee.DeviceJoined, d)
}

//DeviceStateChange is taking its cue from NewDevice
// as far as publishing a status change message goes
// DB functionality (if any) probably wants to be closer to Update, tho
// Testing reveals that this procedure panics if fed a null device
// so now we wiil attempt to test for and do nothing upon that condition

// gw:=d.table.gateway is no good. Can't be gw (compiler's responsibility)
// and d must have been supplied to the method(but might not exist).
// Most likely culprit is an unset table.gateway.

func (d *Device) DeviceStateChange() error {
	fmt.Println("DeviceStateChange called")
	if (d != nil) && (d.NodeID != zigbee.NodeID{zigbee.UInt16{uint16(0)}}) {
		fmt.Println("Setting up gateway table")

		// So THIS statement is the actual villain of the piece!
		gw := d.table.gateway
		// db := d.table.db
		// id := fmt.Sprintf("%s", d.NodeID)
		// d.table.devices.StateChange(id)?
		fmt.Println("Double-check that we are not surviving gw:=d.table.gateway")
		fmt.Printf("eui64 = %v, state = %v before message build", d.Endpoint.EUI64, d.State)
		dscm := zigbee.DeviceStateChangeMessage{
			EUI64: d.Endpoint.EUI64, State: d.State}
		// Still no joy.
		// Starting to wonder if something, somewhere, is treating 0x030
		// as an address. So lets see if either eui64 or state contain 0x30.
		fmt.Printf("eui64 = %v, state = %v pre-publish", d.Endpoint.EUI64, d.State)
		if (&dscm != nil) && (&d.State != nil) && (&d.Endpoint.EUI64 != nil) {
			return gw.Publish(zigbee.DeviceStateChange, &dscm)
		}
	}
	return nil
}

// Device list message publication
// probably wants to be invoked from
// somewhere wihin the existing device list command

// two current issues
// 1 listing multiple devices
// hence need to loop over the devices in the table
// rather than attacking one device
// 2 devices field is in the message, not the device table
// so it probably needs to be assembled from each device
// found by the loop

//This thing isn't playing here at all
//May be better to include it with the device list command
//

//func (d *Device) DeviceList(r *REPL) error {

//	table := r.Gateway.GetDeviceTable()
//	var localdevices []DeviceMessage
// table.Range(func(d *device.Device) bool {
// append nodeid to devices, instead of printing
// localdevices[d] = d.NodeID
//	})

//	gw := d.table.gateway

//	return gw.Publish(zigbee.DeviceList, &zigbee.DeviceListMessage{
//		Devices: localdevices,
//	})
//}
