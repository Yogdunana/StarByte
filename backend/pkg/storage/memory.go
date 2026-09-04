package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

// Memory 内存对象存储，供单测使用
type Memory struct {
	mu      sync.RWMutex
	bucket  string
	objects map[string]memoryObject
}

type memoryObject struct {
	data        []byte
	contentType string
}

// NewMemory 创建内存存储
func NewMemory(bucket string) *Memory {
	if bucket == "" {
		bucket = "starbyte"
	}
	return &Memory{bucket: bucket, objects: make(map[string]memoryObject)}
}

// EnsureBucket 内存实现无需创建 bucket
func (m *Memory) EnsureBucket(_ context.Context) error {
	return nil
}

// Upload 写入内存对象
func (m *Memory) Upload(_ context.Context, objectName string, reader io.Reader, _ int64, contentType string) error {
	data, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("read upload body: %w", err)
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.objects[objectName] = memoryObject{data: data, contentType: contentType}
	return nil
}

// Download 读取内存对象
func (m *Memory) Download(_ context.Context, objectName string) (io.ReadCloser, string, error) {
	m.mu.RLock()
	obj, ok := m.objects[objectName]
	m.mu.RUnlock()
	if !ok {
		return nil, "", fmt.Errorf("object not found: %s", objectName)
	}
	return io.NopCloser(bytes.NewReader(obj.data)), obj.contentType, nil
}

// Delete 删除内存对象
func (m *Memory) Delete(_ context.Context, objectName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.objects, objectName)
	return nil
}

// List 按前缀列举
func (m *Memory) List(_ context.Context, prefix string) ([]ObjectInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []ObjectInfo
	for name, obj := range m.objects {
		if prefix != "" && !strings.HasPrefix(name, prefix) {
			continue
		}
		result = append(result, ObjectInfo{
			Name:        name,
			Size:        int64(len(obj.data)),
			ContentType: obj.contentType,
		})
	}
	return result, nil
}

// PresignedURL 返回可识别的内存占位 URL
func (m *Memory) PresignedURL(_ context.Context, objectName string, expiry time.Duration) (string, error) {
	if expiry <= 0 {
		expiry = PresignExpiry
	}
	return fmt.Sprintf("memory://%s/%s?expires=%d", m.bucket, objectName, int(expiry.Seconds())), nil
}
