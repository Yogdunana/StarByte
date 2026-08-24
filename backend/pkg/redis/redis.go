package redis

import (
	"context"
	"fmt"

	"github.com/Yogdunana/StarByte/backend/pkg/config"
	"github.com/redis/go-redis/v9"
)

var client *redis.Client

// Init creates and verifies a Redis client using the provided configuration.
func Init(cfg *config.RedisConfig) error {
	client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("connect redis: %w", err)
	}

	return nil
}

// Close releases the Redis client resources.
func Close() error {
	if client == nil {
		return nil
	}
	return client.Close()
}

// Client returns the global *redis.Client instance.
func Client() *redis.Client {
	return client
}

// Get returns the global *redis.Client instance. It is an alias for Client
// that exposes the underlying client for arbitrary operations.
func Get() *redis.Client {
	return client
}

// Exists reports whether the given key exists in Redis.
func Exists(ctx context.Context, key string) (bool, error) {
	n, err := client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}
