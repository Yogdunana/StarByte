package service

import (
	"strings"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/audit/dto"
	"github.com/Yogdunana/StarByte/backend/internal/audit/model"
	"github.com/Yogdunana/StarByte/backend/internal/audit/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/audit"
	"github.com/google/uuid"
)

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

func splitFlags(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return []string{}
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func toDiff(log model.AuditLog) []dto.FieldChange {
	before := audit.DesensitizeJSON(log.BeforeJSON)
	after := audit.DesensitizeJSON(log.AfterJSON)
	raw := log.DiffJSON
	if strings.TrimSpace(raw) == "" {
		raw = audit.DiffJSON(before, after)
	}
	src := audit.ParseFieldChanges(raw)
	out := make([]dto.FieldChange, 0, len(src))
	for _, c := range src {
		out = append(out, dto.FieldChange{Path: c.Path, Before: c.Before, After: c.After})
	}
	return out
}

func toListResponse(log model.AuditLog) dto.AuditLogListResponse {
	return dto.AuditLogListResponse{
		ID:              log.ID.String(),
		User:            toUser(log),
		Method:          log.Method,
		Path:            log.Path,
		Module:          log.Module,
		Action:          log.Action,
		RequestBody:     audit.DesensitizeJSON(log.RequestParams),
		ResponseCode:    log.ResponseStatus,
		IPAddress:       log.IP,
		UserAgent:       log.UserAgent,
		DurationMs:      log.DurationMs,
		Timestamp:       log.CreatedAt.Format(time.RFC3339),
		EntityType:      log.EntityType,
		EntityID:        log.EntityID,
		ComplianceFlags: splitFlags(log.ComplianceFlags),
	}
}

func toDetailResponse(log model.AuditLog) dto.AuditLogResponse {
	return dto.AuditLogResponse{
		ID:              log.ID.String(),
		User:            toUser(log),
		Method:          log.Method,
		Path:            log.Path,
		Module:          log.Module,
		Action:          log.Action,
		RequestBody:     audit.DesensitizeJSON(log.RequestParams),
		ResponseCode:    log.ResponseStatus,
		IPAddress:       log.IP,
		UserAgent:       log.UserAgent,
		DurationMs:      log.DurationMs,
		Timestamp:       log.CreatedAt.Format(time.RFC3339),
		EntityType:      log.EntityType,
		EntityID:        log.EntityID,
		BeforeJSON:      audit.DesensitizeJSON(log.BeforeJSON),
		AfterJSON:       audit.DesensitizeJSON(log.AfterJSON),
		Diff:            toDiff(log),
		ComplianceFlags: splitFlags(log.ComplianceFlags),
	}
}

func toTraceItem(log model.AuditLog) dto.AuditTraceItem {
	return dto.AuditTraceItem{
		ID:              log.ID.String(),
		User:            toUser(log),
		Action:          log.Action,
		Method:          log.Method,
		Path:            log.Path,
		Module:          log.Module,
		EntityType:      log.EntityType,
		EntityID:        log.EntityID,
		BeforeJSON:      audit.DesensitizeJSON(log.BeforeJSON),
		AfterJSON:       audit.DesensitizeJSON(log.AfterJSON),
		Diff:            toDiff(log),
		ComplianceFlags: splitFlags(log.ComplianceFlags),
		Timestamp:       log.CreatedAt.Format(time.RFC3339),
	}
}

func toCountItems(rows []repo.CountRow) []dto.CountItem {
	out := make([]dto.CountItem, 0, len(rows))
	for _, row := range rows {
		if strings.TrimSpace(row.Key) == "" {
			continue
		}
		out = append(out, dto.CountItem{Key: row.Key, Count: row.Count})
	}
	return out
}

func toArchiveItem(row model.AuditLogArchive) dto.ArchiveListItem {
	return dto.ArchiveListItem{
		ID:          row.ID.String(),
		ArchiveDate: row.ArchiveDate,
		RecordCount: row.RecordCount,
		MinIOObject: row.MinIOObject,
		Status:      row.Status,
		CreatedAt:   row.CreatedAt.Format(time.RFC3339),
	}
}
