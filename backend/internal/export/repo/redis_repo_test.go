package repo

import (
	"context"
	"testing"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/export/model"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedisRepo_TaskAndFile(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	r := NewRedisRepo(rdb)
	ctx := context.Background()

	task := &model.ExportTask{ID: "t1", Status: model.StatusPending, Filename: "a.csv"}
	require.NoError(t, r.SaveTask(ctx, task))
	got, err := r.GetTask(ctx, "t1")
	require.NoError(t, err)
	assert.Equal(t, "a.csv", got.Filename)

	_, err = r.GetTask(ctx, "missing")
	assert.ErrorIs(t, err, ErrNotFound)

	meta := &model.FileMeta{FileID: "f1", Filename: "a.csv", ExpiresAt: time.Now().Add(time.Minute)}
	require.NoError(t, r.SaveFile(ctx, meta))
	require.NoError(t, r.SaveBlob(ctx, "f1", []byte("hello")))
	fm, err := r.GetFile(ctx, "f1")
	require.NoError(t, err)
	assert.Equal(t, "a.csv", fm.Filename)
	blob, err := r.GetBlob(ctx, "f1")
	require.NoError(t, err)
	assert.Equal(t, "hello", string(blob))
}

func TestRedisRepo_NilClient(t *testing.T) {
	r := NewRedisRepo(nil)
	err := r.SaveTask(context.Background(), &model.ExportTask{ID: "x"})
	assert.Error(t, err)
	_, err = r.GetFile(context.Background(), "f")
	assert.Error(t, err)
	_, err = r.GetBlob(context.Background(), "f")
	assert.Error(t, err)
	_, err = r.GetTask(context.Background(), "x")
	assert.Error(t, err)
}

func TestRedisRepo_MissingFile(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	r := NewRedisRepo(rdb)
	_, err := r.GetFile(context.Background(), "missing")
	assert.ErrorIs(t, err, ErrNotFound)
	_, err = r.GetBlob(context.Background(), "missing")
	assert.ErrorIs(t, err, ErrNotFound)
}
