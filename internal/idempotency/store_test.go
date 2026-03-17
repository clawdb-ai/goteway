package idempotency

import (
	"testing"
	"time"
)

func TestStoreGetExpiresAndEvicts(t *testing.T) {
	s := NewStore()
	base := time.Unix(100, 0)
	now := base
	s.now = func() time.Time { return now }

	s.Put("k1", "h1", map[string]any{"ok": true}, 5*time.Second)
	if _, ok := s.Get("k1"); !ok {
		t.Fatal("expected key to be present before ttl expiry")
	}

	now = base.Add(6 * time.Second)
	if _, ok := s.Get("k1"); ok {
		t.Fatal("expected key to be expired")
	}

	// Expired entry should be removed, and sweep should find nothing more.
	if removed := s.SweepExpired(0); removed != 0 {
		t.Fatalf("expected no extra expired entry after lazy eviction, removed=%d", removed)
	}
}

func TestStorePutGetIsolation(t *testing.T) {
	s := NewStore()
	in := map[string]any{"k": "v"}
	s.Put("key", "hash", in, time.Minute)
	in["k"] = "mutated"

	got, ok := s.Get("key")
	if !ok {
		t.Fatal("expected key to exist")
	}
	if got["k"] != "v" {
		t.Fatalf("expected immutable value v, got %#v", got["k"])
	}

	got["k"] = "changed-by-caller"
	again, ok := s.Get("key")
	if !ok {
		t.Fatal("expected key to exist on second read")
	}
	if again["k"] != "v" {
		t.Fatalf("expected immutable stored value v, got %#v", again["k"])
	}
}

func TestStoreMatchHash(t *testing.T) {
	s := NewStore()
	s.Put("key", "hash-a", map[string]any{"v": 1}, time.Minute)
	if !s.MatchHash("key", "hash-a") {
		t.Fatal("expected hash to match")
	}
	if s.MatchHash("key", "hash-b") {
		t.Fatal("expected hash mismatch")
	}
}

func TestStorePutWithNonPositiveTTLDeletes(t *testing.T) {
	s := NewStore()
	s.Put("key", "hash", map[string]any{"v": 1}, time.Minute)
	s.Put("key", "hash", map[string]any{"v": 2}, 0)

	if _, ok := s.Get("key"); ok {
		t.Fatal("expected key to be deleted for non-positive ttl")
	}
}
