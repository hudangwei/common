package net

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type WSConn struct {
	conn *websocket.Conn
}

func NewWSConn(conn *websocket.Conn) *WSConn {
	return &WSConn{conn}
}

func (c *WSConn) Read(p []byte) (int, error) {
	_, r, err := c.conn.NextReader()
	if err != nil {
		log.Println("websocket read with err:", err)
		return 0, err
	}
	return r.Read(p)
}

func (c *WSConn) Write(p []byte) (int, error) {
	err := c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if err != nil {
		log.Println("websocket write with err:", err)
		return 0, err
	}

	err = c.conn.WriteMessage(websocket.TextMessage, p)
	if err != nil {
		log.Println("websocket write with err:", err)
		return 0, err
	}

	return len(p), nil
}

func (c *WSConn) Close() error {
	if c.conn != nil {
		c.conn.Close()
	}
	return nil
}
