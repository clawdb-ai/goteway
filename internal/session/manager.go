package session

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// Manager stores active client sessions.
type Manager struct {
	nextID  uint64
	clients sync.Map // clientID -> metadata
}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) NewClientID() string {
	n := atomic.AddUint64(&m.nextID, 1)
	return fmt.Sprintf("cli_%d", n)
}

func (m *Manager) Put(clientID string, meta map[string]any) {
	m.clients.Store(clientID, meta)
}

func (m *Manager) Get(clientID string) (map[string]any, bool) {
	v, ok := m.clients.Load(clientID)
	if !ok {
		return nil, false
	}
	out, _ := v.(map[string]any)
	return out, true
}
