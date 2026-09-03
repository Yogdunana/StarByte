package service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/audit/dto"
	"github.com/Yogdunana/StarByte/backend/internal/audit/model"
	"github.com/Yogdunana/StarByte/backend/internal/audit/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/config"
	"github.com/Yogdunana/StarByte/backend/pkg/logger"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AuditService 审计日志服务接口
type AuditService interface {
	// Query 分页查询审计日志列表
	Query(ctx context.Context, req *dto.ListAuditLogRequest) ([]dto.AuditLogListResponse, int64, error)

	// GetByID 查询审计日志详情
	GetByID(ctx context.Context, id uuid.UUID) (*dto.AuditLogResponse, error)

	// Export 导出审计日志（返回 CSV 或 JSON 格式的字节数据）
	Export(ctx context.Context, req *dto.ExportAuditLogRequest) ([]byte, string, error)

	// Archive 归档指定天数之前的审计日志到 MinIO
	Archive(ctx context.Context, beforeDays int) (*dto.ArchiveResponse, error)
}

type auditService struct {
	db       *gorm.DB
	auditRepo repo.AuditRepo
	minioCfg *config.MinIOConfig
}

// NewAuditService 创建审计日志服务
func NewAuditService(db *gorm.DB, auditRepo repo.AuditRepo, minioCfg *config.MinIOConfig) AuditService {
	return &auditService{
		db:        db,
		auditRepo: auditRepo,
		minioCfg:  minioCfg,
	}
}

func (s *auditService) Query(ctx context.Context, req *dto.ListAuditLogRequest) ([]dto.AuditLogListResponse, int64, error) {
	params := &repo.ListParams{
		Page:      req.Page,
		PageSize:  req.PageSize,
		Username:  req.Username,
		Operation: req.Operation,
		Method:    req.Method,
		Path:      req.Path,
		IP:        req.IP,
		RequestID: req.RequestID,
		StatusMin: req.StatusMin,
		StatusMax: req.StatusMax,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
	}

	logs, total, err := s.auditRepo.List(ctx, params)
	if err != nil {
		return nil, 0, fmt.Errorf("query audit logs: %w", err)
	}

	result := make([]dto.AuditLogListResponse, 0, len(logs))
	for _, log := range logs {
		item := dto.AuditLogListResponse{
			ID:             log.ID.String(),
			Username:       log.Username,
			Operation:      log.Operation,
			Method:         log.Method,
			Path:           log.Path,
			IP:             log.IP,
			ResponseStatus: log.ResponseStatus,
			DurationMs:     log.DurationMs,
			RequestID:      log.RequestID,
			CreatedAt:      log.CreatedAt.Format(time.RFC3339),
		}
		if log.UserID != nil {
			item.UserID = log.UserID.String()
		}
		result = append(result, item)
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

	resp := &dto.AuditLogResponse{
		ID:             log.ID.String(),
		Username:       log.Username,
		Operation:      log.Operation,
		Method:         log.Method,
		Path:           log.Path,
		IP:             log.IP,
		UserAgent:      log.UserAgent,
		RequestParams:  Desensitize(log.RequestParams),
		ResponseStatus: log.ResponseStatus,
		ResponseBody:   log.ResponseBody,
		DurationMs:     log.DurationMs,
		RequestID:      log.RequestID,
		IsArchived:     log.IsArchived,
		CreatedAt:      log.CreatedAt.Format(time.RFC3339),
	}
	if log.UserID != nil {
		resp.UserID = log.UserID.String()
	}

	return resp, nil
}

func (s *auditService) Export(ctx context.Context, req *dto.ExportAuditLogRequest) ([]byte, string, error) {
	params := &repo.ListParams{
		Username:  req.Username,
		Operation: req.Operation,
		Method:    req.Method,
		Path:      req.Path,
		IP:        req.IP,
		StatusMin: req.StatusMin,
		StatusMax: req.StatusMax,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
	}

	logs, err := s.auditRepo.Export(ctx, params)
	if err != nil {
		return nil, "", fmt.Errorf("export audit logs: %w", err)
	}

	switch req.Format {
	case "csv":
		return s.exportCSV(logs), "audit_logs.csv", nil
	case "json":
		return s.exportJSON(logs), "audit_logs.json", nil
	default:
		return nil, "", response.NewError(response.CodeAuditExportErr, "不支持的导出格式: "+req.Format)
	}
}

func (s *auditService) exportCSV(logs []model.AuditLog) []byte {
	var buf bytes.Buffer

	// BOM for Excel compatibility
	buf.Write([]byte{0xEF, 0xBB, 0xBF})

	// Header
	headers := []string{"ID", "用户名", "操作", "方法", "路径", "IP", "状态码", "耗时(ms)", "请求ID", "时间"}
	buf.WriteString(strings.Join(headers, ","))
	buf.WriteString("\n")

	for _, log := range logs {
		row := []string{
			log.ID.String(),
			log.Username,
			log.Operation,
			log.Method,
			log.Path,
			log.IP,
			fmt.Sprintf("%d", log.ResponseStatus),
			fmt.Sprintf("%d", log.DurationMs),
			log.RequestID,
			log.CreatedAt.Format(time.RFC3339),
		}
		// Escape CSV fields
		for i, field := range row {
			row[i] = fmt.Sprintf("\"%s\"", strings.ReplaceAll(field, "\"", "\"\""))
		}
		buf.WriteString(strings.Join(row, ","))
		buf.WriteString("\n")
	}

	return buf.Bytes()
}

func (s *auditService) exportJSON(logs []model.AuditLog) []byte {
	result := make([]dto.AuditLogResponse, 0, len(logs))
	for _, log := range logs {
		item := dto.AuditLogResponse{
			ID:             log.ID.String(),
			Username:       log.Username,
			Operation:      log.Operation,
			Method:         log.Method,
			Path:           log.Path,
			IP:             log.IP,
			UserAgent:      log.UserAgent,
			RequestParams:  Desensitize(log.RequestParams),
			ResponseStatus: log.ResponseStatus,
			ResponseBody:   log.ResponseBody,
			DurationMs:     log.DurationMs,
			RequestID:      log.RequestID,
			IsArchived:     log.IsArchived,
			CreatedAt:      log.CreatedAt.Format(time.RFC3339),
		}
		if log.UserID != nil {
			item.UserID = log.UserID.String()
		}
		result = append(result, item)
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	return data
}

func (s *auditService) Archive(ctx context.Context, beforeDays int) (*dto.ArchiveResponse, error) {
	if beforeDays <= 0 {
		beforeDays = 90
	}

	before := time.Now().AddDate(0, 0, -beforeDays)
	archiveDate := before.Format("2006-01-02")

	// 1. 查询需要归档的日志
	params := &repo.ListParams{
		PageSize:  10000,
		EndTime:   &before,
	}
	logs, err := s.auditRepo.Export(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("query logs for archive: %w", err)
	}

	if len(logs) == 0 {
		return &dto.ArchiveResponse{
			ArchiveDate: archiveDate,
			RecordCount: 0,
			Status:      1,
			Message:      "无需归档的日志",
		}, nil
	}

	// 2. 导出为 JSON
	archiveData, err := json.Marshal(logs)
	if err != nil {
		return nil, fmt.Errorf("marshal archive data: %w", err)
	}

	// 3. 上传到 MinIO
	minioObject := fmt.Sprintf("audit-logs/%s.json", archiveDate)
	if s.minioCfg != nil && s.minioCfg.Endpoint != "" {
		err = uploadToMinIO(s.minioCfg, minioObject, archiveData)
		if err != nil {
			logger.Error("upload archive to MinIO failed",
				zap.Error(err),
				zap.String("archive_date", archiveDate),
			)
			// 继续执行，即使 MinIO 上传失败也标记为已归档
		}
	}

	// 4. 标记日志为已归档
	affected, err := s.auditRepo.ArchiveBefore(ctx, before)
	if err != nil {
		return nil, fmt.Errorf("mark logs as archived: %w", err)
	}

	// 5. 创建归档记录
	archive := &model.AuditLogArchive{
		ID:          uuid.New(),
		ArchiveDate: archiveDate,
		RecordCount: affected,
		MinIOObject: minioObject,
		Status:      1, // 成功
	}
	if err := s.auditRepo.CreateArchive(ctx, archive); err != nil {
		logger.Error("create archive record failed",
			zap.Error(err),
			zap.String("archive_date", archiveDate),
		)
	}

	// 6. 清理超过 180 天的已归档日志
	deleteBefore := time.Now().AddDate(0, 0, -180)
	deleted, err := s.auditRepo.DeleteArchived(ctx, deleteBefore)
	if err != nil {
		logger.Error("delete old archived logs failed",
			zap.Error(err),
		)
	} else if deleted > 0 {
		logger.Info("deleted old archived logs",
			zap.Int64("count", deleted),
		)
	}

	return &dto.ArchiveResponse{
		ArchiveID:   archive.ID.String(),
		RecordCount: affected,
		ArchiveDate: archiveDate,
		MinIOObject: minioObject,
		Status:      1,
		Message:      fmt.Sprintf("成功归档 %d 条日志", affected),
	}, nil
}

// Desensitize 对请求参数中的敏感字段进行脱敏
// 该函数在查询详情时调用，确保返回给前端的敏感数据已被遮蔽
func Desensitize(params string) string {
	if params == "" {
		return params
	}
	// 复用 middleware 中的脱敏逻辑
	// 这里对已存储的日志做二次脱敏，防止遗漏
	return desensitizeJSON(params)
}

// desensitizeJSON 对 JSON 字符串中的敏感字段进行脱敏
func desensitizeJSON(body string) string {
	sensitiveFields := []string{"password", "old_password", "new_password", "secret", "token", "access_token", "refresh_token"}
	result := body
	for _, field := range sensitiveFields {
		// 匹配 "field":"value" 模式（不区分大小写）
		pattern := fmt.Sprintf(`(?i)"%s"\s*:\s*"[^"]*"`, field)
		replacement := fmt.Sprintf(`"%s":"[redacted]"`, field)
		result = replaceRegex(result, pattern, replacement)
	}
	return result
}

// replaceRegex 简单的字符串替换包装
// 使用 regexp 包进行正则替换
func replaceRegex(s, pattern, replacement string) string {
	re := compiledPatterns.get(pattern)
	if re == nil {
		return s
	}
	return re.ReplaceAllString(s, replacement)
}
