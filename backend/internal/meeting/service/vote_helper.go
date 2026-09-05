package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/meeting/dto"
	"github.com/Yogdunana/StarByte/backend/internal/meeting/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
)

func (s *meetingService) loadWeight(ctx context.Context) (model.VoteWeightConfig, error) {
	row, err := s.votes.GetConfig(ctx, model.WeightConfigKey)
	if err != nil {
		return model.VoteWeightConfig{}, fmt.Errorf("get weight config: %w", err)
	}
	if row == nil {
		return DefaultWeightConfig(), nil
	}
	return parseWeightConfig(row.ConfigValue), nil
}

func (s *meetingService) mustVote(ctx context.Context, id uuid.UUID) (*model.Vote, error) {
	v, err := s.votes.GetVote(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get vote: %w", err)
	}
	if v == nil {
		return nil, response.NewError(response.CodeVoteNotFound, "投票不存在")
	}
	return v, nil
}

func (s *meetingService) ensureExpiredClosed(ctx context.Context, id uuid.UUID) (*model.Vote, error) {
	v, err := s.mustVote(ctx, id)
	if err != nil {
		return nil, err
	}
	if v.Status == model.VoteOpen && v.EndTime != nil && !v.EndTime.After(time.Now()) {
		v.Status = model.VoteClosed
		v.UpdatedAt = time.Now()
		if err := s.votes.UpdateVote(ctx, v); err != nil {
			return nil, fmt.Errorf("auto close vote: %w", err)
		}
	}
	return v, nil
}

func (s *meetingService) ensureVoteOpen(ctx context.Context, id uuid.UUID) (*model.Vote, error) {
	v, err := s.ensureExpiredClosed(ctx, id)
	if err != nil {
		return nil, err
	}
	if !canCastVote(v.Status) {
		return nil, response.NewError(response.CodeVoteNotOpen, "投票未开始或已结束")
	}
	return v, nil
}

func (s *meetingService) voteDTO(ctx context.Context, v *model.Vote, viewer uuid.UUID) (*dto.VoteResponse, error) {
	v, err := s.ensureExpiredClosed(ctx, v.ID)
	if err != nil {
		return nil, err
	}
	opts, err := s.votes.ListOptions(ctx, v.ID)
	if err != nil {
		return nil, fmt.Errorf("list options: %w", err)
	}
	hasVoted := false
	if viewer != uuid.Nil {
		rec, err := s.votes.GetRecord(ctx, v.ID, viewer)
		if err != nil {
			return nil, err
		}
		hasVoted = rec != nil
	}
	return mapVote(v, opts, hasVoted), nil
}

func validateOptions(opts []dto.VoteOptionInput) error {
	seen := map[string]struct{}{}
	for _, o := range opts {
		if o.Key == "" || o.Label == "" {
			return response.NewError(response.CodeBadRequest, "选项 key/label 不能为空")
		}
		if _, ok := seen[o.Key]; ok {
			return response.NewError(response.CodeBadRequest, "选项 key 重复")
		}
		seen[o.Key] = struct{}{}
	}
	return nil
}
