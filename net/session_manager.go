package net

import (
	"context"
	"sync"
	"time"

	"github.com/hudangwei/common/logger"

	"go.uber.org/zap"
)

type Manager struct {
	group      []*sessionMap
	count      int
	expiration time.Duration
}

type sessionMap struct {
	dur   time.Duration
	items map[uint64]*item
	mu    *sync.RWMutex
}

type item struct {
	expiration   time.Duration
	session      *Session
	lastLiveTime time.Time
	rwLock       *sync.RWMutex
}

func (i *item) isExpire() bool {
	if i.expiration == 0 {
		return false
	}
	i.rwLock.RLock()
	defer i.rwLock.RUnlock()
	return time.Since(i.lastLiveTime) > i.expiration
}

func (i *item) resetTime() {
	i.rwLock.Lock()
	defer i.rwLock.Unlock()
	i.lastLiveTime = time.Now()
}

func newSessionMap(ctx context.Context, interval int64) *sessionMap {
	m := &sessionMap{
		items: make(map[uint64]*item),
		mu:    &sync.RWMutex{},
	}
	m.StartGC(ctx, interval)
	return m
}

func (sm *sessionMap) StartGC(ctx context.Context, interval int64) {
	sm.dur = time.Duration(interval) * time.Second
	go sm.checkExpire(ctx)
}

func (sm *sessionMap) checkExpire(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(sm.dur):
			sm.checkItemExpired()
		}
	}
}

func (sm *sessionMap) checkItemExpired() {
	var l []*item
	func() {
		sm.mu.RLock()
		defer sm.mu.RUnlock()

		for _, v := range sm.items {
			l = append(l, v)
		}
	}()
	for _, v := range l {
		if v.isExpire() {
			logger.Warn("session is expire,so delete session", zap.Uint64("sessionID", v.session.GetID()))
			v.session.Close()
		}
	}
}

func (sm *sessionMap) ResetExpireTime(sessionID uint64) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	if itm, ok := sm.items[sessionID]; ok {
		itm.resetTime()
	}
}

func (sm *sessionMap) Get(sessionID uint64) *Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	if itm, ok := sm.items[sessionID]; ok {
		if itm.isExpire() {
			logger.Warn("session is expire", zap.Uint64("sessionID", sessionID))
			return nil
		}
		return itm.session
	}
	return nil
}

func (sm *sessionMap) Put(session *Session, expiration time.Duration) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.items[session.id] = &item{
		session:      session,
		lastLiveTime: time.Now(),
		expiration:   expiration,
		rwLock:       &sync.RWMutex{},
	}
	return nil
}

func (sm *sessionMap) Delete(sessionID uint64) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.items, sessionID)
	return nil
}

func NewManager(ctx context.Context, count int, interval, expiration int64) *Manager {
	manager := &Manager{
		group:      make([]*sessionMap, count),
		count:      count,
		expiration: time.Duration(expiration) * time.Second,
	}
	for i := 0; i < count; i++ {
		manager.group[i] = newSessionMap(ctx, interval)
	}
	return manager
}

func (m *Manager) getShardingIndex(sessionID uint64) *sessionMap {
	return m.group[int(sessionID>>12)%m.count]
}

func (m *Manager) AddSession(session *Session, expiration time.Duration) {
	shard := m.getShardingIndex(session.id)
	shard.Put(session, expiration)
}

func (m *Manager) GetSession(sessionID uint64) (*Session, bool) {
	shard := m.getShardingIndex(sessionID)
	val := shard.Get(sessionID)
	return val, val != nil
}

func (m *Manager) DelSession(sessionID uint64) {
	shard := m.getShardingIndex(sessionID)
	shard.Delete(sessionID)
}

func (m *Manager) ResetExpireTime(sessionID uint64) {
	shard := m.getShardingIndex(sessionID)
	shard.ResetExpireTime(sessionID)
}

func (m *Manager) NewSession(protocol Protocol) *Session {
	session := NewSession(protocol, m)
	m.AddSession(session, m.expiration)
	return session
}
