package configstore

import (
	"context"
	"sync"
)

// MemoryBackend 进程内 Backend，供单测使用。
type MemoryBackend struct {
	mu   sync.RWMutex
	data map[string]string
}

func NewMemoryBackend() *MemoryBackend {
	return &MemoryBackend{data: map[string]string{}}
}

func (m *MemoryBackend) Load(_ context.Context, key string) (string, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.data[key]
	return v, ok, nil
}

func (m *MemoryBackend) Save(_ context.Context, key, value string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
	return nil
}
