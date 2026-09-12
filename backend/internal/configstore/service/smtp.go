package service

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/configstore/dto"
	"github.com/Yogdunana/StarByte/backend/internal/configstore/model"
	"github.com/Yogdunana/StarByte/backend/pkg/config"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
)

func (s *configService) GetSMTP(ctx context.Context) (*dto.SMTPSettingsResponse, error) {
	cfg, err := s.resolveSMTP(ctx)
	if err != nil {
		return nil, err
	}
	return smtpResponse(cfg), nil
}

func (s *configService) UpdateSMTP(ctx context.Context, operator uuid.UUID, req *dto.UpdateSMTPRequest) (*dto.SMTPSettingsResponse, error) {
	if req == nil {
		return nil, response.NewError(response.CodeConfigInvalidValue, "SMTP 参数无效")
	}
	mode := config.NormalizeSSLMode(req.SSLMode)
	if !config.ValidSSLMode(mode) {
		return nil, response.NewError(response.CodeConfigInvalidValue, "SSL/TLS 模式须为 implicit、starttls 或 none")
	}
	runtime := config.SMTPRuntime{
		Host:     strings.TrimSpace(req.Host),
		Port:     req.Port,
		SSLMode:  mode,
		From:     strings.TrimSpace(req.From),
		FromName: strings.TrimSpace(req.FromName),
		Username: strings.TrimSpace(req.Username),
	}
	if runtime.Username == "" {
		runtime.Username = runtime.From
	}
	row, err := s.rows.GetByKey(ctx, config.SMTPSettingsKey)
	if err != nil {
		return nil, err
	}
	if row != nil {
		previous, parseErr := config.ParseSMTPRuntime(row.ConfigValue)
		if parseErr != nil {
			return nil, response.NewError(response.CodeConfigInvalidValue, "SMTP 配置 JSON 无效")
		}
		runtime.PasswordCiphertext = previous.PasswordCiphertext
	}
	if req.Password != nil && *req.Password != "" {
		if len(*req.Password) > 4096 {
			return nil, response.NewError(response.CodeConfigInvalidValue, "SMTP 密码过长")
		}
		runtime.PasswordCiphertext, err = config.EncryptSMTPPassword(*req.Password)
		if err != nil {
			return nil, response.NewError(response.CodeConfigInvalidValue, err.Error())
		}
	}
	if _, err := runtime.Resolve(s.fallback); err != nil {
		return nil, response.NewError(response.CodeConfigInvalidValue, err.Error())
	}
	raw, err := json.Marshal(runtime)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	if row == nil {
		row = &model.Config{
			ID:          uuid.New(),
			ConfigKey:   config.SMTPSettingsKey,
			ConfigType:  model.TypeJSON,
			Category:    "notification",
			Description: "SMTP 发信设置（密码加密保存）",
			CreatedAt:   now,
		}
		row.ConfigValue = string(raw)
		row.UpdatedBy = &operator
		row.UpdatedAt = now
		if err := s.rows.Create(ctx, row); err != nil {
			return nil, err
		}
	} else {
		row.ConfigValue = string(raw)
		row.IsPublic = false
		row.ConfigType = model.TypeJSON
		row.Category = "notification"
		row.UpdatedBy = &operator
		row.UpdatedAt = now
		if err := s.rows.Update(ctx, row); err != nil {
			return nil, err
		}
	}
	_ = s.store.Invalidate(ctx, config.SMTPSettingsKey)
	cfg, err := s.resolveSMTP(ctx)
	if err != nil {
		return nil, err
	}
	return smtpResponse(cfg), nil
}

func (s *configService) TestSMTP(ctx context.Context, req *dto.TestSMTPRequest) (*dto.TestSMTPResponse, error) {
	if req == nil || strings.TrimSpace(req.To) == "" {
		return nil, response.NewError(response.CodeConfigInvalidValue, "请填写测试收件人")
	}
	cfg, err := s.resolveSMTP(ctx)
	if err != nil {
		return nil, err
	}
	if cfg.Password == "" || cfg.PasswordSource == "" {
		return nil, response.NewError(response.CodeNotificationEmailFail, "未配置 SMTP 密码")
	}
	if s.tester == nil {
		return nil, response.NewError(response.CodeNotificationEmailFail, "邮件发送器不可用")
	}
	if err := s.tester.SendTest(ctx, strings.TrimSpace(req.To)); err != nil {
		return nil, response.NewError(response.CodeNotificationEmailFail, err.Error())
	}
	return &dto.TestSMTPResponse{Sent: true, To: strings.TrimSpace(req.To)}, nil
}

func (s *configService) resolveSMTP(ctx context.Context) (config.EmailConfig, error) {
	cfg := s.fallback
	if cfg.SMTPHost == "" && cfg.From == "" {
		cfg = config.DefaultSMTPRuntime().Overlay(cfg)
	}
	row, err := s.rows.GetByKey(ctx, config.SMTPSettingsKey)
	if err != nil {
		return cfg, err
	}
	if row != nil {
		runtime, perr := config.ParseSMTPRuntime(row.ConfigValue)
		if perr != nil {
			return cfg, response.NewError(response.CodeConfigInvalidValue, "SMTP 配置 JSON 无效")
		}
		return runtime.Resolve(cfg)
	}
	return cfg.ApplyEnvPassword(), nil
}

func smtpResponse(cfg config.EmailConfig) *dto.SMTPSettingsResponse {
	return &dto.SMTPSettingsResponse{
		SMTPSettings: dto.SMTPSettings{
			Host:     cfg.SMTPHost,
			Port:     cfg.SMTPPort,
			SSLMode:  cfg.EffectiveSSLMode(),
			From:     cfg.From,
			FromName: cfg.FromName,
			Username: cfg.EffectiveUsername(),
		},
		PasswordConfigured: cfg.Password != "" && cfg.PasswordSource != "",
		PasswordSource:     cfg.PasswordSource,
	}
}
