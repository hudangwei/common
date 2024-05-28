package net

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hudangwei/common/util"
)

var (
	ErrConnClosing   = errors.New("use of closed connection")
	ErrWriteBlocking = errors.New("write data was blocking")
)
var globalSessionID uint64

type Session struct {
	id        uint64
	codec     Codec
	protocol  Protocol
	manager   *Manager
	closeOnce sync.Once
	closeFlag int32
	closeChan chan struct{}
	dataChan  chan []byte
	sendMutex *sync.Mutex

	extraData map[string]interface{}
	mu        *sync.RWMutex
}

func NewSession(protocol Protocol, manager *Manager) *Session {
	ret := &Session{
		id:        atomic.AddUint64(&globalSessionID, 1),
		protocol:  protocol,
		manager:   manager,
		closeChan: make(chan struct{}),
		dataChan:  make(chan []byte, 100),
		sendMutex: &sync.Mutex{},
		extraData: make(map[string]interface{}),
		mu:        &sync.RWMutex{},
	}
	if protocol != nil {
		protocol.OnConnect(ret)
	}
	return ret
}

func (s *Session) GetExtraData(k string) interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	v, ok := s.extraData[k]
	if !ok {
		return nil
	}
	return v
}

func (s *Session) PutExtraData(k string, v interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.extraData[k] = v
}

func (s *Session) Close() {
	s.closeOnce.Do(func() {
		atomic.StoreInt32(&s.closeFlag, 1)
		close(s.closeChan)
		close(s.dataChan)
		if s.codec != nil {
			s.codec.Close()
		}
		if s.manager != nil {
			s.manager.DelSession(s.id)
		}
		if s.protocol != nil {
			s.protocol.OnDisconnect(s)
		}
	})
}

func (s *Session) IsClosed() bool {
	return atomic.LoadInt32(&s.closeFlag) == 1
}

func (s *Session) AsyncWrite(data []byte, timeout time.Duration) (err error) {
	if s.IsClosed() {
		return ErrConnClosing
	}
	defer func() {
		if e := recover(); e != nil {
			err = ErrConnClosing
		}
	}()
	if timeout == 0 {
		select {
		case s.dataChan <- data:
			return nil
		case <-s.closeChan:
			return ErrConnClosing
		default:
			return ErrWriteBlocking
		}
	} else {
		select {
		case s.dataChan <- data:
			return nil
		case <-s.closeChan:
			return ErrConnClosing
		case <-time.After(timeout):
			return ErrWriteBlocking
		}
	}
}

func (s *Session) Send(data []byte) error {
	if s.IsClosed() {
		return ErrConnClosing
	}
	s.sendMutex.Lock()
	defer s.sendMutex.Unlock()

	err := s.codec.Send(data)
	if err != nil {
		return err
	}
	return nil
}

func (s *Session) readLoop() {
	defer func() {
		if e := recover(); e != nil {
			util.PanicHandler(e)
		}
		s.Close()
	}()

	for {
		select {
		case <-s.closeChan:
			return
		default:
		}
		if s.IsClosed() {
			return
		}
		data, err := s.codec.Recv()
		if err != nil {
			return
		}
		if s.protocol != nil {
			err := s.protocol.Verify(data)
			if err != nil {
				return
			}
			err = s.protocol.OnMessage(s, data)
			if err != nil {
				return
			}
		}
	}
}

func (s *Session) writeLoop() {
	defer func() {
		if e := recover(); e != nil {
			util.PanicHandler(e)
		}
		s.Close()
	}()

	for {
		select {
		case <-s.closeChan:
			return
		case data := <-s.dataChan:
			if s.IsClosed() {
				return
			}
			err := s.codec.Send(data)
			if err != nil {
				return
			}
		}
	}
}

func (s *Session) Do(codec Codec) {
	s.codec = codec
	go s.readLoop()
	go s.writeLoop()
}

func (s *Session) GetID() uint64 {
	return s.id
}

func (s *Session) GetRemoteIP() string {
	if v := s.GetExtraData("remote_addr"); v != nil {
		return v.(string)
	}
	return ""
}

func (s *Session) ResetExpireTime() {
	if s.manager != nil {
		s.manager.ResetExpireTime(s.id)
	}
}
