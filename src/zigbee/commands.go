package main

import (
	"fmt"
	"gateway/zigbee"
	"zigbee/device"
)

// AddCommand adds a device to the device table
func AddCommand(r *REPL, c *Command) error {
	table := r.Gateway.GetDeviceTable()
	d := table.NewDevice()
	// TODO fill in device struct
	if err := d.Update(); err != nil {
		return err
	} else {
		fmt.Printf("added device: %s\n", d.NodeID)
		return nil
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
		fmt.Printf("%s: device\n", d.NodeID)
		return true
	})
	return nil
}
