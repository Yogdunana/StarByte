package configstore

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore_GetSetGetBool(t *testing.T) {
	ctx := context.Background()
	store := New(nil, NewMemoryBackend())

	_, err := store.Get(ctx, "missing")
	assert.ErrorIs(t, err, ErrNotFound)

	require.NoError(t, store.Set(ctx, "site.name", "计协"))
	got, err := store.Get(ctx, "site.name")
	require.NoError(t, err)
	assert.Equal(t, "计协", got)

	require.NoError(t, store.Set(ctx, "flag", "true"))
	ok, err := store.GetBool(ctx, "flag")
	require.NoError(t, err)
	assert.True(t, ok)

	require.NoError(t, store.Set(ctx, "flag", "0"))
	ok, err = store.GetBool(ctx, "flag")
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestStore_EmptyKey(t *testing.T) {
	ctx := context.Background()
	store := New(nil, NewMemoryBackend())
	_, err := store.Get(ctx, "")
	assert.Error(t, err)
	assert.Error(t, store.Set(ctx, "", "x"))
}

func TestParseBool(t *testing.T) {
	ok, err := parseBool("yes")
	require.NoError(t, err)
	assert.True(t, ok)
	ok, err = parseBool("off")
	require.NoError(t, err)
	assert.False(t, ok)
	_, err = parseBool("maybe")
	assert.Error(t, err)
}
