package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/internship/dto"
	"github.com/Yogdunana/StarByte/backend/internal/internship/model"
	"github.com/google/uuid"
)

func DefaultConfig() model.InternshipConfig {
	return model.InternshipConfig{
		AllowStudentEdit:  true,
		AllowMinisterEdit: true,
		RankingVisible:    true,
	}
}

func parseConfig(raw string) model.InternshipConfig {
	cfg := DefaultConfig()
	if raw == "" {
		return cfg
	}
	var parsed model.InternshipConfig
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return cfg
	}
	return parsed
}

func (s *internshipService) loadConfig(ctx context.Context) (model.InternshipConfig, error) {
	row, err := s.rows.GetConfig(ctx, model.ConfigKey)
	if err != nil {
		return DefaultConfig(), fmt.Errorf("get internship config: %w", err)
	}
	if row == nil {
		return DefaultConfig(), nil
	}
	return parseConfig(row.ConfigValue), nil
}

func (s *internshipService) GetConfig(ctx context.Context) (*dto.InternshipConfigResponse, error) {
	cfg, err := s.loadConfig(ctx)
	if err != nil {
		return nil, err
	}
	return mapConfig(cfg), nil
}

func (s *internshipService) UpdateConfig(ctx context.Context, operator uuid.UUID, req *dto.InternshipConfigRequest) (*dto.InternshipConfigResponse, error) {
	cfg, err := s.loadConfig(ctx)
	if err != nil {
		return nil, err
	}
	if req.AllowStudentEdit != nil {
		cfg.AllowStudentEdit = *req.AllowStudentEdit
	}
	if req.AllowMinisterEdit != nil {
		cfg.AllowMinisterEdit = *req.AllowMinisterEdit
	}
	if req.RankingVisible != nil {
		cfg.RankingVisible = *req.RankingVisible
	}
	raw, err := json.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("marshal config: %w", err)
	}
	exist, err := s.rows.GetConfig(ctx, model.ConfigKey)
	if err != nil {
		return nil, fmt.Errorf("get config: %w", err)
	}
	now := time.Now()
	if exist == nil {
		exist = &model.SystemConfig{
			ID:          uuid.New(),
			ConfigKey:   model.ConfigKey,
			ConfigType:  "json",
			Description: "实习权限开关",
			Category:    "internship",
			CreatedAt:   now,
		}
	}
	exist.ConfigValue = string(raw)
	exist.UpdatedBy = &operator
	exist.UpdatedAt = now
	if err := s.rows.UpsertConfig(ctx, exist); err != nil {
		return nil, fmt.Errorf("save config: %w", err)
	}
	return mapConfig(cfg), nil
}

func mapConfig(cfg model.InternshipConfig) *dto.InternshipConfigResponse {
	return &dto.InternshipConfigResponse{
		AllowStudentEdit:  cfg.AllowStudentEdit,
		AllowMinisterEdit: cfg.AllowMinisterEdit,
		RankingVisible:    cfg.RankingVisible,
	}
}
