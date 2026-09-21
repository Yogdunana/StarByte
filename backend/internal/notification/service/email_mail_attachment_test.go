package service

import (
	"bytes"
	"context"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/Yogdunana/StarByte/backend/pkg/storage"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeObjectStorage records Download calls so we can prove the query phase
// rejected inaccessible files before any object-store read happened.
type fakeObjectStorage struct {
	mu        sync.Mutex
	downloads []string
	data      map[string][]byte
}

func (f *fakeObjectStorage) EnsureBucket(context.Context) error { return nil }
func (f *fakeObjectStorage) Upload(context.Context, string, io.Reader, int64, string) error {
	return nil
}
func (f *fakeObjectStorage) Download(_ context.Context, name string) (io.ReadCloser, string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.downloads = append(f.downloads, name)
	if d, ok := f.data[name]; ok {
		return io.NopCloser(bytes.NewReader(d)), "application/octet-stream", nil
	}
	return io.NopCloser(bytes.NewReader(nil)), "", nil
}
func (f *fakeObjectStorage) Delete(context.Context, string) error { return nil }
func (f *fakeObjectStorage) List(context.Context, string) ([]storage.ObjectInfo, error) {
	return nil, nil
}
func (f *fakeObjectStorage) PresignedURL(context.Context, string, time.Duration) (string, error) {
	return "", nil
}

// fakeFileRepo mirrors the production SQL scope: a file is accessible only when
// it is uploaded by operatorID or is public.
type fakeFileRepo struct {
	rows []attachmentFileRow
}

func (f *fakeFileRepo) accessibleFiles(_ context.Context, operatorID uuid.UUID, ids []uuid.UUID) ([]attachmentFileRow, error) {
	out := make([]attachmentFileRow, 0, len(ids))
	for _, id := range ids {
		for _, r := range f.rows {
			if r.ID != id {
				continue
			}
			if r.UploadedBy == operatorID || r.IsPublic {
				out = append(out, r)
			}
		}
	}
	return out, nil
}

func newLoader(repo fileRepository, store storage.ObjectStorage) *attachmentLoader {
	return &attachmentLoader{repo: repo, store: store}
}

func TestAttachmentLoadOwnerCanAccess(t *testing.T) {
	owner := uuid.New()
	fid := uuid.New()
	repo := &fakeFileRepo{rows: []attachmentFileRow{
		{ID: fid, OriginalName: "a.pdf", Name: "stored.pdf", Path: "p/a.pdf", MimeType: "application/pdf", Size: 10, UploadedBy: owner, IsPublic: false},
	}}
	store := &fakeObjectStorage{data: map[string][]byte{"p/a.pdf": []byte("hello")}}
	l := newLoader(repo, store)

	got, err := l.Load(context.Background(), owner, []uuid.UUID{fid})
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "a.pdf", got[0].Name)
	assert.Equal(t, []byte("hello"), got[0].Data)
	assert.Equal(t, []string{"p/a.pdf"}, store.downloads)
}

func TestAttachmentLoadOtherPrivateRejected(t *testing.T) {
	owner := uuid.New()
	other := uuid.New()
	fid := uuid.New()
	repo := &fakeFileRepo{rows: []attachmentFileRow{
		{ID: fid, OriginalName: "secret.pdf", Name: "x.pdf", Path: "p/secret.pdf", MimeType: "application/pdf", Size: 10, UploadedBy: other, IsPublic: false},
	}}
	store := &fakeObjectStorage{data: map[string][]byte{"p/secret.pdf": []byte("topsecret")}}
	l := newLoader(repo, store)

	_, err := l.Load(context.Background(), owner, []uuid.UUID{fid})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "attachment not found")
	assert.Empty(t, store.downloads, "Download must not be called for an inaccessible file")
}

func TestAttachmentLoadPublicAllowed(t *testing.T) {
	owner := uuid.New()
	other := uuid.New()
	fid := uuid.New()
	repo := &fakeFileRepo{rows: []attachmentFileRow{
		{ID: fid, OriginalName: "pub.pdf", Name: "pub.pdf", Path: "p/pub.pdf", MimeType: "application/pdf", Size: 10, UploadedBy: other, IsPublic: true},
	}}
	store := &fakeObjectStorage{data: map[string][]byte{"p/pub.pdf": []byte("public")}}
	l := newLoader(repo, store)

	got, err := l.Load(context.Background(), owner, []uuid.UUID{fid})
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, []string{"p/pub.pdf"}, store.downloads)
}

func TestAttachmentLoadMixedBatchRejected(t *testing.T) {
	owner := uuid.New()
	other := uuid.New()
	mine := uuid.New()
	theirs := uuid.New()
	repo := &fakeFileRepo{rows: []attachmentFileRow{
		{ID: mine, OriginalName: "mine.pdf", Path: "p/mine.pdf", Size: 1, UploadedBy: owner, IsPublic: false},
		{ID: theirs, OriginalName: "theirs.pdf", Path: "p/theirs.pdf", Size: 1, UploadedBy: other, IsPublic: false},
	}}
	store := &fakeObjectStorage{}
	l := newLoader(repo, store)

	_, err := l.Load(context.Background(), owner, []uuid.UUID{mine, theirs})
	require.Error(t, err, "one inaccessible id must reject the whole batch")
	assert.Empty(t, store.downloads, "no Download when any id is inaccessible")
}

func TestWorkerRejectsAttachmentWhenUserIDNil(t *testing.T) {
	sender := &stubMIME{}
	logs := newMemLogs()
	store := &fakeObjectStorage{}
	repo := &fakeFileRepo{}
	l := newLoader(repo, store)
	w := NewEmailWorker(sender, logs, l, newMinuteLimiter(50, time.Now))
	ctx := context.Background()
	_, err := w.Enqueue(ctx, MailJob{
		To:            []string{"a@b.c"},
		Subject:       "s",
		Body:          "b",
		AttachmentIDs: []uuid.UUID{uuid.New()},
		UserID:        nil,
	})
	require.NoError(t, err)
	processQueued(t, w, ctx)
	assert.Equal(t, 0, sender.calls, "mail must not be sent when operator is unknown (fail-closed)")
}
