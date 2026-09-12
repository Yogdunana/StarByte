package service

import (
	"context"
	"fmt"

	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/internal/report/dto"
	"github.com/Yogdunana/StarByte/backend/internal/report/model"
	"github.com/Yogdunana/StarByte/backend/internal/report/repo"
	"github.com/google/uuid"
)

type Service interface {
	List(
		ctx context.Context,
		viewerID uuid.UUID,
		req *dto.ListReportRequest,
		scope *rbacModel.DataScopeCondition,
	) ([]*dto.ReportResponse, int64, int, int, error)
}

type reportService struct {
	reports repo.ReportRepo
}

func New(reports repo.ReportRepo) Service {
	return &reportService{reports: reports}
}

func (s *reportService) List(
	ctx context.Context,
	viewerID uuid.UUID,
	req *dto.ListReportRequest,
	scope *rbacModel.DataScopeCondition,
) ([]*dto.ReportResponse, int64, int, int, error) {
	query := dto.ListReportRequest{}
	if req != nil {
		query = *req
	}
	query.Page, query.PageSize = normalizePage(query.Page, query.PageSize)

	rows, total, err := s.reports.List(ctx, &query, viewerID, scope)
	if err != nil {
		return nil, 0, 0, 0, fmt.Errorf("list reports: %w", err)
	}

	return mapReports(rows), total, query.Page, query.PageSize, nil
}

func normalizePage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

func mapReports(rows []model.Report) []*dto.ReportResponse {
	result := make([]*dto.ReportResponse, 0, len(rows))
	for i := range rows {
		result = append(result, mapReport(&rows[i]))
	}
	return result
}

func mapReport(report *model.Report) *dto.ReportResponse {
	response := &dto.ReportResponse{
		ID:                  report.ID.String(),
		UserID:              report.UserID.String(),
		ReportType:          string(report.ReportType),
		PeriodStart:         report.PeriodStart,
		PeriodEnd:           report.PeriodEnd,
		WorkContent:         report.WorkContent,
		NextPlan:            report.NextPlan,
		ProblemsSuggestions: report.ProblemsSuggestions,
		ReviewStatus:        string(report.ReviewStatus),
		SubmittedAt:         report.SubmittedAt,
		CreatedAt:           report.CreatedAt,
		UpdatedAt:           report.UpdatedAt,
	}
	if report.DepartmentID != nil {
		departmentID := report.DepartmentID.String()
		response.DepartmentID = &departmentID
	}
	return response
}
