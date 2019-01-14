package zigbee

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"gateway/net"
	"strconv"
	"strings"

	"github.com/behrsin/go-v8"
)

type State uint8

func (s State) MarshalV8() interface{} {
	return s.String()
}

type Command struct {
	Command string `json:"commandcli" v8:"command"`
	Delay   uint32 `json:"postDelayMs" v8:"delay"`
}

type EUI64 net.EUI64

func (e EUI64) String() string {
	ne := net.EUI64(e)
	return fmt.Sprintf("0x%s", ne.String())
}

func (e EUI64) bracketString() string {
	ne := net.EUI64(e)
	return fmt.Sprintf("{%s}", ne.String())
}

func (e EUI64) MarshalV8() interface{} {
	ne := net.EUI64(e)
	return ne.String()
}

func (e *EUI64) MarshalJSON() ([]byte, error) {
	s := e.String()
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

type UInt16 struct {
	Value uint16
}

func NewUInt16(s string) (UInt16, error) {
	if v, err := strconv.ParseUint(strings.TrimPrefix(s, "0x"), 16, 16); err != nil {
		return UInt16{0}, err
	} else {
		return UInt16{uint16(v)}, nil
	}
}

func (n UInt16) String() string {
	return fmt.Sprintf("0x%04X", n.Value)
}

func (n *UInt16) MarshalJSON() ([]byte, error) {
	s := n.String()
	return json.Marshal(s)
}

func (n *UInt16) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	if v, err := strconv.ParseUint(strings.TrimPrefix(s, "0x"), 16, 16); err != nil {
		return err
	} else {
		n.Value = uint16(v)
		return nil
	}
}

func (n UInt16) MarshalV8() interface{} {
	return n.Value
}

func V8Uint16(v *v8.Value) UInt16 {
	v64, _ := v.Float64()
	return UInt16{uint16(v64)}
}

type UInt8 struct {
	Value uint8
}

func NewUInt8(s string) (UInt8, error) {
	if v, err := strconv.ParseUint(strings.TrimPrefix(s, "0x"), 16, 8); err != nil {
		return UInt8{0}, err
	} else {
		return UInt8{uint8(v)}, nil
	}
}

func (n UInt8) String() string {
	return fmt.Sprintf("0x%02X", n.Value)
}

func (n *UInt8) MarshalJSON() ([]byte, error) {
	s := n.String()
	return json.Marshal(s)
}

func (n *UInt8) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	if v, err := strconv.ParseUint(strings.TrimPrefix(s, "0x"), 16, 8); err != nil {
		return err
	} else {
		n.Value = uint8(v)
		return nil
	}
}

func (n UInt8) MarshalV8() interface{} {
	return n.Value
}

func V8Uint8(v *v8.Value) UInt8 {
	v64, _ := v.Float64()
	return UInt8{uint8(v64)}
}

type NodeID struct {
	UInt16
}

type ClusterID struct {
	UInt16
}

type DeviceType struct {
	UInt16
}

const (
	ClusterTypeIn  string = "In"
	ClusterTypeOut        = "Out"
)

type ClusterType struct {
	Out bool
}

func (c ClusterType) String() string {
	if c.Out {
		return "out"
	} else {
		return "in"
	}
}

func (c *ClusterType) MarshalJSON() ([]byte, error) {
	if c.Out {
		return json.Marshal(ClusterTypeOut)
	} else {
		return json.Marshal(ClusterTypeIn)
	}
}

func (c *ClusterType) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	c.Out = s == ClusterTypeOut
	return nil
}

func V8ClusterType(clusterType *v8.Value) ClusterType {
	c := ClusterType{
		Out: clusterType.String() == "out",
	}
	return c
}

func (c ClusterType) MarshalV8() interface{} {
	return c.String()
}

type Endpoint uint8

func (e Endpoint) String() string {
	return fmt.Sprintf("0x%02X", int(e))
}

type DeviceEndpointInfo struct {
	EUI64    EUI64    `json:"eui64" v8:"eui64"`
	Endpoint Endpoint `json:"endpoint" v8:"id"`
}

type DeviceEndpoint struct {
	DeviceEndpointInfo
	Clusters []Cluster `json:"clusterInfo" v8:"clusters"`
}

func (d *DeviceEndpoint) Match(cluster Cluster) bool {
	for _, c := range d.Clusters {
		if c.Type == cluster.Type && c.ID == cluster.ID {
			return true
		}
	}
	return false
}

func (d *DeviceEndpoint) MatchAll(clusters []Cluster) bool {
	for _, cluster := range clusters {
		if !d.Match(cluster) {
			return false
		}
	}
	return true
}

func (d *DeviceEndpoint) V8MatchAll(vclusters *v8.Value) (bool, error) {
	if vlength, err := vclusters.Get("length"); err != nil {
		return false, err
	} else if length, err := vlength.Int64(); err != nil && length > 0 {
		clusters := make([]Cluster, length)

		for i := int64(0); i < length; i++ {
			var clusterId ClusterID
			var clusterType ClusterType

			if vcluster, err := vclusters.GetIndex(int(i)); err != nil {
				return false, err
			} else if vclusterId, err := vcluster.Get("id"); err != nil {
				return false, err
			} else if vclusterType, err := vcluster.Get("type"); err != nil {
				return false, err
			} else {
				cluster64, _ := vclusterId.Int64()
				clusterId = ClusterID{UInt16{uint16(cluster64)}}
				clusterType = V8ClusterType(vclusterType)

				clusters[i] = Cluster{clusterType, clusterId}
			}
		}

		return d.MatchAll(clusters), nil
	}

	return true, nil
}

func (e *DeviceEndpoint) EndpointInfo() *DeviceEndpointInfo {
	return &DeviceEndpointInfo{e.EUI64, e.Endpoint}
}

type Cluster struct {
	Type ClusterType `json:"clusterType" v8:"type"`
	ID   ClusterID   `json:"clusterId" v8:"id"`
}

type AttributeID struct {
	UInt16
}

type CommandID struct {
	UInt8
}

type CommandData []byte

func (c CommandData) String() string {
	return fmt.Sprintf("0x%s", strings.ToUpper(hex.EncodeToString(c[:])))
}

func (c CommandData) bracketString() string {
	return fmt.Sprintf("{%s}", strings.ToUpper(hex.EncodeToString(c[:])))
}

func (c *CommandData) MarshalJSON() ([]byte, error) {
	s := c.String()
	return json.Marshal(s)
}

func (c *CommandData) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	} else if b, err := hex.DecodeString(strings.TrimPrefix(s, "0x")); err != nil {
		return err
	} else {
		cb := CommandData(b)
		*c = append(*c, cb...)
		return nil
	}
}
