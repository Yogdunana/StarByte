package cache

import (
	"sync"
	"time"
)

type localEntry struct {
	value     string
	expiresAt time.Time
}

// Local is an in-process L1 cache with per-key TTL.
type Local struct {
	mu      sync.RWMutex
	items   map[string]localEntry
	hits    int64
	misses  int64
	maxSize int
}

func NewLocal(maxSize int) *Local {
	if maxSize <= 0 {
		maxSize = 4096
	}
	return &Local{items: map[string]localEntry{}, maxSize: maxSize}
}

func (l *Local) Get(key string) (string, bool) {
	l.mu.RLock()
	ent, ok := l.items[key]
	l.mu.RUnlock()
	if !ok {
		l.mu.Lock()
		l.misses++
		l.mu.Unlock()
		return "", false
	}
	if !ent.expiresAt.IsZero() && time.Now().After(ent.expiresAt) {
		l.Delete(key)
		l.mu.Lock()
		l.misses++
		l.mu.Unlock()
		return "", false
	}
	l.mu.Lock()
	l.hits++
	l.mu.Unlock()
	return ent.value, true
}

func (l *Local) Set(key, value string, ttl time.Duration) {
	ent := localEntry{value: value}
	if ttl > 0 {
		ent.expiresAt = time.Now().Add(ttl)
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.items) >= l.maxSize {
		// drop one arbitrary expired or first key
		for k, e := range l.items {
			if e.expiresAt.IsZero() || time.Now().After(e.expiresAt) {
				delete(l.items, k)
				break
			}
		}
		if len(l.items) >= l.maxSize {
			for k := range l.items {
				delete(l.items, k)
				break
			}
		}
	}
	l.items[key] = ent
}

func (l *Local) Delete(key string) {
	l.mu.Lock()
	delete(l.items, key)
	l.mu.Unlock()
}

func (l *Local) Len() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.items)
}

func (l *Local) Stats() (hits, misses int64, size int) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.hits, l.misses, len(l.items)
}
