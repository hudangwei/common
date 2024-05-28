package net

import (
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hudangwei/common/logger"
	"go.uber.org/zap"
)

var (
	WSUpgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		EnableCompression: true,
	}
)

type WebsocketServer struct {
	manager  *Manager
	newCodec NewCodecFunc
	protocol Protocol
}

func NewWebsocketServer(manager *Manager, newCodec NewCodecFunc, protocol Protocol) *WebsocketServer {
	return &WebsocketServer{
		manager:  manager,
		newCodec: newCodec,
		protocol: protocol,
	}
}

func (s *WebsocketServer) Start(addr string) {
	mu := http.NewServeMux()
	mu.HandleFunc("/", s.wsHandler)
	go http.ListenAndServe(addr, mu)
}

func (s *WebsocketServer) wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := WSUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	session := s.manager.NewSession(s.protocol)
	session.PutExtraData("remote_addr", conn.RemoteAddr().String())
	ua := r.Header.Get("User-Agent")
	session.PutExtraData("ua", ua)
	codec := s.newCodec(NewWSConn(conn))
	session.Do(codec)
	return
}

func NewWSSession(addr string, newCodec NewCodecFunc, protocol Protocol) *Session {
	// u := &url.URL{Scheme: scheme, Host: addr, Path: "/"}
	c, _, err := websocket.DefaultDialer.Dial(addr, nil)
	if err != nil {
		logger.Warn("websocket dial", zap.Error(err))
		return nil
	}
	session := NewSession(protocol, nil)
	session.Do(newCodec(NewWSConn(c)))
	return session
}

type ReConnectWSClient struct {
	Protocol
	addr     string
	newCodec NewCodecFunc
	dial     func(addr string) (*websocket.Conn, error)

	lock              sync.RWMutex
	session           *Session
	closeChan         chan struct{}
	keepaliveOnce     sync.Once
	keepaliveInterval int
	keepaliveReq      func() []byte
}

func NewReConnectWSClient(addr string, newCodec NewCodecFunc, protocol Protocol, keepaliveInterval int, keepaliveReq func() []byte) (*ReConnectWSClient, error) {
	client := &ReConnectWSClient{
		Protocol: protocol,
		addr:     addr,
		newCodec: newCodec,
		dial: func(addr string) (*websocket.Conn, error) {
			c, _, err := websocket.DefaultDialer.Dial(addr, nil)
			if err != nil {
				logger.Warn("websocket dial", zap.Error(err))
				return nil, err
			}
			return c, nil
		},
		closeChan:         make(chan struct{}),
		keepaliveInterval: keepaliveInterval,
		keepaliveReq:      keepaliveReq,
	}
	err := client.init()
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (c *ReConnectWSClient) Close() {
	close(c.closeChan)
}

func (c *ReConnectWSClient) init() error {
	conn, err := c.dial(c.addr)
	if err != nil {
		return err
	}
	session := NewSession(c, nil)
	session.Do(c.newCodec(NewWSConn(conn)))
	c.lock.Lock()
	c.session = session
	c.lock.Unlock()
	return nil
}

func (c *ReConnectWSClient) reconnect() {
	c.lock.Lock()
	c.session = nil
	c.lock.Unlock()

	for {
		select {
		case <-c.closeChan:
			return
		default:
		}
		if err := c.init(); err == nil {
			break
		}
		time.Sleep(5 * time.Second)
	}
}

func (c *ReConnectWSClient) keepalive() {
	tick := time.NewTicker(time.Second * time.Duration(c.keepaliveInterval))
	go func() {
		defer tick.Stop()
		for {
			select {
			case <-tick.C:
				c.sendKeepAlive()
			case <-c.closeChan:
				return
			}
		}
	}()
}

func (c *ReConnectWSClient) sendKeepAlive() {
	if c.keepaliveReq != nil {
		req := c.keepaliveReq()
		c.lock.RLock()
		if c.session != nil {
			c.session.AsyncWrite(req, time.Second*2)
		}
		c.lock.RUnlock()
	}
}

func (c *ReConnectWSClient) OnConnect(session *Session) {
	c.Protocol.OnConnect(session)
	c.keepaliveOnce.Do(c.keepalive)
}

func (c *ReConnectWSClient) OnDisconnect(session *Session) {
	c.Protocol.OnDisconnect(session)
	c.reconnect()
}

func (c *ReConnectWSClient) Send(data []byte) error {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if c.session != nil {
		return c.session.AsyncWrite(data, time.Second*2)
	}
	return nil
}
