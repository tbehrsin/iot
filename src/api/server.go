package api

import (
	protobuf "api/protocol"
	"fmt"
)

type ServerProtocol struct {
	protocol
	server ServerInterface
}

type ServerInterface interface {
}

func NewServerProtocol(server ServerInterface, read chan []byte, write chan []byte) *ServerProtocol {
	p := &ServerProtocol{}
	p.running = false
	p.read = read
	p.write = write
	p.server = server
	p.transactions = make(map[uint32]chan interface{})
	return p
}

func (p *ServerProtocol) Run() error {
	if p.running {
		return fmt.Errorf("server protocol already running")
	}
	p.running = true

	defer func() {
		for _, ch := range p.transactions {
			close(ch)
		}
	}()

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

func (p *ServerProtocol) ReadFile(path string) ([]byte, error) {
	id := p.nextID
	p.nextID++
	ch := make(chan interface{})
	p.transactions[id] = ch
	defer delete(p.transactions, id)

	if err := p.WriteMessage(p, &protobuf.Message{
		Id: id,
		Message: &protobuf.Message_ReadFileRequest{
			ReadFileRequest: &protobuf.ReadFileRequest{
				Path: path,
			},
		},
	}); err != nil {
		return nil, err
	}

	file := (<-ch).(*protobuf.ReadFileResponse)

	if file.Error != "" {
		return nil, fmt.Errorf(file.Error)
	}
	return file.File, nil
}

func (p *ServerProtocol) onReadFileResponse(id uint32, m *protobuf.ReadFileResponse) error {
	if ch, ok := p.transactions[id]; ok {
		ch <- m
		close(ch)
	}
	return nil
}

func (p *ServerProtocol) IsDir(path string) bool {
	id := p.nextID
	p.nextID++
	ch := make(chan interface{})
	p.transactions[id] = ch
	defer delete(p.transactions, id)

	if err := p.WriteMessage(p, &protobuf.Message{
		Id: id,
		Message: &protobuf.Message_IsDirRequest{
			IsDirRequest: &protobuf.IsDirRequest{
				Path: path,
			},
		},
	}); err != nil {
		return false
	}

	result := (<-ch).(*protobuf.IsDirResponse)

	return result.Value
}

func (p *ServerProtocol) onIsDirResponse(id uint32, m *protobuf.IsDirResponse) error {
	if ch, ok := p.transactions[id]; ok {
		ch <- m
		close(ch)
	}
	return nil
}

func (p *ServerProtocol) IsExist(path string) bool {
	id := p.nextID
	p.nextID++
	ch := make(chan interface{})
	p.transactions[id] = ch
	defer delete(p.transactions, id)

	if err := p.WriteMessage(p, &protobuf.Message{
		Id: id,
		Message: &protobuf.Message_IsExistRequest{
			IsExistRequest: &protobuf.IsExistRequest{
				Path: path,
			},
		},
	}); err != nil {
		return false
	}

	result := (<-ch).(*protobuf.IsExistResponse)

	return result.Value
}

func (p *ServerProtocol) onIsExistResponse(id uint32, m *protobuf.IsExistResponse) error {
	if ch, ok := p.transactions[id]; ok {
		ch <- m
		close(ch)
	}
	return nil
}
