package repo

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/Yogdunana/StarByte/backend/internal/meeting/model"
)

func LockMeeting(ctx context.Context, db *gorm.DB, id uuid.UUID) error {
	var meeting model.Meeting
	return db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("id=?", id).Take(&meeting).Error
}

// LockVote always locks its parent first, matching meeting lifecycle mutations.
func LockVote(ctx context.Context, db *gorm.DB, id uuid.UUID) error {
	var vote model.Vote
	if err := db.WithContext(ctx).Select("meeting_id").Where("id=?", id).Take(&vote).Error; err != nil {
		return err
	}
	if err := LockMeeting(ctx, db, vote.MeetingID); err != nil {
		return err
	}
	return db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("id=?", id).Take(&vote).Error
}
