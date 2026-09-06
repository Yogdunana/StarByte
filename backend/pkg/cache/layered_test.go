package cache

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLayered_L1L2Consistent(t *testing.T) {
	_, rdb := testRedis(t)
	c := NewLayered(rdb)
	ctx := context.Background()

	require.NoError(t, c.Set(ctx, "user:1", "alice", time.Minute))
	v, err := c.Get(ctx, "user:1")
	require.NoError(t, err)
	assert.Equal(t, "alice", v)

	c.L1().Delete("user:1")
	v, err = c.Get(ctx, "user:1")
	require.NoError(t, err)
	assert.Equal(t, "alice", v)

	require.NoError(t, c.Delete(ctx, "user:1"))
	_, err = c.Get(ctx, "user:1")
	assert.ErrorIs(t, err, ErrNotFound)

	hits, misses, size := c.Stats()
	assert.GreaterOrEqual(t, hits+misses, int64(1))
	assert.GreaterOrEqual(t, size, 0)
	assert.NotNil(t, c.L2())
}

func TestLayered_BloomSkipUnknown(t *testing.T) {
	_, rdb := testRedis(t)
	c := NewLayered(rdb)
	_, err := c.Get(context.Background(), "never-seen")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestLayered_GetOrLoadSingleflightAndEmpty(t *testing.T) {
	_, rdb := testRedis(t)
	c := NewLayered(rdb)
	ctx := context.Background()

	var calls int32
	var wg sync.WaitGroup
	wg.Add(20)
	for i := 0; i < 20; i++ {
		go func() {
			defer wg.Done()
			v, err := c.GetOrLoad(ctx, "hot", time.Minute, func() (string, error) {
				atomic.AddInt32(&calls, 1)
				time.Sleep(20 * time.Millisecond)
				return "loaded", nil
			})
			assert.NoError(t, err)
			assert.Equal(t, "loaded", v)
		}()
	}
	wg.Wait()
	assert.Equal(t, int32(1), atomic.LoadInt32(&calls))

	_, err := c.GetOrLoad(ctx, "empty", time.Minute, func() (string, error) { return "", nil })
	assert.ErrorIs(t, err, ErrNotFound)
	_, err = c.Get(ctx, "empty")
	assert.ErrorIs(t, err, ErrNotFound)

	_, err = c.GetOrLoad(ctx, "fail", time.Minute, func() (string, error) { return "", assert.AnError })
	assert.Error(t, err)
}

func TestLayered_GetOrLoadReadsL2WhenBloomEmpty(t *testing.T) {
	_, rdb := testRedis(t)
	c := NewLayered(rdb)
	ctx := context.Background()
	require.NoError(t, c.L2().Set(ctx, "only-l2", "from-redis", time.Minute))
	c.bloom = NewBloom(1<<16, 4)
	v, err := c.GetOrLoad(ctx, "only-l2", time.Minute, func() (string, error) {
		t.Fatal("loader should not run")
		return "", nil
	})
	require.NoError(t, err)
	assert.Equal(t, "from-redis", v)
}

func TestLayered_JitterTTL(t *testing.T) {
	c := NewLayered(nil)
	c.jitter = 0
	assert.Equal(t, time.Second, c.jitterTTL(time.Second))
	c.jitter = 0.5
	got := c.jitterTTL(time.Second)
	assert.GreaterOrEqual(t, got, time.Second)
	assert.LessOrEqual(t, got, time.Second+time.Duration(0.5*float64(time.Second)))
}
