package idempotency

import (
	"sync"
	"time"
)

// Store keeps idempotent operation results for a bounded TTL window.
type Store struct {
	mu    sync.RWMutex
	items map[string]entry
}

type entry struct {
	hash      string
	result    map[string]any
	expiresAt time.Time
}

func NewStore() *Store {
	return &Store{items: make(map[string]entry)}
}

func (s *Store) Get(key string) (result map[string]any, ok bool) {
	now := time.Now()
	s.mu.RLock()
	it, found := s.items[key]
	s.mu.RUnlock()
	if !found || now.After(it.expiresAt) {
		return nil, false
	}
	return it.result, true
}

func (s *Store) Put(key, hash string, result map[string]any, ttl time.Duration) {
	s.mu.Lock()
	s.items[key] = entry{hash: hash, result: result, expiresAt: time.Now().Add(ttl)}
	s.mu.Unlock()
}
