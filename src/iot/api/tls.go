package api

import (
	"net"
	"time"
)

type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (l tcpKeepAliveListener) Accept() (net.Conn, error) {
	t, err := l.AcceptTCP()
	if err != nil {
		return nil, err
	}
	t.SetKeepAlive(true)
	t.SetKeepAlivePeriod(time.Duration(3) * time.Minute)
	return t, nil
}
