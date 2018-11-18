package zigbee

import (
	"encoding/json"
	"fmt"
	"iot/net"
	"strconv"
	"strings"
)

type state uint8

func (s state) MarshalV8() interface{} {
	return s.String()
}

type command struct {
	Command string `json:"command", v8:"command"`
	Delay   uint32 `json:"postDelayMs", v8:"delay"`
}

type EUI64 net.EUI64

func (e *EUI64) MarshalJSON() ([]byte, error) {
	s := fmt.Sprintf("%s", net.EUI64(*e).String())
	return json.Marshal(s)
}

func (e *EUI64) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	if n, err := net.NewEUI64(strings.TrimPrefix(s, "0x")); err != nil {
		return err
	} else {
		ne := EUI64(n)
		copy(e[:], ne[:])
		return nil
	}
}

type _uint16 struct {
	value uint16
}

func (n _uint16) String() string {
	return fmt.Sprintf("%04X", n.value)
}

func (n *_uint16) MarshalJSON() ([]byte, error) {
	s := fmt.Sprintf("0x%s", n.String())
	return json.Marshal(s)
}

func (n *_uint16) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	if v, err := strconv.ParseUint(strings.TrimPrefix(s, "0x"), 16, 16); err != nil {
		return err
	} else {
		n.value = uint16(v)
		return nil
	}
}

func (n _uint16) MarshalV8() interface{} {
	return n.value
}

type _uint8 struct {
	value uint8
}

func (n _uint8) String() string {
	return fmt.Sprintf("%02X", n.value)
}

func (n *_uint8) MarshalJSON() ([]byte, error) {
	s := fmt.Sprintf("0x%s", n.String())
	return json.Marshal(s)
}

func (n *_uint8) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	if v, err := strconv.ParseUint(strings.TrimPrefix(s, "0x"), 16, 8); err != nil {
		return err
	} else {
		n.value = uint8(v)
		return nil
	}
}

func (n _uint8) MarshalV8() interface{} {
	return n.value
}

type nodeID struct {
	_uint16
}

type clusterID struct {
	_uint16
}

type deviceType struct {
	_uint16
}

const (
	clusterTypeIn  string = "In"
	clusterTypeOut        = "Out"
)

type clusterType struct {
	out bool
}

func (c clusterType) String() string {
	if c.out {
		return "out"
	} else {
		return "in"
	}
}

func (c *clusterType) MarshalJSON() ([]byte, error) {
	if c.out {
		return json.Marshal(clusterTypeOut)
	} else {
		return json.Marshal(clusterTypeIn)
	}
}

func (c *clusterType) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	c.out = s == clusterTypeOut
	return nil
}

func (c clusterType) MarshalV8() interface{} {
	return c.String()
}

type endpoint uint8

func (e endpoint) String() string {
	return fmt.Sprintf("%02X", uint8(e))
}

type deviceEndpointInfo struct {
	EUI64    EUI64    `json:"eui64" v8:"eui64"`
	Endpoint endpoint `json:"endpoint" v8:"endpoint"`
}

type deviceEndpoint struct {
	deviceEndpointInfo
	Clusters []cluster `json:"clusterInfo" v8:"clusters"`
}

func (e *deviceEndpoint) EndpointInfo() *deviceEndpointInfo {
	return &deviceEndpointInfo{e.EUI64, e.Endpoint}
}

type cluster struct {
	Type clusterType `json:"clusterType" v8:"type"`
	ID   clusterID   `json:"clusterId" v8:"id"`
}
