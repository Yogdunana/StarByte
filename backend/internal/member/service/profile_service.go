package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/member/dto"
	"github.com/Yogdunana/StarByte/backend/internal/member/model"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
)

func (s *memberService) ListProfiles(ctx context.Context, viewer uuid.UUID, req *dto.ListProfileRequest, scope *rbacModel.DataScopeCondition) ([]*dto.ProfileResponse, int64, error) {
	rows, total, err := s.profs.List(ctx, req, rewriteScope(scope, "p", viewer))
	if err != nil {
		return nil, 0, fmt.Errorf("list profiles: %w", err)
	}
	return mapProfiles(rows), total, nil
}

func (s *memberService) GetProfile(ctx context.Context, viewer, id uuid.UUID, scope *rbacModel.DataScopeCondition) (*dto.ProfileResponse, error) {
	row, err := s.loadVisibleProfile(ctx, viewer, id, scope)
	if err != nil {
		return nil, err
	}
	return mapProfile(row), nil
}

func (s *memberService) UpdateProfile(ctx context.Context, operator, id uuid.UUID, req *dto.UpdateProfileRequest, scope *rbacModel.DataScopeCondition) (*dto.ProfileResponse, error) {
	row, err := s.loadVisibleProfile(ctx, operator, id, scope)
	if err != nil {
		return nil, err
	}
	before := row.MemberProfile
	applyProfileUpdate(&row.MemberProfile, req)
	row.UpdatedAt = time.Now()
	if err := s.profs.Update(ctx, &row.MemberProfile); err != nil {
		return nil, fmt.Errorf("update profile: %w", err)
	}
	diffs := profileDiffs(before, row.MemberProfile, operator)
	if err := s.profs.CreateHistories(ctx, diffs); err != nil {
		return nil, fmt.Errorf("write profile history: %w", err)
	}
	return s.GetProfile(ctx, operator, id, scope)
}

func (s *memberService) UpdateProfileStatus(ctx context.Context, operator, id uuid.UUID, req *dto.UpdateProfileStatusRequest) (*dto.ProfileResponse, error) {
	p, err := s.profs.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get profile: %w", err)
	}
	if p == nil {
		return nil, response.NewError(response.CodeMemberProfileGone, "档案不存在")
	}
	from := p.Status
	p.Status = req.Status
	now := time.Now()
	if req.Status == model.ProfileLeft && p.LeaveDate == nil {
		p.LeaveDate = &now
	}
	if req.Status == model.ProfileActive {
		p.LeaveDate = nil
	}
	p.UpdatedAt = now
	if err := s.profs.Update(ctx, p); err != nil {
		return nil, fmt.Errorf("update profile status: %w", err)
	}
	diffs := []model.ProfileHistory{newFieldHistory(id, operator, "status", fmt.Sprintf("%d", from), fmt.Sprintf("%d", req.Status), req.Reason)}
	if err := s.profs.CreateHistories(ctx, diffs); err != nil {
		return nil, fmt.Errorf("write profile history: %w", err)
	}
	return s.GetProfile(ctx, operator, id, nil)
}

func (s *memberService) ProfileHistory(ctx context.Context, viewer, id uuid.UUID, scope *rbacModel.DataScopeCondition) ([]dto.ProfileHistoryResponse, error) {
	if _, err := s.loadVisibleProfile(ctx, viewer, id, scope); err != nil {
		return nil, err
	}
	rows, err := s.profs.ListHistory(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("list profile history: %w", err)
	}
	out := make([]dto.ProfileHistoryResponse, 0, len(rows))
	for _, h := range rows {
		item := dto.ProfileHistoryResponse{
			ID:        h.ID.String(),
			FieldName: h.FieldName,
			OldValue:  h.OldValue,
			NewValue:  h.NewValue,
			Reason:    h.Reason,
			CreatedAt: h.CreatedAt,
		}
		if h.OperatorID != nil {
			item.OperatorID = h.OperatorID.String()
		}
		out = append(out, item)
	}
	return out, nil
}

func (s *memberService) loadVisibleProfile(ctx context.Context, viewer, id uuid.UUID, scope *rbacModel.DataScopeCondition) (*model.ProfileWithNames, error) {
	row, err := s.profs.GetByIDWithNames(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get profile: %w", err)
	}
	if row == nil {
		return nil, response.NewError(response.CodeMemberProfileGone, "档案不存在")
	}
	if !canAccessRecord(scope, row.UserID, row.DepartmentID, viewer) {
		return nil, response.NewError(response.CodeMemberProfileDenied, "无权操作该档案")
	}
	return row, nil
}

func applyProfileUpdate(p *model.MemberProfile, req *dto.UpdateProfileRequest) {
	if req.RealName != "" {
		p.RealName = req.RealName
	}
	if req.Gender != nil {
		p.Gender = *req.Gender
	}
	if req.Grade != "" {
		p.Grade = req.Grade
	}
	if req.Major != "" {
		p.Major = req.Major
	}
	if req.Skills != nil {
		p.Skills = model.JSONStrings(req.Skills)
	}
	if req.Projects != nil {
		items := make(model.JSONProjects, 0, len(req.Projects))
		for _, it := range req.Projects {
			items = append(items, model.ProjectItem{Name: it.Name, Role: it.Role, Period: it.Period})
		}
		p.Projects = items
	}
	if req.Bio != "" {
		p.Bio = req.Bio
	}
	if req.ContactPhone != "" {
		p.ContactPhone = req.ContactPhone
	}
	if req.ContactEmail != "" {
		p.ContactEmail = req.ContactEmail
	}
}

func profileDiffs(before, after model.MemberProfile, operator uuid.UUID) []model.ProfileHistory {
	pairs := [][3]string{
		{"real_name", before.RealName, after.RealName},
		{"gender", fmt.Sprintf("%d", before.Gender), fmt.Sprintf("%d", after.Gender)},
		{"grade", before.Grade, after.Grade},
		{"major", before.Major, after.Major},
		{"bio", before.Bio, after.Bio},
		{"contact_phone", before.ContactPhone, after.ContactPhone},
		{"contact_email", before.ContactEmail, after.ContactEmail},
		{"skills", stringify(before.Skills), stringify(after.Skills)},
		{"projects", stringify(before.Projects), stringify(after.Projects)},
	}
	var out []model.ProfileHistory
	for _, p := range pairs {
		if p[1] != p[2] {
			out = append(out, newFieldHistory(after.ID, operator, p[0], p[1], p[2], ""))
		}
	}
	return out
}

func newFieldHistory(profileID, operator uuid.UUID, field, oldV, newV, reason string) model.ProfileHistory {
	h := model.ProfileHistory{
		ID:        uuid.New(),
		ProfileID: profileID,
		FieldName: field,
		OldValue:  oldV,
		NewValue:  newV,
		Reason:    reason,
		CreatedAt: time.Now(),
	}
	if operator != uuid.Nil {
		h.OperatorID = &operator
	}
	return h
}

func stringify(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(b)
}
