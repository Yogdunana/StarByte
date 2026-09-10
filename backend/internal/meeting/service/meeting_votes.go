package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/Yogdunana/StarByte/backend/internal/meeting/model"
)

// The parent meeting lock prevents new votes/casts while all open polls close.
func (s *meetingService) closeMeetingVotes(ctx context.Context, id uuid.UUID) error {
	votes, err := s.votes.ListByMeeting(ctx, id)
	if err != nil {
		return fmt.Errorf("list votes to close: %w", err)
	}
	now := time.Now()
	for _, v := range votes {
		if v.Status != model.VoteOpen && v.Status != model.VotePending {
			continue
		}
		v.Status = model.VoteClosed
		if v.EndTime == nil || v.EndTime.After(now) {
			v.EndTime = &now
		}
		v.UpdatedAt = now
		if err := s.votes.UpdateVote(ctx, &v); err != nil {
			return fmt.Errorf("close meeting vote: %w", err)
		}
	}
	return nil
}
