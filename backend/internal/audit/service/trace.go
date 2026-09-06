package service

import (
	"context"
	"regexp"
	"strings"

	"github.com/Yogdunana/StarByte/backend/internal/audit/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

var allowedEntityTypes = map[string]struct{}{
	"user":       {},
	"role":       {},
	"department": {},
	"permission": {},
	"position":   {},
	"member":     {},
	"interview":  {},
	"meeting":    {},
	"task":       {},
	"internship": {},
	"audit_log":  {},
}

var entityIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)

func validateEntity(entityType, entityID string) (string, string, error) {
	typ := strings.ToLower(strings.TrimSpace(entityType))
	id := strings.TrimSpace(entityID)
	if _, ok := allowedEntityTypes[typ]; !ok {
		return "", "", response.NewError(response.CodeAuditEntityInvalid, "无效的实体类型")
	}
	if !entityIDPattern.MatchString(id) {
		return "", "", response.NewError(response.CodeAuditEntityInvalid, "无效的实体ID")
	}
	return typ, id, nil
}

func (s *auditService) Trace(ctx context.Context, entityType, entityID string, req *dto.TraceQueryRequest) ([]dto.AuditTraceItem, int64, error) {
	typ, id, err := validateEntity(entityType, entityID)
	if err != nil {
		return nil, 0, err
	}
	if req == nil {
		req = &dto.TraceQueryRequest{Page: 1, PageSize: 20}
	}
	logs, total, err := s.auditRepo.ListByEntity(ctx, typ, id, req.Page, req.PageSize)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.AuditTraceItem, 0, len(logs))
	for _, log := range logs {
		out = append(out, toTraceItem(log))
	}
	return out, total, nil
}
