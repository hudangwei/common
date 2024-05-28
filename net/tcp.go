package net

import "net"

type TCPServer struct {
	listener net.Listener
	manager  *Manager
	newCodec NewCodecFunc
	protocol Protocol
}

func NewTCPServer(manager *Manager, newCodec NewCodecFunc, protocol Protocol) *TCPServer {
	s := &TCPServer{
		manager:  manager,
		newCodec: newCodec,
		protocol: protocol,
	}
	return s
}

func (s *TCPServer) Start(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.listener = listener
	go s.Serve()
	return nil
}

func (s *TCPServer) Serve() error {
	for {
		conn, err := Accept(s.listener)
		if err != nil {
			return err
		}

		session := s.manager.NewSession(s.protocol)
		session.PutExtraData("remote_addr", conn.RemoteAddr().String())
		codec := s.newCodec(NewTCPConn(conn))
		session.Do(codec)
	}
}

func (s *TCPServer) Stop() {
	s.listener.Close()
}
