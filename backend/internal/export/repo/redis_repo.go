package repo

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/export/model"
	"github.com/redis/go-redis/v9"
)

var ErrNotFound = errors.New("export: not found")

// ExportRepo persists export tasks and temporary files in Redis.
type ExportRepo interface {
	SaveTask(ctx context.Context, task *model.ExportTask) error
	GetTask(ctx context.Context, id string) (*model.ExportTask, error)
	SaveFile(ctx context.Context, meta *model.FileMeta) error
	GetFile(ctx context.Context, id string) (*model.FileMeta, error)
	SaveBlob(ctx context.Context, fileID string, data []byte) error
	GetBlob(ctx context.Context, fileID string) ([]byte, error)
}

type redisRepo struct {
	rdb *redis.Client
}

func NewRedisRepo(rdb *redis.Client) ExportRepo {
	return &redisRepo{rdb: rdb}
}

func (r *redisRepo) SaveTask(ctx context.Context, task *model.ExportTask) error {
	if r.rdb == nil {
		return errors.New("redis unavailable")
	}
	raw, err := json.Marshal(task)
	if err != nil {
		return err
	}
	return r.rdb.Set(ctx, model.TaskKey(task.ID), raw, model.TaskTTL).Err()
}

func (r *redisRepo) GetTask(ctx context.Context, id string) (*model.ExportTask, error) {
	if r.rdb == nil {
		return nil, errors.New("redis unavailable")
	}
	raw, err := r.rdb.Get(ctx, model.TaskKey(id)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	var task model.ExportTask
	if err := json.Unmarshal(raw, &task); err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *redisRepo) SaveFile(ctx context.Context, meta *model.FileMeta) error {
	if r.rdb == nil {
		return errors.New("redis unavailable")
	}
	raw, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	ttl := time.Until(meta.ExpiresAt)
	if ttl <= 0 {
		ttl = model.FileTTL
	}
	return r.rdb.Set(ctx, model.FileKey(meta.FileID), raw, ttl).Err()
}

func (r *redisRepo) GetFile(ctx context.Context, id string) (*model.FileMeta, error) {
	if r.rdb == nil {
		return nil, errors.New("redis unavailable")
	}
	raw, err := r.rdb.Get(ctx, model.FileKey(id)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	var meta model.FileMeta
	if err := json.Unmarshal(raw, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

func (r *redisRepo) SaveBlob(ctx context.Context, fileID string, data []byte) error {
	if r.rdb == nil {
		return errors.New("redis unavailable")
	}
	return r.rdb.Set(ctx, model.BlobKey(fileID), data, model.FileTTL).Err()
}

func (r *redisRepo) GetBlob(ctx context.Context, fileID string) ([]byte, error) {
	if r.rdb == nil {
		return nil, errors.New("redis unavailable")
	}
	raw, err := r.rdb.Get(ctx, model.BlobKey(fileID)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return raw, nil
}
