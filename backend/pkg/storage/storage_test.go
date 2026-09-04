package storage

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemory_CRUDAndPresign(t *testing.T) {
	ctx := context.Background()
	store := NewMemory("starbyte")
	require.NoError(t, store.EnsureBucket(ctx))

	payload := []byte("hello-minio")
	require.NoError(t, store.Upload(ctx, "image/a.txt", bytes.NewReader(payload), int64(len(payload)), "text/plain"))

	rc, ctype, err := store.Download(ctx, "image/a.txt")
	require.NoError(t, err)
	defer rc.Close()
	got, err := io.ReadAll(rc)
	require.NoError(t, err)
	assert.Equal(t, payload, got)
	assert.Equal(t, "text/plain", ctype)

	list, err := store.List(ctx, "image/")
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "image/a.txt", list[0].Name)

	url, err := store.PresignedURL(ctx, "image/a.txt", PresignExpiry)
	require.NoError(t, err)
	assert.Contains(t, url, "image/a.txt")
	assert.Contains(t, url, "expires=3600")

	require.NoError(t, store.Delete(ctx, "image/a.txt"))
	_, _, err = store.Download(ctx, "image/a.txt")
	assert.Error(t, err)
}
