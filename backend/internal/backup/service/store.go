package service

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/Yogdunana/StarByte/backend/internal/backup/model"
	"github.com/Yogdunana/StarByte/backend/pkg/storage"
)

// ArtifactStore persists gzip objects to MinIO and/or a local fallback path.
type ArtifactStore interface {
	Put(ctx context.Context, key string, r io.Reader, size int64) (kind string, err error)
	Get(ctx context.Context, kind, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, kind, key string) error
}

type artifactStore struct {
	objects storage.ObjectStorage
	local   string
}

func newArtifactStore(objects storage.ObjectStorage, localPath string) ArtifactStore {
	return &artifactStore{objects: objects, local: strings.TrimSpace(localPath)}
}

func (s *artifactStore) Put(ctx context.Context, key string, r io.Reader, size int64) (string, error) {
	var minioErr error
	if s.objects != nil {
		minioErr = s.objects.Upload(ctx, key, r, size, "application/gzip")
		if minioErr == nil {
			return model.StorageMinIO, nil
		}
	}
	if s.local == "" {
		if minioErr != nil {
			return "", minioErr
		}
		return "", fmt.Errorf("MinIO 未配置且未设置本地回退路径")
	}
	if seeker, ok := r.(io.Seeker); ok {
		if _, err := seeker.Seek(0, io.SeekStart); err != nil {
			return "", err
		}
	}
	full := filepath.Join(s.local, filepath.FromSlash(key))
	if err := os.MkdirAll(filepath.Dir(full), 0o750); err != nil {
		return "", err
	}
	f, err := os.OpenFile(full, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := io.Copy(f, r); err != nil {
		return "", err
	}
	return model.StorageLocal, nil
}

func (s *artifactStore) Get(ctx context.Context, kind, key string) (io.ReadCloser, error) {
	if kind == model.StorageMinIO && s.objects != nil {
		rc, _, err := s.objects.Download(ctx, key)
		return rc, err
	}
	if kind == model.StorageLocal || s.local != "" {
		full := filepath.Join(s.local, filepath.FromSlash(key))
		return os.Open(full)
	}
	return nil, fmt.Errorf("无法读取备份对象")
}

func (s *artifactStore) Delete(ctx context.Context, kind, key string) error {
	var first error
	if (kind == model.StorageMinIO || kind == "") && s.objects != nil && key != "" {
		if err := s.objects.Delete(ctx, key); err != nil {
			first = err
		}
	}
	if s.local != "" && key != "" {
		full := filepath.Join(s.local, filepath.FromSlash(key))
		if err := os.Remove(full); err != nil && !os.IsNotExist(err) && first == nil {
			first = err
		}
	}
	return first
}

func objectKey(prefix string, recID string, filename string) string {
	prefix = strings.Trim(prefix, "/")
	if prefix == "" {
		prefix = "backups"
	}
	return path.Join(prefix, "postgres", recID, filename)
}
