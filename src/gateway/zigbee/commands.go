package zigbee

import (
	"encoding/binary"
	"fmt"
	"log"
	"strings"
)

const (
	ReadAttributes uint8 = iota
	ReadAttributesResponse
	WriteAttributes
	WriteAttributesUndivided
	WriteAttributesResponse
	WriteAttributesNoResponse
	ConfigureReporting
	ConfigureReportingResponse
	ReadReportingConfiguration
	ReadReportingConfigurationResponse
	ReportAttributes
	DefaultResponse
	DiscoverAttributes
	DiscoverAttributesResponse
	ReadAttributesStructured
	WriteAttributesStructured
	WriteAttributesStructuredResponse
	DiscoverCommandsReceived
	DiscoverCommandsReceivedResponse
	DiscoverCommandsGenerated
	DiscoverCommandsGeneratedResponse
	DiscoverAttributesExtended
	DiscoverAttributesExtendedResponse
)

func (g *Gateway) onReadAttributesResponse(eui64 EUI64, message ZCLResponseMessage) {
	if device, ok := g.devices.Load(message.Endpoint.EUI64); ok {
		attributes := make([]string, 0, 4)
		values := make([]interface{}, 0, 4)

		for i := 0; i < len(message.Data); {
			attribute := binary.LittleEndian.Uint16(message.Data[i : i+2])
			attributes = append(attributes, fmt.Sprintf("%d", attribute))
			status := uint8(message.Data[i+2])
			if status == 0x00 {
				dataType := uint8(message.Data[i+3])
				data := message.Data[i+4:]

				if dt, ok := dataTypes[dataType]; !ok {
					log.Printf("unknown data type (attribute 0x%04x status 0x%02x) type 0x%02x length 0x%04x\n", attribute, status, dataType, len(data))
					i += len(message.Data)
				} else if d, e := dt.Unmarshal(data); e != nil {
					log.Printf("failed to unmarshal data %+v for data type 0x%02x: %+v\n", data, dataType, e)
				} else {
					values = append(values, d)
					i += 4 + dt.Len(data)
				}
			} else {
				i += 3
			}
		}

		for _, endpoint := range device.(*Device).FindEndpointsForCluster(Cluster{ClusterType{}, message.ClusterID}) {
			device.(*Device).Emit(fmt.Sprintf("attr:%d:%d:%s", endpoint.Endpoint, message.ClusterID.Value, strings.Join(attributes, ":")), values)
		}
	}
}
