package net

type Codec interface {
	Send(string) error
	Recv() (interface{}, error)
	Close() error
}

type JsonReadWriter interface {
	WriteJSON(v string) error
	ReadJSON(v interface{}) error
	Close() error
}

type NewCodecFunc func(rw JsonReadWriter) Codec
