package repo

import (
	"context"
	"strings"

	"github.com/Yogdunana/StarByte/backend/internal/audit/model"
)

// CountRow 聚合计数行。
type CountRow struct {
	Key   string
	Count int64
}

var allowedGroupColumns = map[string]string{
	"action":   "action",
	"module":   "module",
	"username": "username",
}

func (r *auditRepo) GroupCount(ctx context.Context, req *ListParams, column string) ([]CountRow, error) {
	col, ok := allowedGroupColumns[column]
	if !ok {
		return nil, nil
	}
	var rows []CountRow
	err := r.buildQuery(ctx, req).
		Select(col + " AS key, COUNT(*) AS count").
		Group(col).
		Order("count DESC").
		Scan(&rows).Error
	return rows, err
}

func (r *auditRepo) GroupCompliance(ctx context.Context, req *ListParams) ([]CountRow, error) {
	var rows []CountRow
	err := r.buildQuery(ctx, req).
		Select("flag AS key, COUNT(*) AS count").
		Where("compliance_flags <> ''").
		Joins("CROSS JOIN LATERAL unnest(string_to_array(compliance_flags, ',')) AS flag").
		Group("flag").
		Order("count DESC").
		Scan(&rows).Error
	if err != nil {
		return r.groupComplianceFallback(ctx, req)
	}
	return rows, nil
}

func (r *auditRepo) groupComplianceFallback(ctx context.Context, req *ListParams) ([]CountRow, error) {
	counts := map[string]int64{}
	err := r.Iterate(ctx, req, model.DefaultIterateBatch, func(logs []model.AuditLog) error {
		for _, log := range logs {
			for _, flag := range strings.Split(log.ComplianceFlags, ",") {
				flag = strings.TrimSpace(flag)
				if flag == "" {
					continue
				}
				counts[flag]++
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	out := make([]CountRow, 0, len(counts))
	for k, v := range counts {
		out = append(out, CountRow{Key: k, Count: v})
	}
	return out, nil
}
