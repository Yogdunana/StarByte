package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/export/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/Yogdunana/StarByte/backend/pkg/storage"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExport_WithObjectStore(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	store := &memStore{objects: map[string][]byte{}}
	svc := NewExportService(repo.NewRedisRepo(rdb), store, nil)
	out, err := svc.ExportTable(context.Background(), "json", testUserID, sampleReq(1))
	require.NoError(t, err)
	meta, err := svc.Download(context.Background(), out.FileID, testUserID, false, false)
	require.NoError(t, err)
	assert.Empty(t, meta.Bytes)
	assert.Empty(t, meta.URL)
	assert.Equal(t, "members.json", meta.Filename)
	assert.Equal(t, 0, store.downloadCalls())
	dl, err := svc.Download(context.Background(), out.FileID, testUserID, false, true)
	require.NoError(t, err)
	assert.Empty(t, dl.URL)
	assert.NotEmpty(t, dl.Bytes)
	assert.True(t, json.Valid(dl.Bytes))
	assert.Equal(t, 1, store.downloadCalls())
}

func TestExport_StoreDownloadErrorNotExpired(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	store := &memStore{objects: map[string][]byte{}}
	svc := NewExportService(repo.NewRedisRepo(rdb), store, nil)
	out, err := svc.ExportTable(context.Background(), "json", testUserID, sampleReq(1))
	require.NoError(t, err)
	store.setDownloadErr(errors.New("minio unavailable"))
	_, err = svc.Download(context.Background(), out.FileID, testUserID, false, true)
	require.Error(t, err)
	assert.Equal(t, response.CodeInternalError, err.(*response.AppError).Code)
}

func TestExport_OwnerAccess(t *testing.T) {
	svc, _ := newTestSvc(t)
	ctx := context.Background()
	out, err := svc.ExportTable(ctx, "json", "owner", sampleReq(1))
	require.NoError(t, err)

	_, err = svc.GetTask(ctx, out.TaskID, "other", false)
	require.Error(t, err)
	assert.Equal(t, response.CodeForbidden, err.(*response.AppError).Code)

	got, err := svc.GetTask(ctx, out.TaskID, "owner", false)
	require.NoError(t, err)
	assert.Equal(t, out.TaskID, got.TaskID)

	got, err = svc.GetTask(ctx, out.TaskID, "other", true)
	require.NoError(t, err)
	assert.Equal(t, out.FileID, got.FileID)

	_, err = svc.Download(ctx, out.FileID, "other", false, true)
	require.Error(t, err)
	assert.Equal(t, response.CodeForbidden, err.(*response.AppError).Code)

	dl, err := svc.Download(ctx, out.FileID, "owner", false, true)
	require.NoError(t, err)
	assert.NotEmpty(t, dl.Bytes)

	dl, err = svc.Download(ctx, out.FileID, "admin", true, false)
	require.NoError(t, err)
	assert.Empty(t, dl.Bytes)
	assert.Equal(t, "members.json", dl.Filename)
}

func TestCanAccessExport(t *testing.T) {
	assert.True(t, canAccessExport("a", "a", false))
	assert.False(t, canAccessExport("a", "b", false))
	assert.True(t, canAccessExport("a", "b", true))
	assert.False(t, canAccessExport("", "a", false))
	assert.True(t, canAccessExport("", "", true))
}

type memStore struct {
	mu          sync.Mutex
	objects     map[string][]byte
	downloads   int
	downloadErr error
}

func (m *memStore) EnsureBucket(context.Context) error { return nil }
func (m *memStore) Upload(_ context.Context, objectName string, reader io.Reader, _ int64, _ string) error {
	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.objects[objectName] = data
	return nil
}
func (m *memStore) Download(_ context.Context, objectName string) (io.ReadCloser, string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.downloads++
	if m.downloadErr != nil {
		return nil, "", m.downloadErr
	}
	data, ok := m.objects[objectName]
	if !ok {
		return nil, "", io.EOF
	}
	copied := append([]byte(nil), data...)
	return io.NopCloser(bytes.NewReader(copied)), "application/octet-stream", nil
}
func (m *memStore) downloadCalls() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.downloads
}
func (m *memStore) setDownloadErr(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.downloadErr = err
}
func (m *memStore) Delete(context.Context, string) error { return nil }
func (m *memStore) List(context.Context, string) ([]storage.ObjectInfo, error) {
	return nil, nil
}
func (m *memStore) PresignedURL(_ context.Context, objectName string, _ time.Duration) (string, error) {
	return "http://minio/" + objectName, nil
}
