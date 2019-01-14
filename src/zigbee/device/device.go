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
	table *Table
}

// NewDevice creates a new device object
// Since the device does not currently exist, there is
// no point attempting to find it's NodeID within the bolt
// database. Rather, the NodeID needs to be derived from the
// table given to it as a parameter.
// After that, it should function similar to update.
// Since update is called (in commands.go) almost immediately
// after NewDevice, arguably, all NewDevice needs to do
// is to return NodeID as the new device.
// HOWEVER, the device data for the new device
// MUST be established before update is allowed to run,
// so here appears to be the logical place to attack that TODO.
// With code similar to that within update, but JSON-marshalling
// from a user-given stream of input parameters that define the device
// to be added, rather than from a database record.
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

	return d.Update()
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
		return d, nil
	}
}

func (d *Device) String() string {
	return fmt.Sprintf("{%s %s %s %d %s %s}", d.NodeID, d.State, d.Type, d.TimeSinceLastMessage, d.Endpoint.EUI64, d.Endpoint.Endpoint)
}

// Delete finds a device record within the BoltDB
// (using NodeID as the key)
func (d *Device) Delete() error {
	db := d.table.db
	id := fmt.Sprintf("%s", d.NodeID)

	d.table.devices.Delete(id)

	if err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(DeviceBucket))
		return b.Delete([]byte(id))
	}); err != nil {
		return err
	} else {
		return nil
	}
}
