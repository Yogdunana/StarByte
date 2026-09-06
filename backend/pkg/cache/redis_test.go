package cache

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore_StringHashListSetZSet(t *testing.T) {
	_, rdb := testRedis(t)
	s := NewStore(rdb)
	ctx := context.Background()

	require.NoError(t, s.Ping(ctx))
	require.NoError(t, s.Set(ctx, "k", "v", time.Minute))
	got, err := s.Get(ctx, "k")
	require.NoError(t, err)
	assert.Equal(t, "v", got)

	ok, err := s.SetNX(ctx, "k", "other", time.Minute)
	require.NoError(t, err)
	assert.False(t, ok)
	ok, err = s.Exists(ctx, "k")
	require.NoError(t, err)
	assert.True(t, ok)
	require.NoError(t, s.Expire(ctx, "k", time.Hour))
	ttl, err := s.TTL(ctx, "k")
	require.NoError(t, err)
	assert.Greater(t, ttl, time.Minute)

	require.NoError(t, s.HSet(ctx, "h", "f", "1"))
	hv, err := s.HGet(ctx, "h", "f")
	require.NoError(t, err)
	assert.Equal(t, "1", hv)
	all, err := s.HGetAll(ctx, "h")
	require.NoError(t, err)
	assert.Equal(t, "1", all["f"])

	require.NoError(t, s.LPush(ctx, "l"))
	require.NoError(t, s.LPush(ctx, "l", "a", "b"))
	pop, err := s.RPop(ctx, "l")
	require.NoError(t, err)
	assert.Equal(t, "a", pop)

	require.NoError(t, s.SAdd(ctx, "set"))
	require.NoError(t, s.SAdd(ctx, "set", "m1"))
	member, err := s.SIsMember(ctx, "set", "m1")
	require.NoError(t, err)
	assert.True(t, member)

	require.NoError(t, s.ZAdd(ctx, "z", 2, "b"))
	require.NoError(t, s.ZAdd(ctx, "z", 1, "a"))
	zr, err := s.ZRange(ctx, "z", 0, -1)
	require.NoError(t, err)
	assert.Equal(t, []string{"a", "b"}, zr)

	n, err := s.Del(ctx, "k")
	require.NoError(t, err)
	assert.Equal(t, int64(1), n)
	assert.Same(t, rdb, s.Client())
}

func TestStore_PipelineTxScanHealth(t *testing.T) {
	_, rdb := testRedis(t)
	s := NewStore(rdb)
	ctx := context.Background()

	require.NoError(t, s.Pipeline(ctx, func(p redis.Pipeliner) error {
		p.Set(ctx, "p1", "1", time.Minute)
		p.Set(ctx, "p2", "2", time.Minute)
		return nil
	}))
	require.NoError(t, s.Tx(ctx, func(p redis.Pipeliner) error {
		p.Set(ctx, "tx", "ok", time.Minute)
		return nil
	}))
	keys, err := s.ScanKeys(ctx, "p*", 10)
	require.NoError(t, err)
	assert.Len(t, keys, 2)
	keys, err = s.ScanKeys(ctx, "*", 1)
	require.NoError(t, err)
	assert.Len(t, keys, 1)

	h, err := s.Health(ctx)
	require.NoError(t, err)
	assert.True(t, h.PingOK)
	assert.GreaterOrEqual(t, h.TotalConns, uint32(1))

	_, err = (*Store)(nil).Health(ctx)
	assert.Error(t, err)
	_, err = NewStore(nil).Health(ctx)
	assert.Error(t, err)
}

func TestStore_PipelineFnError(t *testing.T) {
	_, rdb := testRedis(t)
	s := NewStore(rdb)
	err := s.Pipeline(context.Background(), func(redis.Pipeliner) error {
		return assert.AnError
	})
	assert.Error(t, err)
}
