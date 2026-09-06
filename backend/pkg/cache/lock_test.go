package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLock_AcquireReentrantWatchdog(t *testing.T) {
	_, rdb := testRedis(t)
	ctx := context.Background()

	_, err := Acquire(ctx, rdb, "n", "o", 0)
	require.NoError(t, err)

	lk, err := Acquire(ctx, rdb, "job", "owner-a", time.Second)
	require.NoError(t, err)
	assert.Contains(t, lk.String(), "lock:")
	assert.NotEmpty(t, lk.Token())

	busy, err := Acquire(ctx, rdb, "job", "other", time.Second)
	assert.Nil(t, busy)
	assert.ErrorIs(t, err, ErrLockBusy)

	require.NoError(t, lk.Reenter(ctx))
	require.NoError(t, lk.Unlock(ctx))
	still, err := rdb.Exists(ctx, lockKey("job")).Result()
	require.NoError(t, err)
	assert.Equal(t, int64(1), still)
	require.NoError(t, lk.Unlock(ctx))

	lk, err = Acquire(ctx, rdb, "wd", "o1", 400*time.Millisecond)
	require.NoError(t, err)
	lk.StartWatchdog()
	lk.StartWatchdog()
	time.Sleep(550 * time.Millisecond)
	ok, err := rdb.Exists(ctx, lockKey("wd")).Result()
	require.NoError(t, err)
	assert.Equal(t, int64(1), ok)
	require.NoError(t, lk.Refresh(ctx))
	lk.StopWatchdog()
	require.NoError(t, lk.Unlock(ctx))
	require.ErrorIs(t, lk.Unlock(ctx), ErrLockNotHeld)
}

func TestLock_FairFIFO(t *testing.T) {
	_, rdb := testRedis(t)
	ctx := context.Background()
	held, err := Acquire(ctx, rdb, "fair", "first", time.Second)
	require.NoError(t, err)

	done := make(chan error, 1)
	go func() {
		_, aerr := AcquireFair(ctx, rdb, "fair", "second", time.Second, 800*time.Millisecond)
		done <- aerr
	}()
	time.Sleep(80 * time.Millisecond)
	require.NoError(t, held.Unlock(ctx))
	select {
	case aerr := <-done:
		require.NoError(t, aerr)
	case <-time.After(time.Second):
		t.Fatal("fair lock did not acquire")
	}

	_, err = AcquireFair(ctx, rdb, "busy-fair", "a", 2*time.Second, 200*time.Millisecond)
	require.NoError(t, err)
	_, err = AcquireFair(ctx, rdb, "busy-fair", "b", time.Second, 80*time.Millisecond)
	assert.ErrorIs(t, err, ErrLockBusy)
}

func TestLock_TwoHandlesSameOwner(t *testing.T) {
	_, rdb := testRedis(t)
	ctx := context.Background()
	a, err := Acquire(ctx, rdb, "same", "o", time.Second)
	require.NoError(t, err)
	b, err := Acquire(ctx, rdb, "same", "o", time.Second)
	require.NoError(t, err)
	require.NoError(t, a.Unlock(ctx))
	n, err := rdb.Exists(ctx, lockKey("same")).Result()
	require.NoError(t, err)
	assert.Equal(t, int64(1), n)
	require.NoError(t, b.Unlock(ctx))
	n, err = rdb.Exists(ctx, lockKey("same")).Result()
	require.NoError(t, err)
	assert.Equal(t, int64(0), n)
}

func TestLock_FairReclaimsStaleHead(t *testing.T) {
	_, rdb := testRedis(t)
	ctx := context.Background()
	queue := lockKey("stale") + ":wait"
	require.NoError(t, rdb.RPush(ctx, queue, "dead-ticket").Err())
	lk, err := AcquireFair(ctx, rdb, "stale", "o", time.Second, 400*time.Millisecond)
	require.NoError(t, err)
	require.NoError(t, lk.Unlock(ctx))
}

func TestLock_FairCanceledDoesNotLeak(t *testing.T) {
	_, rdb := testRedis(t)
	ctx, cancel := context.WithCancel(context.Background())
	held, err := Acquire(context.Background(), rdb, "cx", "first", time.Second)
	require.NoError(t, err)

	done := make(chan error, 1)
	go func() {
		_, aerr := AcquireFair(ctx, rdb, "cx", "second", time.Second, 800*time.Millisecond)
		done <- aerr
	}()
	time.Sleep(80 * time.Millisecond)
	cancel()
	require.NoError(t, held.Unlock(context.Background()))
	select {
	case aerr := <-done:
		assert.Error(t, aerr)
	case <-time.After(time.Second):
		t.Fatal("fair lock did not return after cancel")
	}
	n, err := rdb.Exists(context.Background(), lockKey("cx")).Result()
	require.NoError(t, err)
	assert.Equal(t, int64(0), n)
}

func TestLock_ReenterUnlocked(t *testing.T) {
	_, rdb := testRedis(t)
	lk, err := Acquire(context.Background(), rdb, "x", "o", time.Second)
	require.NoError(t, err)
	require.NoError(t, lk.Unlock(context.Background()))
	assert.ErrorIs(t, lk.Reenter(context.Background()), ErrLockNotHeld)
}
