package service

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/Yogdunana/StarByte/backend/internal/audit/dto"
	"github.com/Yogdunana/StarByte/backend/internal/audit/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
)

const maxArchiveDecode = 32 * 1024 * 1024

func (s *auditService) ListArchives(ctx context.Context, req *dto.ArchiveListRequest) ([]dto.ArchiveListItem, int64, error) {
	if req == nil {
		req = &dto.ArchiveListRequest{Page: 1, PageSize: 20}
	}
	rows, total, err := s.auditRepo.ListArchives(ctx, req.Page, req.PageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("list archives: %w", err)
	}
	out := make([]dto.ArchiveListItem, 0, len(rows))
	for _, row := range rows {
		out = append(out, toArchiveItem(row))
	}
	return out, total, nil
}

func (s *auditService) PullArchive(ctx context.Context, id uuid.UUID, req *dto.ArchiveListRequest) (*dto.ArchivePullResponse, error) {
	if req == nil {
		req = &dto.ArchiveListRequest{Page: 1, PageSize: 20}
	}
	page, pageSize := req.Page, req.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = model.DefaultListPageSize
	}
	row, err := s.auditRepo.GetArchiveByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get archive: %w", err)
	}
	if row == nil {
		return nil, response.NewError(response.CodeAuditArchiveNotFound, "归档记录不存在")
	}
	if strings.TrimSpace(row.MinIOObject) == "" {
		return nil, response.NewError(response.CodeAuditArchiveFetch, "归档对象路径为空")
	}
	raw, err := s.downloadObject(ctx, row.MinIOObject)
	if err != nil {
		return nil, response.NewError(response.CodeAuditArchiveFetch, "拉取归档失败: "+err.Error())
	}
	logs, truncated, err := decodeArchiveLogs(raw)
	if err != nil {
		return nil, response.NewError(response.CodeAuditArchiveFetch, "解析归档失败: "+err.Error())
	}
	logs = filterArchiveLogs(logs, req.Keyword)
	total := int64(len(logs))
	start := (page - 1) * pageSize
	if start > len(logs) {
		start = len(logs)
	}
	end := start + pageSize
	if end > len(logs) {
		end = len(logs)
	}
	pageLogs := logs[start:end]
	list := make([]dto.AuditLogListResponse, 0, len(pageLogs))
	for _, log := range pageLogs {
		list = append(list, toListResponse(log))
	}
	return &dto.ArchivePullResponse{
		Archive:   toArchiveItem(*row),
		List:      list,
		Total:     total,
		Page:      page,
		PageSize:  pageSize,
		Truncated: truncated,
	}, nil
}

func (s *auditService) downloadObject(ctx context.Context, objectName string) ([]byte, error) {
	if s.store == nil {
		return nil, fmt.Errorf("MinIO 未配置")
	}
	rc, _, err := s.store.Download(ctx, objectName)
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return io.ReadAll(io.LimitReader(rc, maxArchiveDecode+1))
}

func decodeArchiveLogs(raw []byte) ([]model.AuditLog, bool, error) {
	if len(raw) > maxArchiveDecode {
		return nil, true, fmt.Errorf("归档对象过大")
	}
	payload := raw
	if len(raw) >= 2 && raw[0] == 0x1f && raw[1] == 0x8b {
		zr, err := gzip.NewReader(bytes.NewReader(raw))
		if err != nil {
			return nil, false, err
		}
		defer zr.Close()
		decoded, err := io.ReadAll(io.LimitReader(zr, maxArchiveDecode+1))
		if err != nil {
			return nil, false, err
		}
		if len(decoded) > maxArchiveDecode {
			return nil, true, fmt.Errorf("解压后超过上限")
		}
		payload = decoded
	}
	var logs []model.AuditLog
	if err := json.Unmarshal(payload, &logs); err != nil {
		return nil, false, err
	}
	return logs, false, nil
}

func filterArchiveLogs(logs []model.AuditLog, keyword string) []model.AuditLog {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return logs
	}
	like := strings.ToLower(keyword)
	out := make([]model.AuditLog, 0, len(logs))
	for _, log := range logs {
		hay := strings.ToLower(log.Path + " " + log.Username + " " + log.Module + " " + log.Action)
		if strings.Contains(hay, like) {
			out = append(out, log)
		}
	}
	return out
}
