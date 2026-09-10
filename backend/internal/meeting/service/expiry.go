package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/Yogdunana/StarByte/backend/internal/meeting/repo"
)

// NewVoteExpiryJob is registered only by the server's scheduler. It has no
// request-controlled user or meeting scope and exposes no public HTTP endpoint.
func NewVoteExpiryJob(db *gorm.DB) func(context.Context, string, func(string)) error {
	return func(ctx context.Context, _ string, logf func(string)) error {
		now := time.Now()
		ids, err := repo.DueVotes(ctx, db, now)
		if err != nil {
			return fmt.Errorf("list expired votes: %w", err)
		}
		for _, id := range ids {
			if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
				if err := repo.LockVote(ctx, tx, id); err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						return nil
					}
					return err
				}
				return repo.CloseExpiredVote(ctx, tx, id, now)
			}); err != nil {
				return fmt.Errorf("close expired vote: %w", err)
			}
		}
		logf(fmt.Sprintf("检查到期投票 %d 项", len(ids)))
		return nil
	}
}
