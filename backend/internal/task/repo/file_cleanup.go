package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/Yogdunana/StarByte/backend/internal/task/model"
)

type CleanupRepo interface {
	Create(context.Context, *model.FileCleanup) error
	Pending(context.Context, time.Time) ([]uuid.UUID, error)
	Lock(context.Context, uuid.UUID) (*model.FileCleanup, error)
	Save(context.Context, *model.FileCleanup) error
	OtherReferences(context.Context, uuid.UUID, uuid.UUID) (int64, error)
}
type cleanupRepo struct{ db *gorm.DB }

func NewCleanupRepo(db *gorm.DB) CleanupRepo { return &cleanupRepo{db} }
func (r *cleanupRepo) Create(ctx context.Context, row *model.FileCleanup) error {
	return r.db.WithContext(ctx).Create(row).Error
}
func (r *cleanupRepo) Save(ctx context.Context, row *model.FileCleanup) error {
	return r.db.WithContext(ctx).Save(row).Error
}
func (r *cleanupRepo) Pending(ctx context.Context, now time.Time) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.db.WithContext(ctx).Model(&model.FileCleanup{}).Where("status='pending' AND next_attempt_at<=?", now).Order("next_attempt_at,id").Limit(20).Pluck("id", &ids).Error
	return ids, err
}
func (r *cleanupRepo) Lock(ctx context.Context, id uuid.UUID) (*model.FileCleanup, error) {
	var row model.FileCleanup
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).First(&row, "id=?", id).Error
	return &row, err
}
func (r *cleanupRepo) OtherReferences(ctx context.Context, fileID, attachmentID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.TaskAttachment{}).Where("file_id=? AND id<>?", fileID, attachmentID).Count(&count).Error
	return count, err
}
