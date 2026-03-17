package plugin

import "sync"

// Manifest is a minimal plugin manifest compatibility shape.
type Manifest struct {
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	Type         string   `json:"type"`
	Entry        string   `json:"entry"`
	Capabilities []string `json:"capabilities"`
}

// Registry keeps plugin manifests discovered by external loaders.
type Registry struct {
	mu    sync.RWMutex
	items map[string]Manifest
}

func NewRegistry() *Registry {
	return &Registry{items: make(map[string]Manifest)}
}

func (r *Registry) Upsert(m Manifest) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[m.Name] = m
}

func (r *Registry) List() []Manifest {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Manifest, 0, len(r.items))
	for _, it := range r.items {
		out = append(out, it)
	}
	return out
}
