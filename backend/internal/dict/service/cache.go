package service

import (
	"context"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type DictCache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, val string, ttl time.Duration) error
	Del(ctx context.Context, keys ...string) error
}

type redisCache struct {
	c *redis.Client
}

func NewRedisCache(c *redis.Client) DictCache {
	if c == nil {
		return nil
	}
	return &redisCache{c: c}
}

func (r *redisCache) Get(ctx context.Context, key string) (string, error) {
	return r.c.Get(ctx, key).Result()
}

func (r *redisCache) Set(ctx context.Context, key string, val string, ttl time.Duration) error {
	return r.c.Set(ctx, key, val, ttl).Err()
}

func (r *redisCache) Del(ctx context.Context, keys ...string) error {
	return r.c.Del(ctx, keys...).Err()
}

type memCache struct {
	mu   sync.Mutex
	data map[string]string
}

func newMemCache() *memCache {
	return &memCache{data: map[string]string{}}
}

func (m *memCache) Get(_ context.Context, key string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	val, ok := m.data[key]
	if !ok {
		return "", redis.Nil
	}
	return val, nil
}

func (m *memCache) Set(_ context.Context, key string, val string, _ time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = val
	return nil
}

func (m *memCache) Del(_ context.Context, keys ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, key := range keys {
		delete(m.data, key)
	}
	return nil
}
