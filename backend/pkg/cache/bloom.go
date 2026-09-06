package cache

import (
	"hash/fnv"
	"sync"
)

// Bloom is a simple in-memory bloom filter used to skip Redis on missing keys (穿透).
type Bloom struct {
	mu   sync.RWMutex
	bits []uint64
	k    int
	m    uint64
}

func NewBloom(bitCount, hashCount int) *Bloom {
	if bitCount < 64 {
		bitCount = 1 << 14
	}
	if hashCount < 1 {
		hashCount = 4
	}
	n := (bitCount + 63) / 64
	return &Bloom{bits: make([]uint64, n), k: hashCount, m: uint64(n * 64)}
}

func (b *Bloom) Add(key string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for i := 0; i < b.k; i++ {
		idx := b.index(key, i)
		b.bits[idx/64] |= 1 << (idx % 64)
	}
}

func (b *Bloom) MightHave(key string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for i := 0; i < b.k; i++ {
		idx := b.index(key, i)
		if b.bits[idx/64]&(1<<(idx%64)) == 0 {
			return false
		}
	}
	return true
}

func (b *Bloom) index(key string, i int) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte{byte(i)})
	_, _ = h.Write([]byte(key))
	return h.Sum64() % b.m
}
