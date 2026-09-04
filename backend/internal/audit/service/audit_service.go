package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/audit/dto"
	"github.com/Yogdunana/StarByte/backend/internal/audit/model"
	"github.com/Yogdunana/StarByte/backend/internal/audit/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/audit"
	"github.com/Yogdunana/StarByte/backend/pkg/config"
	"github.com/Yogdunana/StarByte/backend/pkg/events"
	"github.com/Yogdunana/StarByte/backend/pkg/logger"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// AuditService 审计日志服务（含 Issue 要求的 AuditLogger 能力）
type AuditService interface {
	Log(ctx context.Context, entry *model.AuditEntry) error
	LogAsync(ctx context.Context, entry *model.AuditEntry) error
	Query(ctx context.Context, req *dto.ListAuditLogRequest) ([]dto.AuditLogListResponse, int64, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.AuditLogResponse, error)
	Export(ctx context.Context, req *dto.ExportAuditLogRequest) ([]byte, string, error)
	Archive(ctx context.Context, beforeDays int) (*dto.ArchiveResponse, error)
}

type minioUploader func(cfg *config.MinIOConfig, objectName string, data []byte, contentType string) error

type auditService struct {
	auditRepo repo.AuditRepo
	minioCfg  *config.MinIOConfig
	uploadFn  minioUploader
}

func NewAuditService(auditRepo repo.AuditRepo, minioCfg *config.MinIOConfig) AuditService {
	return &auditService{
		auditRepo: auditRepo,
		minioCfg:  minioCfg,
		uploadFn:  uploadToMinIO,
	}
}

func (s *auditService) Log(ctx context.Context, entry *model.AuditEntry) error {
	if entry == nil {
		return nil
	}
	return s.auditRepo.Create(ctx, toAuditLog(entry))
}

func (s *auditService) LogAsync(ctx context.Context, entry *model.AuditEntry) error {
	go func() {
		if err := s.Log(context.Background(), entry); err != nil {
			logger.Error("audit log: async write failed", zap.Error(err))
		}
	}()
	return nil
}

func (s *auditService) Query(ctx context.Context, req *dto.ListAuditLogRequest) ([]dto.AuditLogListResponse, int64, error) {
	params := toListParams(req.UserID, req.Username, req.Action, req.Module, req.Keyword, req.IPAddress, req.Method, req.StartTime, req.EndTime)
	params.Page = req.Page
	params.PageSize = req.PageSize
	logs, total, err := s.auditRepo.List(ctx, params)
	if err != nil {
		return nil, 0, fmt.Errorf("query audit logs: %w", err)
	}
	result := make([]dto.AuditLogListResponse, 0, len(logs))
	for _, log := range logs {
		result = append(result, toListResponse(log))
	}
	return result, total, nil
}

func (s *auditService) GetByID(ctx context.Context, id uuid.UUID) (*dto.AuditLogResponse, error) {
	log, err := s.auditRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get audit log: %w", err)
	}
	if log == nil {
		return nil, response.NewError(response.CodeAuditNotFound, "审计日志不存在")
	}
	resp := toDetailResponse(*log)
	return &resp, nil
}

func toAuditLog(entry *model.AuditEntry) *model.AuditLog {
	ts := entry.Timestamp
	if ts.IsZero() {
		ts = time.Now()
	}
	log := &model.AuditLog{
		ID:             uuid.New(),
		Username:       entry.Username,
		RealName:       entry.RealName,
		Operation:      entry.Method + " " + entry.Path,
		Method:         entry.Method,
		Path:           entry.Path,
		Module:         entry.Module,
		Action:         entry.Action,
		IP:             entry.IPAddress,
		UserAgent:      entry.UserAgent,
		RequestParams:  string(entry.RequestBody),
		ResponseStatus: entry.ResponseCode,
		DurationMs:     int(entry.Duration),
		RequestID:      entry.RequestID,
		CreatedAt:      ts,
	}
	if entry.UserID != uuid.Nil {
		id := entry.UserID
		log.UserID = &id
	}
	return log
}

func toListParams(userID, username, action, module, keyword, ip, method string, start, end *time.Time) *repo.ListParams {
	params := &repo.ListParams{
		Username:  username,
		Action:    action,
		Module:    module,
		Keyword:   keyword,
		IP:        ip,
		Method:    method,
		StartTime: start,
		EndTime:   end,
	}
	if userID != "" {
		if parsed, err := uuid.Parse(userID); err == nil {
			params.UserID = &parsed
		}
	}
	return params
}

func toUser(log model.AuditLog) dto.AuditUser {
	u := dto.AuditUser{Username: log.Username, RealName: log.RealName}
	if log.UserID != nil {
		u.ID = log.UserID.String()
	}
	return u
}

func toListResponse(log model.AuditLog) dto.AuditLogListResponse {
	return dto.AuditLogListResponse{
		ID:           log.ID.String(),
		User:         toUser(log),
		Method:       log.Method,
		Path:         log.Path,
		Module:       log.Module,
		Action:       log.Action,
		RequestBody:  audit.DesensitizeJSON(log.RequestParams),
		ResponseCode: log.ResponseStatus,
		IPAddress:    log.IP,
		UserAgent:    log.UserAgent,
		DurationMs:   log.DurationMs,
		Timestamp:    log.CreatedAt.Format(time.RFC3339),
	}
}

func toDetailResponse(log model.AuditLog) dto.AuditLogResponse {
	return dto.AuditLogResponse{
		ID:           log.ID.String(),
		User:         toUser(log),
		Method:       log.Method,
		Path:         log.Path,
		Module:       log.Module,
		Action:       log.Action,
		RequestBody:  audit.DesensitizeJSON(log.RequestParams),
		ResponseCode: log.ResponseStatus,
		IPAddress:    log.IP,
		UserAgent:    log.UserAgent,
		DurationMs:   log.DurationMs,
		Timestamp:    log.CreatedAt.Format(time.RFC3339),
	}
}

// RegisterAuthEvents 订阅登录/登出事件并异步写入审计日志。
func RegisterAuthEvents(bus *events.EventBus, svc AuditService) {
	if bus == nil || svc == nil {
		return
	}
	bus.Subscribe(events.EventUserLogin, func(ctx context.Context, e events.Event) error {
		ev, ok := e.(events.UserLoginEvent)
		if !ok {
			return nil
		}
		return svc.LogAsync(ctx, &model.AuditEntry{
			UserID:       ev.UserID,
			Username:     ev.Username,
			RealName:     ev.RealName,
			Method:       "POST",
			Path:         "/api/v1/auth/login",
			Module:       "auth",
			Action:       model.ActionLogin,
			IPAddress:    ev.IP,
			UserAgent:    ev.UserAgent,
			ResponseCode: 200,
			Timestamp:    time.Now(),
		})
	})
	bus.Subscribe(events.EventUserLogout, func(ctx context.Context, e events.Event) error {
		ev, ok := e.(events.UserLogoutEvent)
		if !ok {
			return nil
		}
		return svc.LogAsync(ctx, &model.AuditEntry{
			UserID:       ev.UserID,
			Username:     ev.Username,
			RealName:     ev.RealName,
			Method:       "POST",
			Path:         "/api/v1/auth/logout",
			Module:       "auth",
			Action:       model.ActionLogout,
			IPAddress:    ev.IP,
			UserAgent:    ev.UserAgent,
			ResponseCode: 200,
			Timestamp:    time.Now(),
		})
	})
}
