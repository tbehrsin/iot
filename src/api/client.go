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
	ReadFile(path string, writer io.Writer) error
	IsDir(path string) bool
	IsExist(path string) bool
}

func NewClientProtocol(client ClientInterface, read chan []byte, write chan []byte) *ClientProtocol {
	p := &ClientProtocol{}
	p.running = false
	p.read = read
	p.write = write
	p.client = client
	p.transactions = make(map[uint32]chan interface{})
	return p
}

func (p *ClientProtocol) Run() error {
	if p.running {
		return fmt.Errorf("client protocol already running")
	}

	for {
		message, more := <-p.read

		if more {
			if err := p.ReadMessage(p, message); err != nil {
				return err
			}
		} else {
			break
		}
	}

	return nil
}

func (p *ClientProtocol) onReadFileRequest(id uint32, m *protobuf.ReadFileRequest) error {
	var buf bytes.Buffer
	if err := p.client.ReadFile(m.Path, &buf); err != nil {
		return p.WriteMessage(p, &protobuf.Message{
			Id: id,
			Message: &protobuf.Message_ReadFileResponse{
				ReadFileResponse: &protobuf.ReadFileResponse{
					Error: err.Error(),
				},
			},
		})
	} else {
		return p.WriteMessage(p, &protobuf.Message{
			Id: id,
			Message: &protobuf.Message_ReadFileResponse{
				ReadFileResponse: &protobuf.ReadFileResponse{
					File: buf.Bytes(),
				},
			},
		})
	}
}

func (p *ClientProtocol) onIsDirRequest(id uint32, m *protobuf.IsDirRequest) error {
	return p.WriteMessage(p, &protobuf.Message{
		Id: id,
		Message: &protobuf.Message_IsDirResponse{
			IsDirResponse: &protobuf.IsDirResponse{
				Value: p.client.IsDir(m.Path),
			},
		},
	})
}

func (p *ClientProtocol) onIsExistRequest(id uint32, m *protobuf.IsExistRequest) error {
	return p.WriteMessage(p, &protobuf.Message{
		Id: id,
		Message: &protobuf.Message_IsExistResponse{
			IsExistResponse: &protobuf.IsExistResponse{
				Value: p.client.IsExist(m.Path),
			},
		},
	})
}
