package net

import (
	"encoding/json"
	"io"
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

func (c *WSConn) ReadJSON(v interface{}) error {
	_, r, err := c.conn.NextReader()
	if err != nil {
		log.Println("websocket read with err:", err)
		return err
	}
	decoder := json.NewDecoder(r)
	decoder.UseNumber()
	err = decoder.Decode(v)
	if err == io.EOF {
		// One value is expected in the message.
		err = io.ErrUnexpectedEOF
	}
	if err != nil {
		log.Println("json unmarshal with err:", err)
	}
	return err
}

func (c *WSConn) WriteJSON(v string) error {
	err := c.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	if err != nil {
		log.Println("websocket write with err:", err)
		return err
	}

	err = c.conn.WriteMessage(websocket.TextMessage, []byte(v))
	if err != nil {
		log.Println("websocket write with err:", err)
		return err
	}

	return nil
}

func (c *WSConn) Close() error {
	if c.conn != nil {
		c.conn.Close()
	}
	return nil
}
