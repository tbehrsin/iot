package api

import (
	protobuf "api/protocol"
	"fmt"

	"github.com/golang/protobuf/proto"
)

type protocol struct {
	running bool
	write   chan []byte
	nextID  uint32
}

type clientProtocolInterface interface {
	onReadFileRequest(id uint32, m *protobuf.ReadFileRequest)
	onIsDirRequest(id uint32, m *protobuf.IsDirRequest)
	onIsExistRequest(id uint32, m *protobuf.IsExistRequest)
}

type serverProtocolInterface interface {
	onReadFileResponse(id uint32, m *protobuf.ReadFileResponse)
	onIsDirResponse(id uint32, m *protobuf.IsDirResponse)
	onIsExistResponse(id uint32, m *protobuf.IsExistResponse)
}

type errorProtocolInterface interface {
	onProtocolError(err error)
}

func (p *protocol) ReadMessage(i interface{}, message []byte) {
	var container protobuf.Message
	if err := proto.Unmarshal(message, &container); err != nil {
		go i.(errorProtocolInterface).onProtocolError(err)
		return
	}

	switch m := container.Message.(type) {
	case *protobuf.Message_ReadFileRequest:
		go i.(clientProtocolInterface).onReadFileRequest(container.Id, m.ReadFileRequest)
	case *protobuf.Message_ReadFileResponse:
		go i.(serverProtocolInterface).onReadFileResponse(container.Id, m.ReadFileResponse)
	case *protobuf.Message_IsDirRequest:
		go i.(clientProtocolInterface).onIsDirRequest(container.Id, m.IsDirRequest)
	case *protobuf.Message_IsDirResponse:
		go i.(serverProtocolInterface).onIsDirResponse(container.Id, m.IsDirResponse)
	case *protobuf.Message_IsExistRequest:
		go i.(clientProtocolInterface).onIsExistRequest(container.Id, m.IsExistRequest)
	case *protobuf.Message_IsExistResponse:
		go i.(serverProtocolInterface).onIsExistResponse(container.Id, m.IsExistResponse)
	default:
		go i.(errorProtocolInterface).onProtocolError(fmt.Errorf("unknown message: %+v", m))
	}
}

func (p *protocol) WriteMessage(i interface{}, message *protobuf.Message) {
	if b, err := proto.Marshal(message); err != nil {
		go i.(errorProtocolInterface).onProtocolError(err)
	} else {
		p.write <- b
	}
}
