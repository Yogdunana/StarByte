package repo

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/model"
)

type RuntimeRepo interface {
	LockInstance(context.Context, uuid.UUID) error
	LockDefinition(context.Context, uuid.UUID) (*model.FlowDefinition, error)
	ActiveUser(context.Context, uuid.UUID) (bool, error)
}
type runtimeRepo struct{ db *gorm.DB }

func NewRuntimeRepo(db *gorm.DB) RuntimeRepo { return &runtimeRepo{db} }
func (r *runtimeRepo) LockInstance(ctx context.Context, id uuid.UUID) error {
	var instance model.FlowInstance
	return r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Select("id").Where("id=?", id).Take(&instance).Error
}

func (r *runtimeRepo) ActiveUser(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("users").Where("id = ? AND status = 0 AND deleted_at IS NULL", id).Count(&count).Error
	return count == 1, err
}

func (r *runtimeRepo) LockDefinition(ctx context.Context, id uuid.UUID) (*model.FlowDefinition, error) {
	var definition model.FlowDefinition
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).Take(&definition).Error
	return &definition, err
}
