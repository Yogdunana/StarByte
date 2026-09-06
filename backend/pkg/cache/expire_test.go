package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestStore_SubscribeExpired(t *testing.T) {
	_, rdb := testRedis(t)
	s := NewStore(rdb)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	require.NoError(t, s.SubscribeExpired(ctx, func(string) {}))
	time.Sleep(20 * time.Millisecond)
	require.NoError(t, NewStore(nil).SubscribeExpired(ctx, nil))
}
