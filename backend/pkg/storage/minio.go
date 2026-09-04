package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/Yogdunana/StarByte/backend/pkg/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIO 基于官方 SDK 的对象存储实现
type MinIO struct {
	client *minio.Client
	bucket string
}

// NewMinIO 初始化 MinIO 客户端。调用方应再执行 EnsureBucket。
func NewMinIO(cfg config.MinIOConfig) (*MinIO, error) {
	if cfg.Endpoint == "" || cfg.Bucket == "" {
		return nil, fmt.Errorf("MinIO endpoint or bucket not configured")
	}
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("init MinIO client: %w", err)
	}
	return &MinIO{client: client, bucket: cfg.Bucket}, nil
}

// EnsureBucket 确保 bucket 存在，不存在则创建
func (m *MinIO) EnsureBucket(ctx context.Context) error {
	exists, err := m.client.BucketExists(ctx, m.bucket)
	if err != nil {
		return fmt.Errorf("check MinIO bucket: %w", err)
	}
	if exists {
		return nil
	}
	if err := m.client.MakeBucket(ctx, m.bucket, minio.MakeBucketOptions{}); err != nil {
		return fmt.Errorf("create MinIO bucket: %w", err)
	}
	return nil
}

// Upload 上传对象
func (m *MinIO) Upload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) error {
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	_, err := m.client.PutObject(ctx, m.bucket, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("upload to MinIO: %w", err)
	}
	return nil
}

// Download 下载对象，调用方负责关闭 ReadCloser
func (m *MinIO) Download(ctx context.Context, objectName string) (io.ReadCloser, string, error) {
	obj, err := m.client.GetObject(ctx, m.bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, "", fmt.Errorf("download from MinIO: %w", err)
	}
	stat, err := obj.Stat()
	if err != nil {
		_ = obj.Close()
		return nil, "", fmt.Errorf("stat MinIO object: %w", err)
	}
	return obj, stat.ContentType, nil
}

// Delete 删除对象
func (m *MinIO) Delete(ctx context.Context, objectName string) error {
	if err := m.client.RemoveObject(ctx, m.bucket, objectName, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("delete MinIO object: %w", err)
	}
	return nil
}

// List 按前缀列举对象
func (m *MinIO) List(ctx context.Context, prefix string) ([]ObjectInfo, error) {
	var result []ObjectInfo
	for obj := range m.client.ListObjects(ctx, m.bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	}) {
		if obj.Err != nil {
			return nil, fmt.Errorf("list MinIO objects: %w", obj.Err)
		}
		result = append(result, ObjectInfo{
			Name:        obj.Key,
			Size:        obj.Size,
			ContentType: obj.ContentType,
		})
	}
	return result, nil
}

// PresignedURL 生成预签名 GET URL
func (m *MinIO) PresignedURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	if expiry <= 0 {
		expiry = PresignExpiry
	}
	u, err := m.client.PresignedGetObject(ctx, m.bucket, objectName, expiry, url.Values{})
	if err != nil {
		return "", fmt.Errorf("presign MinIO URL: %w", err)
	}
	return u.String(), nil
}
