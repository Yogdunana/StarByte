package cache

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"
)

const emptySentinel = "__empty__"

var ErrNotFound = errors.New("cache: key not found")

// Layered is L1 (local) + L2 (Redis) with singleflight, empty-value caching,
// bloom filter, and TTL jitter to reduce 雪崩.
type Layered struct {
	l1       *Local
	l2       *Store
	sf       *flight
	bloom    *Bloom
	emptyTTL time.Duration
	jitter   float64
}

func NewLayered(rdb *redis.Client) *Layered {
	return &Layered{
		l1:       NewLocal(4096),
		l2:       NewStore(rdb),
		sf:       newFlight(),
		bloom:    NewBloom(1<<16, 4),
		emptyTTL: 8 * time.Second,
		jitter:   0.12,
	}
}

func (c *Layered) L1() *Local { return c.l1 }
func (c *Layered) L2() *Store { return c.l2 }

func (c *Layered) jitterTTL(ttl time.Duration) time.Duration {
	if ttl <= 0 || c.jitter <= 0 {
		return ttl
	}
	delta := time.Duration(float64(ttl) * c.jitter * rand.Float64())
	return ttl + delta
}

func (c *Layered) Get(ctx context.Context, key string) (string, error) {
	if v, ok := c.l1.Get(key); ok {
		if v == emptySentinel {
			return "", ErrNotFound
		}
		return v, nil
	}
	if c.bloom != nil && !c.bloom.MightHave(key) {
		return "", ErrNotFound
	}
	v, err := c.l2.Get(ctx, key)
	if err == redis.Nil {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}
	c.l1.Set(key, v, time.Minute)
	if v == emptySentinel {
		return "", ErrNotFound
	}
	return v, nil
}

func (c *Layered) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	ttl = c.jitterTTL(ttl)
	if c.bloom != nil {
		c.bloom.Add(key)
	}
	c.l1.Set(key, value, ttl)
	return c.l2.Set(ctx, key, value, ttl)
}

func (c *Layered) Delete(ctx context.Context, key string) error {
	c.l1.Delete(key)
	_, err := c.l2.Del(ctx, key)
	return err
}

func (c *Layered) Stats() (hits, misses int64, size int) {
	return c.l1.Stats()
}

func (c *Layered) fromL1(key string) (string, error, bool) {
	v, ok := c.l1.Get(key)
	if !ok {
		return "", nil, false
	}
	if v == emptySentinel {
		return "", ErrNotFound, true
	}
	return v, nil, true
}

// GetOrLoad fills L1/L2 using loader on miss. Concurrent loads share one call.
// Bloom is not used here: a restart with an empty filter must still read L2,
// and brand-new keys must still run the loader then Add to the filter.
func (c *Layered) GetOrLoad(ctx context.Context, key string, ttl time.Duration, loader func() (string, error)) (string, error) {
	if v, err, ok := c.fromL1(key); ok {
		return v, err
	}
	return c.sf.Do(key, func() (string, error) {
		if v, err, ok := c.fromL1(key); ok {
			return v, err
		}
		v, err := c.l2.Get(ctx, key)
		if err == nil {
			c.l1.Set(key, v, time.Minute)
			if c.bloom != nil {
				c.bloom.Add(key)
			}
			if v == emptySentinel {
				return "", ErrNotFound
			}
			return v, nil
		}
		if err != redis.Nil {
			return "", err
		}
		val, loadErr := loader()
		if loadErr != nil {
			return "", loadErr
		}
		if val == "" {
			_ = c.Set(ctx, key, emptySentinel, c.emptyTTL)
			return "", ErrNotFound
		}
		if setErr := c.Set(ctx, key, val, ttl); setErr != nil {
			return "", setErr
		}
		return val, nil
	})
}
