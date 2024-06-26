package net

import "io"

type Codec interface {
	Send([]byte) error
	Recv() (interface{}, error)
	Close() error
}

type ReadWriteCloser interface {
	io.Writer
	io.Closer
	Reader() (io.Reader, error)
}

type NewCodecFunc func(ReadWriteCloser) Codec
