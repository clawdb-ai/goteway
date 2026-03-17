package session

import (
	"strconv"
	"sync"
	"sync/atomic"
)

// Manager stores active client sessions.
type Manager struct {
	nextID  atomic.Uint64
	clients sync.Map // clientID -> ClientMeta
}

// ClientMeta contains normalized session metadata.
type ClientMeta struct {
	Client map[string]any
}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) NewClientID() string {
	n := m.nextID.Add(1)
	return formatClientID(n)
}

func (m *Manager) Put(clientID string, meta ClientMeta) {
	m.clients.Store(clientID, ClientMeta{
		Client: cloneMap(meta.Client),
	})
}

func (m *Manager) Get(clientID string) (ClientMeta, bool) {
	v, ok := m.clients.Load(clientID)
	if !ok {
		return ClientMeta{}, false
	}
	meta, _ := v.(ClientMeta)
	return ClientMeta{
		Client: cloneMap(meta.Client),
	}, true
}

func (m *Manager) Delete(clientID string) {
	m.clients.Delete(clientID)
}

const clientIDPrefix = "cli_"

func formatClientID(n uint64) string {
	var b [len(clientIDPrefix) + 20]byte
	copy(b[:], clientIDPrefix)
	out := strconv.AppendUint(b[:len(clientIDPrefix)], n, 10)
	return string(out)
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
