package api

import (
	protobuf "api/protocol"
	"bytes"
	"fmt"
	"io"
)

type ClientProtocol struct {
	protocol
	client ClientInterface
}

type ClientInterface interface {
	ProtocolError(err error)
	ReadFile(path string, writer io.Writer) error
	IsDir(path string) bool
	IsExist(path string) bool
}

func (p *ClientProtocol) Run(client ClientInterface, read chan []byte, write chan []byte) error {
	if p.running {
		return fmt.Errorf("client protocol already running")
	}
	p.running = true
	p.write = write
	p.client = client

	go func() {
		for {
			message, more := <-read

			if more {
				p.ReadMessage(p, message)
			} else {
				break
			}
		}
	}()

	return nil
}

func (p *ClientProtocol) onProtocolError(err error) {
	p.client.ProtocolError(err)
}

func (p *ClientProtocol) onReadFileRequest(id uint32, m *protobuf.ReadFileRequest) {
	var buf bytes.Buffer
	if err := p.client.ReadFile(m.Path, &buf); err != nil {
		p.WriteMessage(p, &protobuf.Message{
			Id: id,
			Message: &protobuf.Message_ReadFileResponse{
				ReadFileResponse: &protobuf.ReadFileResponse{
					Error: err.Error(),
				},
			},
		})
	} else {
		p.WriteMessage(p, &protobuf.Message{
			Id: id,
			Message: &protobuf.Message_ReadFileResponse{
				ReadFileResponse: &protobuf.ReadFileResponse{
					File: buf.Bytes(),
				},
			},
		})
	}
}

func (p *ClientProtocol) onIsDirRequest(id uint32, m *protobuf.IsDirRequest) {
	p.WriteMessage(p, &protobuf.Message{
		Id: id,
		Message: &protobuf.Message_IsDirResponse{
			IsDirResponse: &protobuf.IsDirResponse{
				Value: p.client.IsDir(m.Path),
			},
		},
	})
}

func (p *ClientProtocol) onIsExistRequest(id uint32, m *protobuf.IsExistRequest) {
	p.WriteMessage(p, &protobuf.Message{
		Id: id,
		Message: &protobuf.Message_IsExistResponse{
			IsExistResponse: &protobuf.IsExistResponse{
				Value: p.client.IsExist(m.Path),
			},
		},
	})
}
