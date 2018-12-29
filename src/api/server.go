package api

import (
	protobuf "api/protocol"
	"fmt"
)

type ServerProtocol struct {
	protocol
	server       ServerInterface
	transactions map[uint32]chan interface{}
}

type ServerInterface interface {
	ProtocolError(err error)
}

func (p *ServerProtocol) Run(server ServerInterface, read chan []byte, write chan []byte) error {
	if p.running {
		return fmt.Errorf("server protocol already running")
	}
	p.running = true
	p.write = write
	p.server = server
	p.transactions = make(map[uint32]chan interface{})

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

func (p *ServerProtocol) onProtocolError(err error) {
	p.server.ProtocolError(err)
}

func (p *ServerProtocol) ReadFile(path string) ([]byte, error) {
	id := p.nextID
	p.nextID++
	ch := make(chan interface{})
	p.transactions[id] = ch

	p.WriteMessage(p, &protobuf.Message{
		Id: id,
		Message: &protobuf.Message_ReadFileRequest{
			ReadFileRequest: &protobuf.ReadFileRequest{
				Path: path,
			},
		},
	})

	file := (<-ch).(*protobuf.ReadFileResponse)
	delete(p.transactions, id)
	close(ch)

	if file.Error != "" {
		return nil, fmt.Errorf(file.Error)
	}
	return file.File, nil
}

func (p *ServerProtocol) onReadFileResponse(id uint32, m *protobuf.ReadFileResponse) {
	if ch, ok := p.transactions[id]; ok {
		ch <- m
	}
}

func (p *ServerProtocol) IsDir(path string) bool {
	id := p.nextID
	p.nextID++
	ch := make(chan interface{})
	p.transactions[id] = ch

	p.WriteMessage(p, &protobuf.Message{
		Id: id,
		Message: &protobuf.Message_IsDirRequest{
			IsDirRequest: &protobuf.IsDirRequest{
				Path: path,
			},
		},
	})

	result := (<-ch).(*protobuf.IsDirResponse)
	delete(p.transactions, id)
	close(ch)

	return result.Value
}

func (p *ServerProtocol) onIsDirResponse(id uint32, m *protobuf.IsDirResponse) {
	if ch, ok := p.transactions[id]; ok {
		ch <- m
	}
}

func (p *ServerProtocol) IsExist(path string) bool {
	id := p.nextID
	p.nextID++
	ch := make(chan interface{})
	p.transactions[id] = ch

	p.WriteMessage(p, &protobuf.Message{
		Id: id,
		Message: &protobuf.Message_IsExistRequest{
			IsExistRequest: &protobuf.IsExistRequest{
				Path: path,
			},
		},
	})

	result := (<-ch).(*protobuf.IsExistResponse)
	delete(p.transactions, id)
	close(ch)

	return result.Value
}

func (p *ServerProtocol) onIsExistResponse(id uint32, m *protobuf.IsExistResponse) {
	if ch, ok := p.transactions[id]; ok {
		ch <- m
	}
}
