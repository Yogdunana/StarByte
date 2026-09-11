package repo

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/Yogdunana/StarByte/backend/internal/task/model"
)

func LockTask(ctx context.Context, db *gorm.DB, id uuid.UUID) error {
	return db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Select("id").First(&model.Task{}, "id = ?", id).Error
}
