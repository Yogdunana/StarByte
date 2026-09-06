package ratelimit

import (
	"context"
	"math"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// tokenBucketScript is a Redis token bucket. Returns {allowed, remaining, retry_after_sec}.
var tokenBucketScript = redis.NewScript(`
	local key = KEYS[1]
	local rate = tonumber(ARGV[1])
	local burst = tonumber(ARGV[2])
	local now = tonumber(ARGV[3])
	local requested = tonumber(ARGV[4])

	local data = redis.call('HMGET', key, 'tokens', 'ts')
	local tokens = tonumber(data[1])
	local ts = tonumber(data[2])
	if tokens == nil then
		tokens = burst
		ts = now
	end
	local elapsed = math.max(0, now - ts) / 1000.0
	tokens = math.min(burst, tokens + elapsed * rate)
	local allowed = 0
	local retry_after = 0
	if tokens >= requested then
		tokens = tokens - requested
		allowed = 1
	else
		if rate > 0 then
			retry_after = math.ceil((requested - tokens) / rate)
		else
			retry_after = 1
		end
	end
	redis.call('HSET', key, 'tokens', tokens, 'ts', now)
	local ttl = 2
	if rate > 0 then
		ttl = math.ceil(burst / rate) + 2
	end
	redis.call('EXPIRE', key, ttl)
	return {allowed, math.floor(tokens), retry_after}
`)

// Store talks to Redis. Nil or failing Redis falls back to a process-local bucket.
type Store struct {
	rdb   *redis.Client
	now   func() time.Time
	local *localLimiter
}

func NewStore(rdb *redis.Client) *Store {
	return &Store{rdb: rdb, now: time.Now, local: newLocalLimiter()}
}

type Result struct {
	Allowed    bool
	Remaining  int64
	RetryAfter int
}

func (s *Store) Allow(ctx context.Context, key string, b Bucket) (Result, error) {
	if s == nil {
		return Result{Allowed: true, Remaining: int64(b.Burst)}, nil
	}
	if b.Rate <= 0 || b.Burst <= 0 {
		return Result{Allowed: true}, nil
	}
	if s.rdb == nil {
		return s.local.allow(key, b), nil
	}
	now := s.now()
	raw, err := tokenBucketScript.Run(ctx, s.rdb, []string{key},
		strconv.FormatFloat(b.Rate, 'f', 6, 64),
		strconv.FormatFloat(b.Burst, 'f', 6, 64),
		now.UnixMilli(),
		1,
	).Result()
	if err != nil {
		return s.local.allow(key, b), nil
	}
	vals, ok := raw.([]interface{})
	if !ok || len(vals) < 3 {
		return s.local.allow(key, b), nil
	}
	allowed, _ := vals[0].(int64)
	remaining, _ := vals[1].(int64)
	retry, _ := vals[2].(int64)
	return Result{
		Allowed:    allowed == 1,
		Remaining:  remaining,
		RetryAfter: int(math.Max(1, float64(retry))),
	}, nil
}
