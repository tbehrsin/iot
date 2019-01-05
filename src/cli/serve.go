package main

import (
	"api"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/gorilla/websocket"
)

type FileServer struct {
	protocol *api.ClientProtocol
	write    chan []byte
	conn     *websocket.Conn
	errors   chan error
	Done     chan error
	closed   bool
}

func Serve(dir string, conn *websocket.Conn) (fs *FileServer) {
	errors := make(chan error, 1)
	done := make(chan error, 1)
	write := make(chan []byte)
	read := make(chan []byte)

	fs = &FileServer{
		write:  write,
		conn:   conn,
		errors: errors,
		Done:   done,
	}

	p := api.NewClientProtocol(fs, read, write)
	fs.protocol = p

	go func() {
		err := <-fs.errors
		if fs.closed {
			fs.Done <- nil
		} else {
			fs.Done <- err
		}
		defer close(fs.Done)
		defer close(errors)
		defer conn.Close()
		defer close(read)
		defer close(write)
	}()

	go func() {
		defer func() {
			if err := recover(); err != nil {
				select {
				case errors <- fmt.Errorf("%s", err):
				default:
				}
			}
		}()

		for {
			message, more := <-write

			if more {
				conn.WriteMessage(websocket.BinaryMessage, message)
			} else {
				break
			}
		}
	}()

	go func() {
		if err := p.Run(); err != nil {
			select {
			case errors <- err:
			default:
			}
		}
	}()

	go func() {
		for {
			if mt, message, err := conn.ReadMessage(); err != nil {
				select {
				case errors <- err:
				default:
				}
				break
			} else if mt == websocket.BinaryMessage {
				read <- message
			} else if mt == websocket.CloseMessage {
				break
			}
		}
	}()

	return fs
}

func (fs *FileServer) Wait() error {
	err := <-fs.Done
	return err
}

func (fs *FileServer) Close() {
	fs.closed = true
	fs.conn.Close()
}

func (fs *FileServer) ReadFile(path string, writer io.Writer) error {
	if b, err := ioutil.ReadFile(strings.TrimPrefix(path, "/")); err != nil {
		return err
	} else {
		writer.Write(b)
		return nil
	}
}

func (fs *FileServer) IsDir(path string) bool {
	if fi, err := os.Lstat(strings.TrimPrefix(path, "/")); err == nil && fi.Mode().IsDir() {
		return true
	}
	return false
}

func (fs *FileServer) IsExist(path string) bool {
	if _, err := os.Stat(strings.TrimPrefix(path, "/")); os.IsExist(err) {
		return true
	}
	return false
}
