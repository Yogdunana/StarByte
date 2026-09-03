package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
	auditRepo repo.AuditRepo
	minioCfg  *config.MinIOConfig
}

// NewAuditService 创建审计日志服务
func NewAuditService(auditRepo repo.AuditRepo, minioCfg *config.MinIOConfig) AuditService {
	return &auditService{
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
		ResponseBody:   Desensitize(log.ResponseBody),
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
		RequestID: req.RequestID,
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
		data, err := s.exportJSON(logs)
		if err != nil {
			return nil, "", fmt.Errorf("marshal json export: %w", err)
		}
		return data, "audit_logs.json", nil
	default:
		return nil, "", response.NewError(response.CodeAuditExportErr, "不支持的导出格式: "+req.Format)
	}
}

func (s *auditService) exportCSV(logs []model.AuditLog) []byte {
	var buf bytes.Buffer

	// BOM for Excel compatibility
	buf.Write([]byte{0xEF, 0xBB, 0xBF})

	// Header
	headers := []string{"ID", "用户ID", "用户名", "操作", "方法", "路径", "IP", "UserAgent", "状态码", "耗时(ms)", "请求ID", "已归档", "时间"}
	buf.WriteString(strings.Join(headers, ","))
	buf.WriteString("\n")

	for _, log := range logs {
		userID := ""
		if log.UserID != nil {
			userID = log.UserID.String()
		}
		isArchived := "否"
		if log.IsArchived {
			isArchived = "是"
		}
		row := []string{
			log.ID.String(),
			userID,
			log.Username,
			log.Operation,
			log.Method,
			log.Path,
			log.IP,
			log.UserAgent,
			fmt.Sprintf("%d", log.ResponseStatus),
			fmt.Sprintf("%d", log.DurationMs),
			log.RequestID,
			isArchived,
			log.CreatedAt.Format(time.RFC3339),
		}
		// Escape CSV fields
		for i, field := range row {
			row[i] = `"` + strings.ReplaceAll(field, `"`, `""`) + `"`
		}
		buf.WriteString(strings.Join(row, ","))
		buf.WriteString("\n")
	}

	return buf.Bytes()
}

func (s *auditService) exportJSON(logs []model.AuditLog) ([]byte, error) {
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
			ResponseBody:   Desensitize(log.ResponseBody),
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

	return json.MarshalIndent(result, "", "  ")
}

func (s *auditService) Archive(ctx context.Context, beforeDays int) (*dto.ArchiveResponse, error) {
	if beforeDays <= 0 {
		beforeDays = 90
	}

	before := time.Now().AddDate(0, 0, -beforeDays)
	archiveDate := before.Format("2006-01-02")

	// 1. 查询需要归档的日志（仅未归档的）
	notArchived := false
	params := &repo.ListParams{
		PageSize:   10000,
		EndTime:    &before,
		IsArchived: &notArchived,
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
			Message:     "无需归档的日志",
		}, nil
	}

	// 2. 导出为 JSON
	archiveData, err := json.Marshal(logs)
	if err != nil {
		return nil, fmt.Errorf("marshal archive data: %w", err)
	}

	// 3. 上传到 MinIO
	minioUploadOK := true
	minioObject := fmt.Sprintf("audit-logs/%s.json", archiveDate)
	if s.minioCfg != nil && s.minioCfg.Endpoint != "" {
		err = uploadToMinIO(s.minioCfg, minioObject, archiveData)
		if err != nil {
			minioUploadOK = false
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
	archiveStatus := 1 // 成功
	if !minioUploadOK {
		archiveStatus = 2 // MinIO 上传失败
	}
	archive := &model.AuditLogArchive{
		ID:          uuid.New(),
		ArchiveDate: archiveDate,
		RecordCount: affected,
		MinIOObject: minioObject,
		Status:      archiveStatus,
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
		Message:     fmt.Sprintf("成功归档 %d 条日志", affected),
	}, nil
}

// Desensitize 对请求参数中的敏感字段进行脱敏
// 该函数在查询详情和导出时调用，确保返回给前端的敏感数据已被遮蔽
func Desensitize(params string) string {
	if params == "" {
		return params
	}
	return desensitizeJSON(params)
}

// desensitizeJSON 对 JSON 字符串中的敏感字段进行脱敏
// 使用预编译的正则表达式，线程安全
func desensitizeJSON(body string) string {
	result := body
	quote := string('"')
	for _, field := range sensitiveFields {
		re := getPattern(field)
		if re == nil {
			continue
		}
		// Build replacement: "field":"[redacted]"
		replacement := quote + field + quote + `:` + quote + `[redacted]` + quote
		result = re.ReplaceAllString(result, replacement)
	}
	return result
}
