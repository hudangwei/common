package net

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
)

const (
	PACKET_HEADER_SIZE   = 6
	PACKET_MAX_RECV_SIZE = 4 * 1024 //4kb
)

var (
	ErrBadConn      = errors.New("connection was bad")
	ErrHeaderLength = errors.New("invalid length")
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

func (c *TCPConn) ReadJSON(v interface{}) error {
	var header [PACKET_HEADER_SIZE]byte
	if n, err := io.ReadFull(c.reader, header[:]); err != nil || n != PACKET_HEADER_SIZE {
		return ErrBadConn
	}
	packetSize := int(binary.LittleEndian.Uint16(header[:2]))
	if packetSize < PACKET_HEADER_SIZE || packetSize > PACKET_MAX_RECV_SIZE {
		return ErrHeaderLength
	}
	dataLen := packetSize - PACKET_HEADER_SIZE
	data := make([]byte, dataLen)
	if n, err := io.ReadFull(c.reader, data[:]); err != nil || n != dataLen {
		return ErrBadConn
	}
	if err := json.Unmarshal(data, v); err != nil {
		log.Println("json unmarshal with err:", err)
		return err
	}
	return nil
}

func (c *TCPConn) WriteJSON(v string) error {
	packet := make([]byte, PACKET_HEADER_SIZE)
	binary.LittleEndian.PutUint16(packet, uint16(len(v)+PACKET_HEADER_SIZE)) //packetLen
	packet = append(packet, []byte(v)...)
	_, err := c.conn.Write(packet)
	return err
}

func (c *TCPConn) Close() error {
	if c.conn != nil {
		c.conn.Close()
	}
	return nil
}
