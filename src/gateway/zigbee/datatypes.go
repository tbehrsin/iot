package zigbee

import (
	"encoding/binary"
	"fmt"
)

type DataType struct {
	Name      string
	ID        uint8
	Unmarshal func([]byte) (interface{}, error)
	Marshal   func(interface{}) ([]byte, error)
	Len       func([]byte) int
}

func makeDataTypes(dataTypes []DataType) map[uint8]DataType {
	out := map[uint8]DataType{}
	for _, dataType := range dataTypes {
		out[dataType.ID] = dataType
	}
	return out
}

func makeUintDataType(name string, id uint8, length int) DataType {
	return DataType{
		name,
		id,
		func(in []byte) (interface{}, error) {
			switch length {
			case 1:
				return uint8(in[0]), nil
			case 2:
				return binary.LittleEndian.Uint16(in[0:2]), nil
			case 3:
				fallthrough
			case 4:
				return binary.LittleEndian.Uint32(append(make([]byte, 4-len(in)), in...)), nil
			case 5:
				fallthrough
			case 6:
				fallthrough
			case 7:
				fallthrough
			case 8:
				return binary.LittleEndian.Uint64(append(make([]byte, 8-len(in)), in...)), nil
			}
			return nil, fmt.Errorf("invalid %s of length %d", name, length)
		},
		func(interface{}) ([]byte, error) {
			return nil, nil
		},
		func([]byte) int {
			return length
		},
	}
}

var dataTypes = makeDataTypes([]DataType{
	DataType{
		"bool",
		0x10,
		func(in []byte) (interface{}, error) {
			return in[0] == 0x01, nil
		},
		func(in interface{}) ([]byte, error) {
			if in == true {
				return []byte{0x01}, nil
			} else {
				return []byte{0x00}, nil
			}
		},
		func(in []byte) int {
			return 1
		},
	},
	makeUintDataType("uint8", 0x20, 1),
	makeUintDataType("uint16", 0x21, 2),
	makeUintDataType("uint24", 0x22, 3),
	makeUintDataType("uint32", 0x23, 4),
	makeUintDataType("uint40", 0x24, 5),
	makeUintDataType("uint48", 0x25, 6),
	makeUintDataType("uint56", 0x26, 7),
	makeUintDataType("uint64", 0x27, 8),
	makeUintDataType("enum8", 0x30, 1),
	makeUintDataType("enum16", 0x31, 2),
	DataType{
		"string",
		0x42,
		func(in []byte) (interface{}, error) {
			length := int(uint8(in[0]))
			if length != len(in[1:]) {
				return nil, fmt.Errorf("invalid string length %d for string \"%s\" of length %d", length, in[1:], len(in[1:]))
			}
			return string(in[1:]), nil
		},
		func(in interface{}) ([]byte, error) {
			s := []byte(fmt.Sprintf("%s", in))
			return append([]byte{byte(len(s))}, s...), nil
		},
		func(in []byte) int {
			return 1 + int(uint8(in[0]))
		},
	},
})
