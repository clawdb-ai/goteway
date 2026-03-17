package plugin

import (
	"slices"
	"sync"
)

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
	r.items[m.Name] = cloneManifest(m)
}

func (r *Registry) List() []Manifest {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Manifest, 0, len(r.items))
	for _, it := range r.items {
		out = append(out, cloneManifest(it))
	}
	slices.SortFunc(out, func(a, b Manifest) int {
		switch {
		case a.Name < b.Name:
			return -1
		case a.Name > b.Name:
			return 1
		default:
			return 0
		}
	})
	return out
}

func cloneManifest(m Manifest) Manifest {
	if len(m.Capabilities) == 0 {
		return m
	}
	out := m
	out.Capabilities = append([]string(nil), m.Capabilities...)
	return out
}
