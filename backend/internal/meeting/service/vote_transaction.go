package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Yogdunana/StarByte/backend/internal/meeting/dto"
	"github.com/Yogdunana/StarByte/backend/internal/meeting/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

func (s *meetingService) voteTransaction(ctx context.Context, id uuid.UUID, fn func(*meetingService) error) error {
	if s.db == nil {
		return fn(s)
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := repo.LockVote(ctx, tx, id); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return response.NewError(response.CodeVoteNotFound, "投票不存在")
			}
			return err
		}
		return fn(s.bind(tx))
	})
}
func (s *meetingService) CastVote(ctx context.Context, id, user uuid.UUID, option string) error {
	return s.voteTransaction(ctx, id, func(bound *meetingService) error { return bound.castVote(ctx, id, user, option) })
}
func (s *meetingService) CloseVote(ctx context.Context, id uuid.UUID) (*dto.VoteResponse, error) {
	var result *dto.VoteResponse
	err := s.voteTransaction(ctx, id, func(bound *meetingService) error { var err error; result, err = bound.closeVote(ctx, id); return err })
	return result, err
}
