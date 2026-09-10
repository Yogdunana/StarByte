package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Yogdunana/StarByte/backend/internal/meeting/model"
)

func DueVotes(ctx context.Context, db *gorm.DB, now time.Time) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := db.WithContext(ctx).Model(&model.Vote{}).Where("status=? AND end_time <= ?", model.VoteOpen, now).Order("end_time ASC, id ASC").Limit(200).Pluck("id", &ids).Error
	return ids, err
}
func CloseExpiredVote(ctx context.Context, db *gorm.DB, id uuid.UUID, now time.Time) error {
	return db.WithContext(ctx).Model(&model.Vote{}).Where("id=? AND status=? AND end_time <= ?", id, model.VoteOpen, now).Updates(map[string]interface{}{"status": model.VoteClosed, "updated_at": now}).Error
}
