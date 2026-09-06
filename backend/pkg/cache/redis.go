package cache

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// PoolHealth is Redis connection-pool + ping status for the admin UI.
type PoolHealth struct {
	PingOK     bool   `json:"ping_ok"`
	Hits       uint32 `json:"hits"`
	Misses     uint32 `json:"misses"`
	Timeouts   uint32 `json:"timeouts"`
	TotalConns uint32 `json:"total_conns"`
	IdleConns  uint32 `json:"idle_conns"`
	StaleConns uint32 `json:"stale_conns"`
}

// Store wraps a go-redis client for String/Hash/List/Set/ZSet, pipeline, and scan.
type Store struct {
	rdb *redis.Client
}

func NewStore(rdb *redis.Client) *Store {
	return &Store{rdb: rdb}
}

func (s *Store) Client() *redis.Client { return s.rdb }

func (s *Store) Ping(ctx context.Context) error {
	return s.rdb.Ping(ctx).Err()
}

func (s *Store) Get(ctx context.Context, key string) (string, error) {
	return s.rdb.Get(ctx, key).Result()
}

func (s *Store) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return s.rdb.Set(ctx, key, value, ttl).Err()
}

func (s *Store) SetNX(ctx context.Context, key, value string, ttl time.Duration) (bool, error) {
	return s.rdb.SetNX(ctx, key, value, ttl).Result()
}

func (s *Store) Del(ctx context.Context, keys ...string) (int64, error) {
	return s.rdb.Del(ctx, keys...).Result()
}

func (s *Store) Exists(ctx context.Context, key string) (bool, error) {
	n, err := s.rdb.Exists(ctx, key).Result()
	return n > 0, err
}

func (s *Store) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return s.rdb.Expire(ctx, key, ttl).Err()
}

func (s *Store) TTL(ctx context.Context, key string) (time.Duration, error) {
	return s.rdb.TTL(ctx, key).Result()
}

func (s *Store) HSet(ctx context.Context, key, field, value string) error {
	return s.rdb.HSet(ctx, key, field, value).Err()
}

func (s *Store) HGet(ctx context.Context, key, field string) (string, error) {
	return s.rdb.HGet(ctx, key, field).Result()
}

func (s *Store) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return s.rdb.HGetAll(ctx, key).Result()
}

func (s *Store) LPush(ctx context.Context, key string, values ...string) error {
	if len(values) == 0 {
		return nil
	}
	args := make([]interface{}, len(values))
	for i, v := range values {
		args[i] = v
	}
	return s.rdb.LPush(ctx, key, args...).Err()
}

func (s *Store) RPop(ctx context.Context, key string) (string, error) {
	return s.rdb.RPop(ctx, key).Result()
}

func (s *Store) SAdd(ctx context.Context, key string, members ...string) error {
	if len(members) == 0 {
		return nil
	}
	args := make([]interface{}, len(members))
	for i, v := range members {
		args[i] = v
	}
	return s.rdb.SAdd(ctx, key, args...).Err()
}

func (s *Store) SIsMember(ctx context.Context, key, member string) (bool, error) {
	return s.rdb.SIsMember(ctx, key, member).Result()
}

func (s *Store) ZAdd(ctx context.Context, key string, score float64, member string) error {
	return s.rdb.ZAdd(ctx, key, redis.Z{Score: score, Member: member}).Err()
}

func (s *Store) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return s.rdb.ZRange(ctx, key, start, stop).Result()
}

func (s *Store) Pipeline(ctx context.Context, fn func(redis.Pipeliner) error) error {
	pipe := s.rdb.Pipeline()
	if err := fn(pipe); err != nil {
		return err
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (s *Store) Tx(ctx context.Context, fn func(redis.Pipeliner) error) error {
	_, err := s.rdb.TxPipelined(ctx, fn)
	return err
}

// ScanKeys returns keys matching pattern (SCAN, not KEYS).
func (s *Store) ScanKeys(ctx context.Context, pattern string, limit int64) ([]string, error) {
	if limit <= 0 {
		limit = 200
	}
	var (
		cursor uint64
		out    []string
	)
	for {
		keys, next, err := s.rdb.Scan(ctx, cursor, pattern, 64).Result()
		if err != nil {
			return nil, err
		}
		out = append(out, keys...)
		cursor = next
		if cursor == 0 || int64(len(out)) >= limit {
			break
		}
	}
	if int64(len(out)) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (s *Store) Health(ctx context.Context) (*PoolHealth, error) {
	h := &PoolHealth{}
	if s == nil || s.rdb == nil {
		return h, errors.New("cache: redis client is nil")
	}
	if err := s.Ping(ctx); err != nil {
		return h, err
	}
	h.PingOK = true
	st := s.rdb.PoolStats()
	if st != nil {
		h.Hits = st.Hits
		h.Misses = st.Misses
		h.Timeouts = st.Timeouts
		h.TotalConns = st.TotalConns
		h.IdleConns = st.IdleConns
		h.StaleConns = st.StaleConns
	}
	return h, nil
}
