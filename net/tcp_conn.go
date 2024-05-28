package net

import (
	"bufio"
	"io"
	"net"
)

type TCPConn struct {
	conn   net.Conn
	reader *bufio.Reader
}

func NewTCPConn(conn net.Conn) *TCPConn {
	return &TCPConn{
		conn:   conn,
		reader: bufio.NewReader(conn),
	}
}

func (c *TCPConn) Reader() (io.Reader, error) {
	return c.reader, nil
}

func (c *TCPConn) Write(p []byte) (int, error) {
	return c.conn.Write(p)
}

func (c *TCPConn) Close() error {
	if c.conn != nil {
		c.conn.Close()
	}
	return nil
}
