package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/Yogdunana/StarByte/backend/internal/meeting/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

func (s *meetingService) freezeElectorate(ctx context.Context, v *model.Vote) ([]model.Elector, error) {
	if s.electorate == nil {
		return nil, nil
	} // In-memory legacy fixtures only.
	cfg := model.VoteWeightConfig{DefaultWeight: 1, Weights: map[string]float64{}}
	if v.VoteType == model.VoteWeighted {
		var err error
		cfg, err = s.loadWeight(ctx)
		if err != nil {
			return nil, err
		}
	}
	candidates, err := s.electorate.Candidates(ctx, v.MeetingID)
	if err != nil {
		return nil, fmt.Errorf("load voting electorate: %w", err)
	}
	if v.VoteType == model.VoteWeighted {
		for _, candidate := range candidates {
			if candidate.PositionCode != "" {
				if _, ok := configuredWeight(cfg, candidate.PositionCode); !ok {
					return nil, response.NewError(response.CodeBadRequest, "参会人绑定的职务尚未配置权重，请先完善计票设置")
				}
			}
		}
	}
	rows := buildElectorate(v.ID, candidates, cfg, v.VoteType)
	if len(rows) == 0 {
		return nil, response.NewError(response.CodeBadRequest, "没有可参与投票的有效参会人")
	}
	raw, err := json.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("snapshot voting weights: %w", err)
	}
	v.ElectorateFrozen = true
	v.EligibleCount = len(rows)
	v.WeightSnapshot = string(raw)
	return rows, nil
}
func buildElectorate(vote uuid.UUID, candidates []model.ElectorCandidate, cfg model.VoteWeightConfig, kind int16) []model.Elector {
	weights := map[uuid.UUID]float64{}
	for _, candidate := range candidates {
		weight := 0.0
		if kind == model.VoteEqual {
			weight = 1
		} else {
			known := false
			for _, code := range []string{candidate.PositionCode, candidate.RoleCode} {
				if code == "user" || code == "member" || code == "probationary" || code == "admin" || code == "super_admin" {
					known = true
					continue
				}
				if value, ok := configuredWeight(cfg, code); ok {
					known = true
					if value > weight {
						weight = value
					}
				}
			}
			if !known {
				weight = cfg.DefaultWeight
			}
		}
		prior, exists := weights[candidate.UserID]
		if !exists || weight > prior {
			weights[candidate.UserID] = weight
		}
	}
	rows := make([]model.Elector, 0, len(weights))
	for user, weight := range weights {
		if weight <= 0 {
			continue
		}
		rows = append(rows, model.Elector{VoteID: vote, UserID: user, Weight: weight})
	}
	return rows
}
func (s *meetingService) ballotWeight(ctx context.Context, v *model.Vote, user uuid.UUID) (float64, error) {
	if v.ElectorateFrozen {
		if s.electorate == nil {
			return 0, fmt.Errorf("voting electorate repository unavailable")
		}
		elector, err := s.electorate.Get(ctx, v.ID, user)
		if err != nil {
			return 0, fmt.Errorf("load frozen voting weight: %w", err)
		}
		if elector == nil {
			return 0, response.NewError(response.CodeVoteNoAccess, "你不在本轮投票开始时的参会名单中")
		}
		return elector.Weight, nil
	}
	// Preserve legacy behavior without inventing historical role snapshots.
	if v.VoteType == model.VoteEqual {
		return 1, nil
	}
	cfg, err := s.loadWeight(ctx)
	if err != nil {
		return 0, err
	}
	person, err := s.meetings.GetUser(ctx, user)
	if err != nil {
		return 0, fmt.Errorf("load voting position: %w", err)
	}
	code := ""
	if person != nil {
		code = person.PositionCode
	}
	return ResolveWeight(cfg, code, v.VoteType), nil
}
