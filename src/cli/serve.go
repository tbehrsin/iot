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
}

func Serve(dir string, conn *websocket.Conn) {
	p := &api.ClientProtocol{}

	write := make(chan []byte)
	read := make(chan []byte)

	fs := &FileServer{
		p,
		write,
		conn,
	}

	defer conn.Close()
	defer close(read)
	defer close(write)

	go func() {
		for {
			message, more := <-write

			if more {
				conn.WriteMessage(websocket.BinaryMessage, message)
			} else {
				break
			}
		}
	}()

	p.Run(fs, read, write)

	for {
		if mt, message, err := conn.ReadMessage(); err != nil {
			return
		} else if mt == websocket.BinaryMessage {
			read <- message
		} else if mt == websocket.CloseMessage {
			return
		}
	}
}

func (fs *FileServer) ProtocolError(err error) {
	fs.conn.Close()
	fmt.Println("protocol error:", err)
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
