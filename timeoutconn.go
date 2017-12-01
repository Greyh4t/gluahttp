package gluahttp

import (
	"net"
	"time"
)

func NewTimeoutConn(netConn net.Conn, timeout time.Duration) net.Conn {
	return &TimeoutConn{netConn, timeout}
}

type TimeoutConn struct {
	net.Conn
	timeout time.Duration
}

func (c *TimeoutConn) Read(b []byte) (int, error) {
	if c.timeout > 0 {
		err := c.Conn.SetReadDeadline(time.Now().Add(c.timeout))
		if err != nil {
			return 0, err
		}
	}
	return c.Conn.Read(b)
}

func (c *TimeoutConn) Write(b []byte) (int, error) {
	if c.timeout > 0 {
		err := c.Conn.SetWriteDeadline(time.Now().Add(c.timeout))
		if err != nil {
			return 0, err
		}
	}
	return c.Conn.Write(b)
}
