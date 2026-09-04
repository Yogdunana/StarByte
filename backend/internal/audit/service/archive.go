package service

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/audit/dto"
	"github.com/Yogdunana/StarByte/backend/internal/audit/model"
	"github.com/Yogdunana/StarByte/backend/internal/audit/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/logger"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (s *auditService) Archive(ctx context.Context, beforeDays int) (*dto.ArchiveResponse, error) {
	if beforeDays <= 0 {
		beforeDays = model.DefaultArchiveDays
	}
	before := time.Now().AddDate(0, 0, -beforeDays)
	runAt := time.Now()
	objectName := archiveObjectPath(runAt)

	params := &repo.ListParams{EndTime: &before}
	count, err := s.auditRepo.Count(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("count logs for archive: %w", err)
	}
	if count == 0 {
		return &dto.ArchiveResponse{
			ArchiveDate: runAt.Format("2006-01-02"),
			RecordCount: 0,
			Status:      1,
			Message:     "无需归档的日志",
		}, nil
	}

	payload, err := s.collectArchiveJSON(ctx, params)
	if err != nil {
		return nil, err
	}
	gzData, err := gzipBytes(payload)
	if err != nil {
		return nil, fmt.Errorf("gzip archive: %w", err)
	}

	if s.minioCfg == nil || s.minioCfg.Endpoint == "" || s.uploadFn == nil {
		return nil, response.NewError(response.CodeAuditArchiveErr, "归档失败: MinIO 未配置")
	}
	if err := s.uploadFn(s.minioCfg, objectName, gzData, "application/gzip"); err != nil {
		logger.Error("upload archive to MinIO failed", zap.Error(err), zap.String("object", objectName))
		return nil, response.NewError(response.CodeAuditArchiveErr, "归档失败: "+err.Error())
	}

	deleted, err := s.auditRepo.DeleteBefore(ctx, before)
	if err != nil {
		logger.Error("delete archived logs failed after MinIO upload",
			zap.Error(err),
			zap.String("object", objectName),
		)
	}

	archive := &model.AuditLogArchive{
		ID:          uuid.New(),
		ArchiveDate: runAt.Format("2006-01-02"),
		RecordCount: deleted,
		MinIOObject: objectName,
		Status:      1,
	}
	if deleted == 0 {
		archive.RecordCount = count
	}
	if err := s.auditRepo.CreateArchive(ctx, archive); err != nil {
		logger.Error("create archive record failed", zap.Error(err))
	}

	return &dto.ArchiveResponse{
		ArchiveID:   archive.ID.String(),
		RecordCount: archive.RecordCount,
		ArchiveDate: archive.ArchiveDate,
		MinIOObject: objectName,
		Status:      1,
		Message:     fmt.Sprintf("成功归档 %d 条日志", archive.RecordCount),
	}, nil
}

func (s *auditService) collectArchiveJSON(ctx context.Context, params *repo.ListParams) ([]byte, error) {
	var logs []model.AuditLog
	err := s.auditRepo.Iterate(ctx, params, model.DefaultIterateBatch, func(batch []model.AuditLog) error {
		logs = append(logs, batch...)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("query logs for archive: %w", err)
	}
	data, err := json.Marshal(logs)
	if err != nil {
		return nil, fmt.Errorf("marshal archive data: %w", err)
	}
	return data, nil
}

func gzipBytes(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	if _, err := zw.Write(data); err != nil {
		_ = zw.Close()
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func archiveObjectPath(t time.Time) string {
	return fmt.Sprintf("audit-logs/%s/%s/audit_logs_%s.json.gz",
		t.Format("2006"), t.Format("01"), t.Format("20060102"))
}
