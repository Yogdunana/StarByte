package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/audit/dto"
	"github.com/Yogdunana/StarByte/backend/internal/audit/repo"
	exportdto "github.com/Yogdunana/StarByte/backend/internal/export/dto"
	exportsvc "github.com/Yogdunana/StarByte/backend/internal/export/service"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

const reportNote = "一期未包含频繁删除/批量导出告警，仅汇总合规标记与操作分布。"

func (s *auditService) Report(ctx context.Context, req *dto.ReportRequest) (*dto.ReportResponse, []byte, string, error) {
	if req == nil {
		req = &dto.ReportRequest{}
	}
	format := strings.ToLower(strings.TrimSpace(req.Format))
	if format == "" {
		format = "json"
	}
	switch format {
	case "json", "csv", "pdf", "excel":
	default:
		return nil, nil, "", response.NewError(response.CodeAuditExportErr, "不支持的报告格式: "+format)
	}

	start, end := req.StartTime, req.EndTime
	if start == nil && end == nil {
		now := time.Now()
		from := now.AddDate(0, 0, -30)
		start = &from
		end = &now
	}
	params := &repo.ListParams{Module: req.Module, StartTime: start, EndTime: end}

	total, err := s.auditRepo.Count(ctx, params)
	if err != nil {
		return nil, nil, "", fmt.Errorf("count report: %w", err)
	}
	byAction, err := s.auditRepo.GroupCount(ctx, params, "action")
	if err != nil {
		return nil, nil, "", fmt.Errorf("group action: %w", err)
	}
	byModule, err := s.auditRepo.GroupCount(ctx, params, "module")
	if err != nil {
		return nil, nil, "", fmt.Errorf("group module: %w", err)
	}
	byUser, err := s.auditRepo.GroupCount(ctx, params, "username")
	if err != nil {
		return nil, nil, "", fmt.Errorf("group user: %w", err)
	}
	byCompliance, err := s.auditRepo.GroupCompliance(ctx, params)
	if err != nil {
		return nil, nil, "", fmt.Errorf("group compliance: %w", err)
	}

	report := &dto.ReportResponse{
		Total:        total,
		ByAction:     toCountItems(byAction),
		ByModule:     toCountItems(byModule),
		ByCompliance: toCountItems(byCompliance),
		TopOperators: limitCounts(toCountItems(byUser), 10),
		Note:         reportNote,
	}
	if start != nil {
		report.StartTime = start.Format(time.RFC3339)
	}
	if end != nil {
		report.EndTime = end.Format(time.RFC3339)
	}
	if format == "json" {
		return report, nil, "", nil
	}

	table := reportTable(report)
	data, filename, err := exportsvc.RenderTable(format, table)
	if err != nil {
		return nil, nil, "", response.NewError(response.CodeAuditReportErr, "生成报告失败: "+err.Error())
	}
	return report, data, filename, nil
}

func limitCounts(items []dto.CountItem, n int) []dto.CountItem {
	if len(items) <= n {
		return items
	}
	return items[:n]
}

func reportTable(report *dto.ReportResponse) *exportdto.TableExportRequest {
	rows := [][]string{
		{"汇总", "total", fmt.Sprintf("%d", report.Total)},
	}
	appendCountRows := func(category string, items []dto.CountItem) {
		for _, item := range items {
			rows = append(rows, []string{category, item.Key, fmt.Sprintf("%d", item.Count)})
		}
	}
	appendCountRows("动作", report.ByAction)
	appendCountRows("模块", report.ByModule)
	appendCountRows("合规标记", report.ByCompliance)
	appendCountRows("操作人", report.TopOperators)
	return &exportdto.TableExportRequest{
		Filename: "audit_compliance_report",
		Title:    "审计合规报告",
		Columns:  []string{"类别", "键", "数量"},
		Rows:     rows,
		Sheet:    "报告",
	}
}
