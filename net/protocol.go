package net

type Protocol interface {
	OnConnect(session *Session)
	OnDisconnect(session *Session)
	Verify(data interface{}) error
	OnMessage(session *Session, data interface{}) error
}
