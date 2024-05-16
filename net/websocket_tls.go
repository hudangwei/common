package net

import (
	"net/http"
)

type WebsocketTLSServer struct {
	manager  *Manager
	newCodec NewCodecFunc
	protocol Protocol
}

func NewWebsocketTLSServer(manager *Manager, newCodec NewCodecFunc, protocol Protocol) *WebsocketTLSServer {
	return &WebsocketTLSServer{
		manager:  manager,
		newCodec: newCodec,
		protocol: protocol,
	}
}

func (s *WebsocketTLSServer) Start(addr, certFile, privateFile string) {
	mu := http.NewServeMux()
	mu.HandleFunc("/", s.wsHandler)
	go http.ListenAndServeTLS(addr, certFile, privateFile, mu)
}

func (s *WebsocketTLSServer) wsHandler(w http.ResponseWriter, r *http.Request) {
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
}
