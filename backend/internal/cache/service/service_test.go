package service

import (
	"context"
	"testing"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/cache/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newSvc(t *testing.T) CacheService {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	svc := NewCacheService(rdb)
	t.Cleanup(svc.Close)
	return svc
}

func TestCacheService_StatsDeleteWarmup(t *testing.T) {
	svc := newSvc(t)
	ctx := context.Background()

	_, err := svc.Warmup(ctx, &dto.WarmupRequest{
		Entries: []dto.WarmupEntry{{Key: "demo:a", Value: "1", TTLSeconds: 60}},
	})
	require.NoError(t, err)

	st, err := svc.Stats(ctx, "demo:*")
	require.NoError(t, err)
	assert.True(t, st.Healthy)
	assert.GreaterOrEqual(t, st.KeyCount, 1)

	require.NoError(t, svc.DeleteKey(ctx, "demo:a"))
	err = svc.DeleteKey(ctx, "demo:a")
	require.Error(t, err)
	app, ok := err.(*response.AppError)
	require.True(t, ok)
	assert.Equal(t, response.CodeCacheKeyNotFound, app.Code)

	_, err = svc.Warmup(ctx, &dto.WarmupRequest{
		Entries: []dto.WarmupEntry{
			{Key: "demo:b", Value: "2", TTLSeconds: 0},
			{Key: "demo:c", Value: "3", TTLSeconds: 30},
		},
	})
	require.NoError(t, err)
	del, err := svc.DeletePattern(ctx, "demo:*")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, del.Deleted, int64(2))
}

func TestCacheService_InvalidInputs(t *testing.T) {
	svc := newSvc(t)
	ctx := context.Background()

	err := svc.DeleteKey(ctx, "  ")
	require.Error(t, err)
	_, err = svc.DeletePattern(ctx, "*")
	require.Error(t, err)
	_, err = svc.DeletePattern(ctx, "x")
	require.Error(t, err)
	_, err = svc.Warmup(ctx, nil)
	require.Error(t, err)
	_, err = svc.Warmup(ctx, &dto.WarmupRequest{})
	require.Error(t, err)
	_, err = svc.Warmup(ctx, &dto.WarmupRequest{Entries: []dto.WarmupEntry{{Key: "  "}}})
	require.Error(t, err)

	del, err := svc.DeletePattern(ctx, "none:*")
	require.NoError(t, err)
	assert.Equal(t, int64(0), del.Deleted)

	st, err := svc.Stats(ctx, "")
	require.NoError(t, err)
	assert.Equal(t, "*", st.Pattern)
}

func TestCacheService_WarmupScanPrefix(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	require.NoError(t, rdb.Set(context.Background(), "dict:status", "ok", time.Minute).Err())
	svc := NewCacheService(rdb)
	t.Cleanup(svc.Close)
	out, err := svc.Warmup(context.Background(), &dto.WarmupRequest{ScanPrefix: "dict:*"})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, out.Loaded, 1)
}
