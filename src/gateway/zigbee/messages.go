package zigbee

type HeartbeatMessage struct {
	NetworkUp bool `json:"networkUp"`
}

type DeviceListMessage struct {
	Devices []DeviceMessage `json:"devices"`
}

type RelayListMessage struct {
}

type SettingsMessage struct {
	NCPStackVersion string `json:"ncpStackVersion"`
	NetworkUp       bool   `json:"networkUp"`
}

type DeviceMessage struct {
	NodeID               NodeID         `json:"nodeId"`
	State                State          `json:"deviceState"`
	Type                 DeviceType     `json:"type"`
	TimeSinceLastMessage uint32         `json:"timeSinceLastMessage"`
	Endpoint             DeviceEndpoint `json:"deviceEndpoint"`
}

type DeviceLeftMessage struct {
	EUI64 EUI64 `json:"eui64"`
}

type DeviceStateChangeMessage struct {
	EUI64 EUI64 `json:"eui64"`
	State State `json:"deviceState"`
}

type OTAEventMessage struct {
}

type ExecutedMessage struct {
	Command string `json:"command"`
}

type ZCLResponseMessage struct {
	ClusterID       ClusterID          `json:"clusterId"`
	CommandID       CommandID          `json:"commandId"`
	Data            CommandData        `json:"commandData"`
	ClusterSpecific bool               `json:"clusterSpecific"`
	Endpoint        DeviceEndpointInfo `json:"deviceEndpoint"`
}

type ZDOResponseMessage struct {
}

type APSResponseMessage struct {
}

type CommandListMessage struct {
	Commands []Command `json:"commands"`
}

type PublishStateMessage struct {
}

type UpdateSettingsMessage struct {
	MeasureStatistics bool `json:"measureStatistics"`
}

const (
	StateJustJoined              State = 0x00
	StateHaveActive                    = 0x01
	StateHaveEndpointDescription       = 0x02
	StateJoined                        = 0x10
	StateUnresponsive                  = 0x11
	StateLeaveSent                     = 0x20
	StateLeft                          = 0x30
	StateUnknown                       = 0xff
)

func (s State) String() string {
	switch s {
	case StateJustJoined:
		return "just-joined"
	case StateHaveActive:
		return "have-active"
	case StateHaveEndpointDescription:
		return "have-endpoint-description"
	case StateJoined:
		return "joined"
	case StateUnresponsive:
		return "unresponsive"
	case StateLeaveSent:
		return "leave-sent"
	case StateLeft:
		return "left"
	case StateUnknown:
		return "unknown"
	default:
		return "unknown"
	}
}
