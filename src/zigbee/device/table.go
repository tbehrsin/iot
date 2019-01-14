package device

import (
	"encoding/json"
	"fmt"
	"gateway/zigbee"
	"sync"

	"github.com/boltdb/bolt"
)

type Table struct {
	devices sync.Map
	db      *bolt.DB
	gateway *zigbee.Gateway
}

func LoadTable(path string, gateway *zigbee.Gateway) (*Table, error) {
	if db, err := bolt.Open(path, 0600, nil); err != nil {
		return nil, err
	} else {
		if err := db.Update(func(tx *bolt.Tx) error {
			if _, err := tx.CreateBucketIfNotExists([]byte(DeviceBucket)); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return nil, err
		}

		table := &Table{
			db:      db,
			gateway: gateway,
		}

		if err := db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(DeviceBucket))
			if err := b.ForEach(func(_, v []byte) error {
				var device Device
				if err := json.Unmarshal(v, &device); err != nil {
					return err
				} else {
					table.devices.Store(fmt.Sprintf("%s", device.NodeID), &device)
				}
				return nil
			}); err != nil {
				return err
			}

			return nil
		}); err != nil {
			return nil, err
		} else {

			return table, nil
		}
	}
}

func (t *Table) Get(id zigbee.NodeID) (*Device, error) {
	n := fmt.Sprintf("%s", id)
	if d, ok := t.devices.Load(n); !ok {
		return nil, fmt.Errorf("device with node-id %s not found", n)
	} else {
		return d.(*Device), nil
	}
}

func (t *Table) Range(callback func(device *Device) bool) {
	t.devices.Range(func(_, v interface{}) bool {
		return callback(v.(*Device))
	})
}
