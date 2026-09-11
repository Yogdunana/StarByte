package repo

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCASStore(t *testing.T) (CASTicketStore, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	return NewCASStore(rdb), mr
}

func TestCASStore_StateRoundTrip(t *testing.T) {
	store, _ := newCASStore(t)
	ctx := context.Background()

	require.NoError(t, store.PutState(ctx, "st1", `{"redirect":"/tasks"}`, time.Minute))
	got, ok, err := store.TakeState(ctx, "st1")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, `{"redirect":"/tasks"}`, got)

	_, ok, err = store.TakeState(ctx, "st1")
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestCASStore_CodeRoundTrip(t *testing.T) {
	store, _ := newCASStore(t)
	ctx := context.Background()

	require.NoError(t, store.PutCode(ctx, "c1", []byte(`{"ok":true}`), time.Minute))
	got, err := store.TakeCode(ctx, "c1")
	require.NoError(t, err)
	assert.Equal(t, []byte(`{"ok":true}`), got)

	_, err = store.TakeCode(ctx, "c1")
	require.ErrorIs(t, err, redis.Nil)
}

func TestCASStore_Unavailable(t *testing.T) {
	ctx := context.Background()
	store := NewCASStore(nil)

	require.Error(t, store.PutState(ctx, "st", "x", time.Minute))
	_, ok, err := store.TakeState(ctx, "st")
	require.NoError(t, err)
	assert.False(t, ok)

	require.Error(t, store.PutCode(ctx, "c", []byte("x"), time.Minute))
	_, err = store.TakeCode(ctx, "c")
	require.ErrorIs(t, err, redis.Nil)

	require.Error(t, store.PutState(ctx, "", "x", time.Minute))
	_, err = store.TakeCode(ctx, "")
	require.ErrorIs(t, err, redis.Nil)
}
