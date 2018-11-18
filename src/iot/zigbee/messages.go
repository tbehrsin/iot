package zigbee

type heartbeatMessage struct {
	NetworkUp bool `json:"networkUp"`
}

type deviceListMessage struct {
	Devices []deviceMessage `json:"devices"`
}

type settingsMessage struct {
	NCPStackVersion string `json:"ncpStackVersion"`
	NetworkUp       bool   `json:"networkUp"`
}

type deviceMessage struct {
	NodeID               nodeID         `json:"nodeId"`
	State                state          `json:"deviceState"`
	Type                 deviceType     `json:"type"`
	TimeSinceLastMessage uint32         `json:"timeSinceLastMessage"`
	Endpoint             deviceEndpoint `json:"deviceEndpoint"`
}

type deviceLeftMessage struct {
	EUI64 EUI64 `json:"eui64"`
}

type deviceStateChangeMessage struct {
	EUI64 EUI64 `json:"eui64"`
	State state `json:"deviceState"`
}

type commandMessage struct {
	Commands []command `json:"commands"`
}

type publishStateMessage struct {
}

type updateSettingsMessage struct {
	MeasureStatistics bool `json:"measureStatistics"`
}

const (
	StateJustJoined              state = 0x00
	StateHaveActive                    = 0x01
	StateHaveEndpointDescription       = 0x02
	StateJoined                        = 0x10
	StateUnresponsive                  = 0x11
	StateLeaveSent                     = 0x20
	StateLeft                          = 0x30
	StateUnknown                       = 0xff
)

func (s state) String() string {
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
