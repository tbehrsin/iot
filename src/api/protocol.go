package api

import (
	protobuf "api/protocol"
	"fmt"

	"github.com/golang/protobuf/proto"
)

type protocol struct {
	running      bool
	read         chan []byte
	write        chan []byte
	transactions map[uint32]chan interface{}
	nextID       uint32
}

type clientProtocolInterface interface {
	onReadFileRequest(id uint32, m *protobuf.ReadFileRequest) error
	onIsDirRequest(id uint32, m *protobuf.IsDirRequest) error
	onIsExistRequest(id uint32, m *protobuf.IsExistRequest) error
}

type serverProtocolInterface interface {
	onReadFileResponse(id uint32, m *protobuf.ReadFileResponse) error
	onIsDirResponse(id uint32, m *protobuf.IsDirResponse) error
	onIsExistResponse(id uint32, m *protobuf.IsExistResponse) error
}

func (p *protocol) ReadMessage(i interface{}, message []byte) error {
	var container protobuf.Message
	if err := proto.Unmarshal(message, &container); err != nil {
		return err
	}

	switch m := container.Message.(type) {
	case *protobuf.Message_ReadFileRequest:
		return i.(clientProtocolInterface).onReadFileRequest(container.Id, m.ReadFileRequest)
	case *protobuf.Message_ReadFileResponse:
		return i.(serverProtocolInterface).onReadFileResponse(container.Id, m.ReadFileResponse)
	case *protobuf.Message_IsDirRequest:
		return i.(clientProtocolInterface).onIsDirRequest(container.Id, m.IsDirRequest)
	case *protobuf.Message_IsDirResponse:
		return i.(serverProtocolInterface).onIsDirResponse(container.Id, m.IsDirResponse)
	case *protobuf.Message_IsExistRequest:
		return i.(clientProtocolInterface).onIsExistRequest(container.Id, m.IsExistRequest)
	case *protobuf.Message_IsExistResponse:
		return i.(serverProtocolInterface).onIsExistResponse(container.Id, m.IsExistResponse)
	}

	return nil
}

func (p *protocol) WriteMessage(i interface{}, message *protobuf.Message) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic in write message: %+v", r)
		}
	}()

	if b, err := proto.Marshal(message); err != nil {
		return err
	} else {
		p.write <- b
	}

	return nil
}
