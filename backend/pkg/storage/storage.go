package storage

import (
	"context"
	"io"
	"time"
)

// PresignExpiry 预签名 URL 默认有效期（1 小时）
const PresignExpiry = time.Hour

// ObjectInfo 对象存储中的文件元数据
type ObjectInfo struct {
	Name        string
	Size        int64
	ContentType string
}

// ObjectStorage 对象存储抽象，供文件模块与后续审计归档复用
type ObjectStorage interface {
	EnsureBucket(ctx context.Context) error
	Upload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) error
	Download(ctx context.Context, objectName string) (io.ReadCloser, string, error)
	Delete(ctx context.Context, objectName string) error
	List(ctx context.Context, prefix string) ([]ObjectInfo, error)
	PresignedURL(ctx context.Context, objectName string, expiry time.Duration) (string, error)
}
