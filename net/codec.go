package net

import "io"

type Codec interface {
	Send(string) error
	Recv() (interface{}, error)
	Close() error
}

type NewCodecFunc func(io.ReadWriteCloser) Codec
