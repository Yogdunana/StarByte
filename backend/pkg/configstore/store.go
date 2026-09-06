package configstore

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	cachePrefix = "configstore:"
	cacheTTL    = 10 * time.Minute
)

// Backend 从持久层读写配置值。实现放在业务模块，避免本包依赖 GORM。
type Backend interface {
	Load(ctx context.Context, key string) (value string, ok bool, err error)
	Save(ctx context.Context, key, value string) error
}

// Store 运行时业务配置 SDK。不要与 pkg/config 的 YAML loader 混淆。
type Store interface {
	Get(ctx context.Context, key string) (string, error)
	GetBool(ctx context.Context, key string) (bool, error)
	Set(ctx context.Context, key, value string) error
	Invalidate(ctx context.Context, key string) error
}

type redisStore struct {
	rdb     *redis.Client
	backend Backend
}

// New 用 Redis 缓存包装 Backend。rdb 可为 nil（仅走 Backend，便于单测）。
func New(rdb *redis.Client, backend Backend) Store {
	return &redisStore{rdb: rdb, backend: backend}
}

func cacheKey(key string) string {
	return cachePrefix + key
}

func (s *redisStore) Get(ctx context.Context, key string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("configstore: empty key")
	}
	if s.rdb != nil {
		cached, err := s.rdb.Get(ctx, cacheKey(key)).Result()
		if err == nil {
			return cached, nil
		}
		if err != redis.Nil {
			return "", fmt.Errorf("configstore cache get: %w", err)
		}
	}
	value, ok, err := s.backend.Load(ctx, key)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", ErrNotFound
	}
	if s.rdb != nil {
		_ = s.rdb.Set(ctx, cacheKey(key), value, cacheTTL).Err()
	}
	return value, nil
}

func (s *redisStore) GetBool(ctx context.Context, key string) (bool, error) {
	raw, err := s.Get(ctx, key)
	if err != nil {
		return false, err
	}
	v, err := parseBool(raw)
	if err != nil {
		return false, err
	}
	return v, nil
}

func (s *redisStore) Set(ctx context.Context, key, value string) error {
	if key == "" {
		return fmt.Errorf("configstore: empty key")
	}
	if err := s.backend.Save(ctx, key, value); err != nil {
		return err
	}
	return s.Invalidate(ctx, key)
}

func (s *redisStore) Invalidate(ctx context.Context, key string) error {
	if s.rdb == nil || key == "" {
		return nil
	}
	if err := s.rdb.Del(ctx, cacheKey(key)).Err(); err != nil {
		return fmt.Errorf("configstore invalidate: %w", err)
	}
	return nil
}

func parseBool(raw string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "true", "yes", "on":
		return true, nil
	case "0", "false", "no", "off", "":
		return false, nil
	default:
		v, err := strconv.ParseBool(raw)
		if err != nil {
			return false, fmt.Errorf("configstore: not a boolean: %q", raw)
		}
		return v, nil
	}
}
