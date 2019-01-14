package main

import (
	"crypto/md5"
	"fmt"
	iotnet "gateway/net"
	"gateway/zigbee"
	"io"
	"zigbee/device"
)

// AddCommand adds a device to the device table
func AddCommand(r *REPL, c *Command) error {
	table := r.Gateway.GetDeviceTable()
	if d, err := table.NewDevice(); err != nil {
		return err
	} else {
		// generate an eui64 from NodeID
		h := md5.New()
		io.WriteString(h, fmt.Sprintf("%s", d.NodeID))
		eui64, _ := iotnet.NewEUI64(fmt.Sprintf("%X", h.Sum(nil)[:8]))
		d.Endpoint.EUI64 = zigbee.EUI64(eui64)
		d.Endpoint.Endpoint = 0x01

		if _, err := d.Update(); err != nil {
			return err
		} else {
			fmt.Printf("added device: %s\n", d)
			return nil
		}
	}
}

// RemoveCommand removes a device from the device table
func RemoveCommand(r *REPL, c *Command) error {
	table := r.Gateway.GetDeviceTable()
	if len(c.Args) != 1 {
		return fmt.Errorf("needs NodeID as argument")
	} else if n, err := zigbee.NewUInt16(c.Args[0]); err != nil {
		return err
	} else {
		id := zigbee.NodeID{n}
		if d, err := table.Get(id); err != nil {
			return err
		} else if err := d.Delete(); err != nil {
			return err
		}
		return nil
	}
}

// ListCommand lists all the devices in the device table
func ListCommand(r *REPL, c *Command) error {
	table := r.Gateway.GetDeviceTable()
	table.Range(func(d *device.Device) bool {
		fmt.Printf("%s: %s\n", d.NodeID, d)
		return true
	})
	return nil
}
