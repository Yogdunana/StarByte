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

const defaultArchiveDecodeLimit = 32 * 1024 * 1024

var archiveDecodeLimit = defaultArchiveDecodeLimit

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
	defer func() { _ = rc.Close() }()
	return io.ReadAll(io.LimitReader(rc, int64(archiveDecodeLimit)+1))
}

func decodeArchiveLogs(raw []byte) ([]model.AuditLog, bool, error) {
	limit := archiveDecodeLimit
	truncated := false
	payload := raw
	if len(raw) >= 2 && raw[0] == 0x1f && raw[1] == 0x8b {
		zr, err := gzip.NewReader(bytes.NewReader(raw))
		if err != nil {
			return nil, false, err
		}
		defer func() { _ = zr.Close() }()
		decoded, err := io.ReadAll(io.LimitReader(zr, int64(limit)+1))
		if err != nil && len(decoded) == 0 {
			return nil, false, err
		}
		if err != nil || len(decoded) > limit {
			truncated = true
			if len(decoded) > limit {
				decoded = decoded[:limit]
			}
		}
		payload = decoded
	} else if len(raw) > limit {
		truncated = true
		payload = raw[:limit]
	}
	logs, err := unmarshalAuditLogs(payload)
	if err == nil {
		return logs, truncated, nil
	}
	if !truncated {
		return nil, false, err
	}
	return decodeCompleteAuditLogs(payload), true, nil
}

func unmarshalAuditLogs(payload []byte) ([]model.AuditLog, error) {
	var logs []model.AuditLog
	if err := json.Unmarshal(payload, &logs); err != nil {
		return nil, err
	}
	return logs, nil
}

func decodeCompleteAuditLogs(payload []byte) []model.AuditLog {
	dec := json.NewDecoder(bytes.NewReader(payload))
	tok, err := dec.Token()
	if err != nil {
		return nil
	}
	delim, ok := tok.(json.Delim)
	if !ok || delim != '[' {
		return nil
	}
	logs := make([]model.AuditLog, 0)
	for dec.More() {
		var log model.AuditLog
		if err := dec.Decode(&log); err != nil {
			break
		}
		logs = append(logs, log)
	}
	return logs
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
