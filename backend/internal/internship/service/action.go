package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/internship/dto"
	"github.com/Yogdunana/StarByte/backend/internal/internship/model"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
)

func (s *internshipService) Complete(ctx context.Context, operator, id uuid.UUID, req *dto.CompleteRequest, scope *rbacModel.DataScopeCondition) (*dto.InternshipResponse, error) {
	row, err := s.mustRow(ctx, id)
	if err != nil {
		return nil, err
	}
	if !canAccessRecord(scope, row.UserID, row.DepartmentID, operator) || !canCompleteOrDelete(row, operator, scope) {
		return nil, response.NewError(response.CodeInternshipNoAccess, "无权完成该实习记录")
	}
	if row.Status == model.StatusDone {
		return nil, response.NewError(response.CodeInternshipDupComplete, "实习已完成")
	}
	if row.Status != model.StatusOngoing {
		return nil, response.NewError(response.CodeInternshipInvalidState, "实习状态不允许该操作")
	}
	now := time.Now()
	if row.EndDate == nil {
		today := dateOnly(now)
		row.EndDate = &today
	}
	row.Status = model.StatusDone
	if req != nil {
		if req.Report != "" {
			row.Report = req.Report
		}
		if req.Achievements != "" {
			row.Achievements = req.Achievements
		}
	}
	row.UpdatedBy = &operator
	row.UpdatedAt = now
	if err := s.rows.Update(ctx, row); err != nil {
		return nil, fmt.Errorf("complete internship: %w", err)
	}
	return s.Get(ctx, operator, id, nil)
}

func (s *internshipService) SubmitReport(ctx context.Context, operator, id uuid.UUID, report string, scope *rbacModel.DataScopeCondition) (*dto.InternshipResponse, error) {
	row, err := s.mustRow(ctx, id)
	if err != nil {
		return nil, err
	}
	if !canAccessRecord(scope, row.UserID, row.DepartmentID, operator) {
		return nil, response.NewError(response.CodeInternshipNoAccess, "无权提交该实习报告")
	}
	if row.Status != model.StatusOngoing && row.UserID != operator && !isAllScope(scope) {
		return nil, response.NewError(response.CodeInternshipInvalidState, "实习状态不允许该操作")
	}
	if row.UserID != operator && !isAllScope(scope) {
		return nil, response.NewError(response.CodeInternshipNoAccess, "仅本人或社长可提交报告")
	}
	row.Report = report
	row.UpdatedBy = &operator
	row.UpdatedAt = time.Now()
	if err := s.rows.Update(ctx, row); err != nil {
		return nil, fmt.Errorf("submit report: %w", err)
	}
	return s.Get(ctx, operator, id, nil)
}

func (s *internshipService) MentorComment(ctx context.Context, operator, id uuid.UUID, comment string, scope *rbacModel.DataScopeCondition) (*dto.InternshipResponse, error) {
	row, err := s.mustRow(ctx, id)
	if err != nil {
		return nil, err
	}
	if !canAccessRecord(scope, row.UserID, row.DepartmentID, operator) {
		return nil, response.NewError(response.CodeInternshipNoAccess, "无权评价该实习记录")
	}
	row.MentorComment = comment
	row.UpdatedBy = &operator
	row.UpdatedAt = time.Now()
	if err := s.rows.Update(ctx, row); err != nil {
		return nil, fmt.Errorf("mentor comment: %w", err)
	}
	return s.Get(ctx, operator, id, nil)
}
