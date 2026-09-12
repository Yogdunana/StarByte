package service

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/configstore/dto"
	"github.com/Yogdunana/StarByte/backend/internal/configstore/model"
	"github.com/Yogdunana/StarByte/backend/internal/configstore/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/config"
	"github.com/Yogdunana/StarByte/backend/pkg/configstore"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
)

// SMTPTester sends an admin-only probe using the live mail channel.
type SMTPTester interface {
	SendTest(ctx context.Context, to string) error
}

var keyPattern = regexp.MustCompile(`^[a-z][a-z0-9_.]{1,99}$`)

type ConfigService interface {
	List(ctx context.Context, q dto.ListQuery) ([]dto.ConfigResponse, error)
	GetByKey(ctx context.Context, key string) (*dto.ConfigResponse, error)
	Create(ctx context.Context, operator uuid.UUID, req *dto.CreateConfigRequest) (*dto.ConfigResponse, error)
	Update(ctx context.Context, operator uuid.UUID, id uuid.UUID, req *dto.UpdateConfigRequest) (*dto.ConfigResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetSMTP(ctx context.Context) (*dto.SMTPSettingsResponse, error)
	UpdateSMTP(ctx context.Context, operator uuid.UUID, req *dto.UpdateSMTPRequest) (*dto.SMTPSettingsResponse, error)
	TestSMTP(ctx context.Context, req *dto.TestSMTPRequest) (*dto.TestSMTPResponse, error)
}

type configService struct {
	rows     repo.ConfigRepo
	store    configstore.Store
	fallback config.EmailConfig
	tester   SMTPTester
}

func NewConfigService(rows repo.ConfigRepo, store configstore.Store) *configService {
	return &configService{rows: rows, store: store}
}

func (s *configService) WithSMTP(fallback config.EmailConfig, tester SMTPTester) *configService {
	s.fallback = fallback
	s.tester = tester
	return s
}

func (s *configService) List(ctx context.Context, q dto.ListQuery) ([]dto.ConfigResponse, error) {
	rows, err := s.rows.List(ctx, q.Category, q.Keyword)
	if err != nil {
		return nil, fmt.Errorf("list configs: %w", err)
	}
	out := make([]dto.ConfigResponse, 0, len(rows))
	for i := range rows {
		out = append(out, toResp(&rows[i]))
	}
	return out, nil
}

func (s *configService) GetByKey(ctx context.Context, key string) (*dto.ConfigResponse, error) {
	row, err := s.rows.GetByKey(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("get config: %w", err)
	}
	if row == nil {
		return nil, response.NewError(response.CodeConfigNotFound, "配置不存在")
	}
	return ptrResp(row), nil
}

func (s *configService) Create(ctx context.Context, operator uuid.UUID, req *dto.CreateConfigRequest) (*dto.ConfigResponse, error) {
	key := strings.TrimSpace(req.ConfigKey)
	if !keyPattern.MatchString(key) {
		return nil, response.NewError(response.CodeConfigInvalidKey, "配置键须为小写字母开头，仅含字母数字._")
	}
	if !model.ValidType(req.ConfigType) {
		return nil, response.NewError(response.CodeConfigInvalidType, "不支持的配置类型")
	}
	if !model.ValidCategory(req.Category) {
		return nil, response.NewError(response.CodeConfigInvalidType, "不支持的配置分组")
	}
	if err := validateValue(req.ConfigType, req.ConfigValue); err != nil {
		return nil, err
	}
	exist, err := s.rows.GetByKey(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("check key: %w", err)
	}
	if exist != nil {
		return nil, response.NewError(response.CodeConfigKeyExists, "配置键已存在")
	}
	now := time.Now()
	row := &model.Config{
		ID:          uuid.New(),
		ConfigKey:   key,
		ConfigValue: req.ConfigValue,
		ConfigType:  req.ConfigType,
		Description: req.Description,
		Category:    req.Category,
		IsPublic:    req.IsPublic,
		UpdatedBy:   &operator,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.rows.Create(ctx, row); err != nil {
		return nil, fmt.Errorf("create config: %w", err)
	}
	_ = s.store.Invalidate(ctx, key)
	return ptrResp(row), nil
}

func (s *configService) Update(ctx context.Context, operator uuid.UUID, id uuid.UUID, req *dto.UpdateConfigRequest) (*dto.ConfigResponse, error) {
	row, err := s.rows.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get config: %w", err)
	}
	if row == nil {
		return nil, response.NewError(response.CodeConfigNotFound, "配置不存在")
	}
	if req.ConfigType != nil {
		if !model.ValidType(*req.ConfigType) {
			return nil, response.NewError(response.CodeConfigInvalidType, "不支持的配置类型")
		}
		row.ConfigType = *req.ConfigType
	}
	if req.Category != nil {
		if !model.ValidCategory(*req.Category) {
			return nil, response.NewError(response.CodeConfigInvalidType, "不支持的配置分组")
		}
		row.Category = *req.Category
	}
	if req.ConfigValue != nil {
		row.ConfigValue = *req.ConfigValue
	}
	if req.Description != nil {
		row.Description = *req.Description
	}
	if req.IsPublic != nil {
		row.IsPublic = *req.IsPublic
	}
	if err := validateValue(row.ConfigType, row.ConfigValue); err != nil {
		return nil, err
	}
	row.UpdatedBy = &operator
	row.UpdatedAt = time.Now()
	if err := s.rows.Update(ctx, row); err != nil {
		return nil, fmt.Errorf("update config: %w", err)
	}
	_ = s.store.Invalidate(ctx, row.ConfigKey)
	return ptrResp(row), nil
}

func (s *configService) Delete(ctx context.Context, id uuid.UUID) error {
	row, err := s.rows.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get config: %w", err)
	}
	if row == nil {
		return response.NewError(response.CodeConfigNotFound, "配置不存在")
	}
	if _, ok := model.ProtectedKeys()[row.ConfigKey]; ok {
		return response.NewError(response.CodeConfigProtected, "该配置由业务模块占用，不能删除")
	}
	if err := s.rows.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete config: %w", err)
	}
	_ = s.store.Invalidate(ctx, row.ConfigKey)
	return nil
}

func validateValue(typ, raw string) error {
	switch typ {
	case model.TypeNumber:
		if _, err := strconv.ParseFloat(strings.TrimSpace(raw), 64); err != nil {
			return response.NewError(response.CodeConfigInvalidValue, "数值类型格式不正确")
		}
	case model.TypeBoolean:
		switch strings.ToLower(strings.TrimSpace(raw)) {
		case "true", "false", "1", "0":
		default:
			return response.NewError(response.CodeConfigInvalidValue, "布尔值须为 true/false")
		}
	case model.TypeJSON:
		if raw == "" {
			return response.NewError(response.CodeConfigInvalidValue, "JSON 不能为空")
		}
		if !json.Valid([]byte(raw)) {
			return response.NewError(response.CodeConfigInvalidValue, "JSON 格式不正确")
		}
	}
	return nil
}

func toResp(row *model.Config) dto.ConfigResponse {
	return dto.ConfigResponse{
		ID:          row.ID.String(),
		ConfigKey:   row.ConfigKey,
		ConfigValue: row.ConfigValue,
		ConfigType:  row.ConfigType,
		Description: row.Description,
		Category:    row.Category,
		IsPublic:    row.IsPublic,
		UpdatedAt:   row.UpdatedAt,
		CreatedAt:   row.CreatedAt,
	}
}

func ptrResp(row *model.Config) *dto.ConfigResponse {
	r := toResp(row)
	return &r
}
