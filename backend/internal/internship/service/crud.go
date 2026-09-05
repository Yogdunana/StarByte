package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/internship/dto"
	"github.com/Yogdunana/StarByte/backend/internal/internship/model"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
)

func (s *internshipService) Create(ctx context.Context, operator uuid.UUID, req *dto.CreateInternshipRequest, scope *rbacModel.DataScopeCondition) (*dto.InternshipResponse, error) {
	if err := validateRange(req.StartDate, req.EndDate); err != nil {
		return nil, err
	}
	ownerID := operator
	if id := parseUUIDPtr(req.UserID); id != nil {
		if *id != operator && !isAllScope(scope) && !canAccessRecord(scope, *id, nil, operator) {
			owner, err := s.rows.GetUser(ctx, *id)
			if err != nil {
				return nil, fmt.Errorf("lookup user: %w", err)
			}
			if owner == nil {
				return nil, response.NewError(response.CodeBadRequest, "实习对象不存在")
			}
			if !canAccessRecord(scope, *id, owner.DepartmentID, operator) {
				return nil, response.NewError(response.CodeInternshipNoAccess, "无权为该成员创建实习")
			}
		}
		ownerID = *id
	}
	owner, err := s.rows.GetUser(ctx, ownerID)
	if err != nil {
		return nil, fmt.Errorf("lookup owner: %w", err)
	}
	if owner == nil {
		return nil, response.NewError(response.CodeBadRequest, "实习对象不存在")
	}
	mentorID, err := s.lookupOptionalUser(ctx, req.MentorID)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	row := &model.Internship{
		ID:           uuid.New(),
		UserID:       ownerID,
		DepartmentID: owner.DepartmentID,
		Title:        strings.TrimSpace(req.Title),
		Organization: strings.TrimSpace(req.Organization),
		Description:  req.Description,
		StartDate:    req.StartDate,
		EndDate:      req.EndDate,
		Status:       model.StatusOngoing,
		Type:         req.Type,
		MentorID:     mentorID,
		Skills:       encodeSkills(req.Skills),
		Achievements: req.Achievements,
		CreatedBy:    operator,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if req.EndDate != nil && !req.EndDate.After(now) {
		row.Status = model.StatusDone
	}
	if err := s.rows.Create(ctx, row); err != nil {
		return nil, fmt.Errorf("create internship: %w", err)
	}
	return s.Get(ctx, operator, row.ID, nil)
}

func (s *internshipService) List(ctx context.Context, viewer uuid.UUID, req *dto.ListInternshipRequest, scope *rbacModel.DataScopeCondition) ([]*dto.InternshipResponse, int64, error) {
	rows, total, err := s.rows.List(ctx, req, rewriteScope(scope, viewer))
	if err != nil {
		return nil, 0, fmt.Errorf("list internships: %w", err)
	}
	out := make([]*dto.InternshipResponse, 0, len(rows))
	for i := range rows {
		out = append(out, mapInternship(&rows[i]))
	}
	return out, total, nil
}

func (s *internshipService) Get(ctx context.Context, viewer, id uuid.UUID, scope *rbacModel.DataScopeCondition) (*dto.InternshipResponse, error) {
	row, err := s.loadVisible(ctx, viewer, id, scope)
	if err != nil {
		return nil, err
	}
	return mapInternship(row), nil
}

func (s *internshipService) Update(ctx context.Context, operator, id uuid.UUID, req *dto.UpdateInternshipRequest, scope *rbacModel.DataScopeCondition) (*dto.InternshipResponse, error) {
	row, err := s.mustRow(ctx, id)
	if err != nil {
		return nil, err
	}
	if !canAccessRecord(scope, row.UserID, row.DepartmentID, operator) {
		return nil, response.NewError(response.CodeInternshipNoAccess, "无权操作该实习记录")
	}
	if model.IsClosed(row.Status) {
		return nil, response.NewError(response.CodeInternshipClosed, "实习已结束，无法修改")
	}
	cfg, err := s.loadConfig(ctx)
	if err != nil {
		return nil, err
	}
	if !canEdit(row, operator, scope, cfg) {
		return nil, response.NewError(response.CodeInternshipNoAccess, "无权修改该实习记录")
	}
	if err := applyUpdate(row, req); err != nil {
		return nil, err
	}
	if err := validateRange(row.StartDate, row.EndDate); err != nil {
		return nil, err
	}
	row.UpdatedBy = &operator
	row.UpdatedAt = time.Now()
	if err := s.rows.Update(ctx, row); err != nil {
		return nil, fmt.Errorf("update internship: %w", err)
	}
	return s.Get(ctx, operator, id, nil)
}

func (s *internshipService) Delete(ctx context.Context, operator, id uuid.UUID, scope *rbacModel.DataScopeCondition) error {
	row, err := s.mustRow(ctx, id)
	if err != nil {
		return err
	}
	if !canAccessRecord(scope, row.UserID, row.DepartmentID, operator) || !canCompleteOrDelete(row, operator, scope) {
		return response.NewError(response.CodeInternshipNoAccess, "无权删除该实习记录")
	}
	if err := s.rows.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete internship: %w", err)
	}
	return nil
}

func (s *internshipService) ListMine(ctx context.Context, userID uuid.UUID, status *int16) ([]*dto.InternshipResponse, error) {
	rows, err := s.rows.ListByUser(ctx, userID, status)
	if err != nil {
		return nil, fmt.Errorf("list my internships: %w", err)
	}
	out := make([]*dto.InternshipResponse, 0, len(rows))
	for i := range rows {
		out = append(out, mapInternship(&rows[i]))
	}
	return out, nil
}

func (s *internshipService) mustRow(ctx context.Context, id uuid.UUID) (*model.Internship, error) {
	row, err := s.rows.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get internship: %w", err)
	}
	if row == nil {
		return nil, response.NewError(response.CodeInternshipNotFound, "实习记录不存在")
	}
	return row, nil
}

func (s *internshipService) loadVisible(ctx context.Context, viewer, id uuid.UUID, scope *rbacModel.DataScopeCondition) (*model.InternshipWithNames, error) {
	row, err := s.rows.GetByIDWithNames(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get internship: %w", err)
	}
	if row == nil {
		return nil, response.NewError(response.CodeInternshipNotFound, "实习记录不存在")
	}
	if !canAccessRecord(scope, row.UserID, row.DepartmentID, viewer) {
		return nil, response.NewError(response.CodeInternshipNoAccess, "无权查看该实习记录")
	}
	return row, nil
}

func (s *internshipService) lookupOptionalUser(ctx context.Context, raw string) (*uuid.UUID, error) {
	id := parseUUIDPtr(raw)
	if id == nil {
		return nil, nil
	}
	u, err := s.rows.GetUser(ctx, *id)
	if err != nil {
		return nil, fmt.Errorf("lookup mentor: %w", err)
	}
	if u == nil {
		return nil, response.NewError(response.CodeBadRequest, "指导老师不存在")
	}
	return id, nil
}

func validateRange(start time.Time, end *time.Time) error {
	if start.IsZero() {
		return response.NewError(response.CodeBadRequest, "开始日期不能为空")
	}
	if end != nil && end.Before(start) {
		return response.NewError(response.CodeBadRequest, "结束日期不能早于开始日期")
	}
	return nil
}

func applyUpdate(row *model.Internship, req *dto.UpdateInternshipRequest) error {
	if req.Title != nil {
		row.Title = strings.TrimSpace(*req.Title)
	}
	if req.Organization != nil {
		row.Organization = strings.TrimSpace(*req.Organization)
	}
	if req.Description != nil {
		row.Description = *req.Description
	}
	if req.EndDate != nil {
		row.EndDate = req.EndDate
	}
	if req.Skills != nil {
		row.Skills = encodeSkills(req.Skills)
	}
	if req.Achievements != nil {
		row.Achievements = *req.Achievements
	}
	if req.MentorID != nil {
		id, ok := parseOptionalUUIDField(req.MentorID)
		if !ok {
			return response.NewError(response.CodeBadRequest, "无效的指导老师")
		}
		row.MentorID = id
	}
	return nil
}
