package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLocal_HitMissExpireEvict(t *testing.T) {
	l := NewLocal(0)
	assert.Equal(t, 0, l.Len())

	l.Set("a", "1", 50*time.Millisecond)
	v, ok := l.Get("a")
	assert.True(t, ok)
	assert.Equal(t, "1", v)
	_, ok = l.Get("missing")
	assert.False(t, ok)

	time.Sleep(60 * time.Millisecond)
	_, ok = l.Get("a")
	assert.False(t, ok)

	small := NewLocal(2)
	small.Set("k1", "1", 0)
	small.Set("k2", "2", 0)
	small.Set("k3", "3", 0)
	assert.Equal(t, 2, small.Len())
	small.Delete("k3")
	hits, misses, size := small.Stats()
	assert.GreaterOrEqual(t, misses, int64(0))
	assert.GreaterOrEqual(t, hits, int64(0))
	assert.Equal(t, 1, size)
}

func TestLocal_NeverExpireNotPreferred(t *testing.T) {
	l := NewLocal(2)
	l.Set("perm", "p", 0)
	l.Set("temp", "t", time.Millisecond)
	time.Sleep(3 * time.Millisecond)
	l.Set("new", "n", time.Minute)
	v, ok := l.Get("perm")
	assert.True(t, ok)
	assert.Equal(t, "p", v)
}

func TestLocal_EvictExpiredFirst(t *testing.T) {
	l := NewLocal(1)
	l.Set("old", "x", time.Millisecond)
	time.Sleep(3 * time.Millisecond)
	l.Set("new", "y", time.Minute)
	v, ok := l.Get("new")
	assert.True(t, ok)
	assert.Equal(t, "y", v)
}
