package idempotency

import (
	"sync"
	"time"
)

// Store keeps idempotent operation results for a bounded TTL window.
type Store struct {
	shards [storeShardCount]shard
	now    func() time.Time
}

type entry struct {
	hash      string
	result    map[string]any
	expiresAt int64
}

type shard struct {
	mu    sync.RWMutex
	items map[string]entry
}

func NewStore() *Store {
	s := &Store{now: time.Now}
	for i := range s.shards {
		s.shards[i].items = make(map[string]entry)
	}
	return s
}

const storeShardCount = 64

func (s *Store) Get(key string) (result map[string]any, ok bool) {
	nowUnix := s.now().UnixNano()
	sh := s.shardFor(key)

	sh.mu.RLock()
	it, found := sh.items[key]
	sh.mu.RUnlock()
	if !found {
		return nil, false
	}
	if nowUnix > it.expiresAt {
		sh.mu.Lock()
		if cur, stillFound := sh.items[key]; stillFound && nowUnix > cur.expiresAt {
			delete(sh.items, key)
		}
		sh.mu.Unlock()
		return nil, false
	}
	return cloneMap(it.result), true
}

func (s *Store) Put(key, hash string, result map[string]any, ttl time.Duration) {
	if key == "" {
		return
	}
	if ttl <= 0 {
		s.Delete(key)
		return
	}
	expiresAt := s.now().Add(ttl).UnixNano()
	sh := s.shardFor(key)
	sh.mu.Lock()
	sh.items[key] = entry{
		hash:      hash,
		result:    cloneMap(result),
		expiresAt: expiresAt,
	}
	sh.mu.Unlock()
}

func (s *Store) Delete(key string) {
	sh := s.shardFor(key)
	sh.mu.Lock()
	delete(sh.items, key)
	sh.mu.Unlock()
}

// MatchHash reports whether key exists, is not expired, and hash is equal.
func (s *Store) MatchHash(key, hash string) bool {
	nowUnix := s.now().UnixNano()
	sh := s.shardFor(key)

	sh.mu.RLock()
	it, found := sh.items[key]
	sh.mu.RUnlock()
	if !found {
		return false
	}
	if nowUnix > it.expiresAt {
		sh.mu.Lock()
		if cur, stillFound := sh.items[key]; stillFound && nowUnix > cur.expiresAt {
			delete(sh.items, key)
		}
		sh.mu.Unlock()
		return false
	}
	return it.hash == hash
}

// SweepExpired removes expired entries up to maxPerShard per shard.
// Use maxPerShard <= 0 to remove all expired entries.
func (s *Store) SweepExpired(maxPerShard int) (removed int) {
	nowUnix := s.now().UnixNano()
	for i := range s.shards {
		sh := &s.shards[i]
		sh.mu.Lock()
		n := 0
		for k, it := range sh.items {
			if nowUnix <= it.expiresAt {
				continue
			}
			delete(sh.items, k)
			removed++
			n++
			if maxPerShard > 0 && n >= maxPerShard {
				break
			}
		}
		sh.mu.Unlock()
	}
	return removed
}

func (s *Store) shardFor(key string) *shard {
	idx := fnv1a64(key) % storeShardCount
	return &s.shards[idx]
}

func fnv1a64(s string) uint64 {
	var h uint64 = 1469598103934665603
	for i := 0; i < len(s); i++ {
		h ^= uint64(s[i])
		h *= 1099511628211
	}
	return h
}

func cloneMap(in map[string]any) map[string]any {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
