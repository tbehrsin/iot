package api

import (
	"crypto/tls"
	"log"
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

type TLSListener struct {
	inner  net.Listener
	config *tls.Config
}

func (l *TLSListener) Accept() (net.Conn, error) {
	if conn, err := l.inner.Accept(); err != nil {
		return nil, err
	} else {
		tlsConn := tls.Server(conn, l.config)
		if err := tlsConn.Handshake(); err != nil {
			return nil, err
		}
		connWrap := &TLSConnection{tlsConn, conn}
		connWrap.SetDeadline(time.Time{})
		return connWrap, nil
	}
}

func (l *TLSListener) Addr() net.Addr {
	return l.inner.Addr()
}

func (l *TLSListener) Close() error {
	return l.inner.Close()
}

type TLSConnection struct {
	tlsConn *tls.Conn
	conn    net.Conn
}

func (c *TLSConnection) Read(b []byte) (n int, err error) {
	return c.tlsConn.Read(b)
}

func (c *TLSConnection) Write(b []byte) (n int, err error) {
	return c.tlsConn.Write(b)
}

func (c *TLSConnection) Close() error {
	return c.tlsConn.Close()
}

func (c *TLSConnection) LocalAddr() net.Addr {
	return c.tlsConn.LocalAddr()
}

func (c *TLSConnection) RemoteAddr() net.Addr {
	return c.tlsConn.RemoteAddr()
}

func (c *TLSConnection) SetDeadline(t time.Time) error {
	if err := c.tlsConn.SetDeadline(t); err != nil {
		return err
	}
	return c.conn.SetDeadline(t)
}

func (c *TLSConnection) SetReadDeadline(t time.Time) error {
	if err := c.tlsConn.SetReadDeadline(t); err != nil {
		return err
	}
	return c.conn.SetReadDeadline(t)
}

func (c *TLSConnection) SetWriteDeadline(t time.Time) error {
	if err := c.tlsConn.SetWriteDeadline(t); err != nil {
		return err
	}
	return c.conn.SetWriteDeadline(t)
}

// https://www.stavros.io/posts/proxying-two-connections-go/
func chanFromConn(conn net.Conn) chan []byte {
	c := make(chan []byte)

	go func() {
		b := make([]byte, 1024)

		for {
			n, err := conn.Read(b)
			if n > 0 {
				res := make([]byte, n)
				// Copy the buffer so it doesn't get changed while read by the recipient.
				copy(res, b[:n])
				c <- res
			}
			if err != nil {
				c <- nil
				break
			}
		}
	}()

	return c
}

// https://www.stavros.io/posts/proxying-two-connections-go/
func pipe(conn1 net.Conn, conn2 net.Conn) {
	chan1 := chanFromConn(conn1)
	chan2 := chanFromConn(conn2)
	defer log.Println("closing pipe")
	for {
		select {
		case b1 := <-chan1:
			if b1 == nil {
				return
			} else {
				conn2.Write(b1)
			}
		case b2 := <-chan2:
			if b2 == nil {
				return
			} else {
				conn1.Write(b2)
			}
		}
	}
}
