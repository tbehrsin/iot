package device

import (
	"encoding/json"
	"fmt"
	"gateway/zigbee"
	"math/rand"

	"github.com/boltdb/bolt"
)

const DeviceBucket = "Device"

type Device struct {
	zigbee.DeviceMessage
	table *Table
}

func (t *Table) NewDevice() *Device {
	d := &Device{
		table: t,
	}
	d.NodeID = zigbee.NodeID{zigbee.UInt16{uint16(rand.Intn(0x10000))}}
	return d
}

func (d *Device) Update() error {
	db := d.table.db
	id := fmt.Sprintf("%s", d.NodeID)

	d.table.devices.Store(id, d)
	if buf, err := json.Marshal(d); err != nil {
		return err
	} else if err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(DeviceBucket))
		return b.Put([]byte(id), buf)
	}); err != nil {
		return err
	} else {
		return nil
	}
}

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
