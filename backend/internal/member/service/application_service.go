package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/member/dto"
	"github.com/Yogdunana/StarByte/backend/internal/member/model"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
)

func (s *memberService) Submit(ctx context.Context, userID uuid.UUID, req *dto.SubmitApplicationRequest) (*dto.ApplicationResponse, error) {
	open, err := s.apps.HasOpenApplication(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("check open application: %w", err)
	}
	if open {
		return nil, response.NewError(response.CodeMemberAppDuplicate, "已有待处理的入会申请")
	}
	now := time.Now()
	app := &model.MemberApplication{
		ID:             uuid.New(),
		UserID:         userID,
		Type:           int16(req.ApplicantType),
		RealName:       req.RealName,
		StudentNo:      req.StudentNo,
		DepartmentID:   parseOptionalUUID(req.DepartmentID),
		Reason:         req.Reason,
		Skills:         nonemptyStrings(req.Skills),
		Experience:     req.Experience,
		ContactPhone:   req.ContactPhone,
		ContactEmail:   req.ContactEmail,
		Status:         model.AppPending,
		CurrentStage:   stageLabel(model.AppPending),
		RequiredFields: model.JSONStrings{},
		SubmittedAt:    now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := s.apps.Create(ctx, app); err != nil {
		return nil, fmt.Errorf("create application: %w", err)
	}
	if err := s.recordAppHistory(ctx, app.ID, 0, model.AppPending, &userID, "提交申请", nil); err != nil {
		return nil, err
	}
	return s.GetApplication(ctx, userID, app.ID, nil)
}

func (s *memberService) Resubmit(ctx context.Context, userID, id uuid.UUID, req *dto.ResubmitApplicationRequest) (*dto.ApplicationResponse, error) {
	app, err := s.apps.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get application: %w", err)
	}
	if app == nil {
		return nil, response.NewError(response.CodeMemberAppNotFound, "申请不存在")
	}
	if app.UserID != userID {
		return nil, response.NewError(response.CodeMemberProfileDenied, "无权操作该申请")
	}
	next, ok := nextStatus(app.Type, app.Status, actionResubmit)
	if !ok {
		return nil, response.NewError(response.CodeMemberAppInvalid, "当前状态不允许补充提交")
	}
	if missing := missingRequired(app, req); len(missing) > 0 {
		return nil, response.NewError(response.CodeMemberFieldRequired, "请补充字段: "+joinFields(missing))
	}
	applyResubmit(app, req)
	from := app.Status
	app.Status = next
	app.CurrentStage = stageLabel(next)
	app.RequiredFields = nil
	app.UpdatedAt = time.Now()
	if err := s.apps.Update(ctx, app); err != nil {
		return nil, fmt.Errorf("update application: %w", err)
	}
	if err := s.recordAppHistory(ctx, app.ID, from, next, &userID, "补充材料后重新提交", nil); err != nil {
		return nil, err
	}
	return s.GetApplication(ctx, userID, app.ID, nil)
}

func (s *memberService) ListApplications(ctx context.Context, viewer uuid.UUID, req *dto.ListApplicationRequest, scope *rbacModel.DataScopeCondition) ([]*dto.ApplicationResponse, int64, error) {
	rows, total, err := s.apps.List(ctx, req, rewriteScope(scope, "a", viewer))
	if err != nil {
		return nil, 0, fmt.Errorf("list applications: %w", err)
	}
	return mapApplications(rows), total, nil
}

func (s *memberService) GetApplication(ctx context.Context, viewer, id uuid.UUID, scope *rbacModel.DataScopeCondition) (*dto.ApplicationResponse, error) {
	row, err := s.apps.GetByIDWithNames(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get application: %w", err)
	}
	if row == nil {
		return nil, response.NewError(response.CodeMemberAppNotFound, "申请不存在")
	}
	if row.UserID != viewer && !canAccessRecord(scope, row.UserID, row.DepartmentID, viewer) {
		return nil, response.NewError(response.CodeMemberProfileDenied, "无权查看该申请")
	}
	return mapApplication(row), nil
}

func (s *memberService) MyApplications(ctx context.Context, userID uuid.UUID) ([]*dto.ApplicationResponse, error) {
	rows, err := s.apps.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list my applications: %w", err)
	}
	return mapApplications(rows), nil
}

func (s *memberService) ApplicationHistory(ctx context.Context, viewer, id uuid.UUID, scope *rbacModel.DataScopeCondition) ([]dto.ApplicationHistoryResponse, error) {
	if _, err := s.GetApplication(ctx, viewer, id, scope); err != nil {
		return nil, err
	}
	rows, err := s.apps.ListHistory(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("list application history: %w", err)
	}
	out := make([]dto.ApplicationHistoryResponse, 0, len(rows))
	for _, h := range rows {
		item := dto.ApplicationHistoryResponse{
			ID:         h.ID.String(),
			FromStatus: h.FromStatus,
			ToStatus:   h.ToStatus,
			Comment:    h.Comment,
			CreatedAt:  h.CreatedAt,
		}
		if h.OperatorID != nil {
			item.OperatorID = h.OperatorID.String()
		}
		out = append(out, item)
	}
	return out, nil
}

func (s *memberService) ListDepartments(ctx context.Context) ([]dto.DepartmentOption, error) {
	rows, err := s.apps.ListDepartments(ctx)
	if err != nil {
		return nil, fmt.Errorf("list departments: %w", err)
	}
	out := make([]dto.DepartmentOption, 0, len(rows))
	for _, d := range rows {
		out = append(out, dto.DepartmentOption{ID: d.ID.String(), Name: d.Name})
	}
	return out, nil
}

func (s *memberService) recordAppHistory(ctx context.Context, appID uuid.UUID, from, to int16, operator *uuid.UUID, comment string, extra any) error {
	h := &model.ApplicationHistory{
		ID:            uuid.New(),
		ApplicationID: appID,
		FromStatus:    from,
		ToStatus:      to,
		OperatorID:    operator,
		Comment:       comment,
		Extra:         jsonExtra(extra),
		CreatedAt:     time.Now(),
	}
	if err := s.apps.CreateHistory(ctx, h); err != nil {
		return fmt.Errorf("create application history: %w", err)
	}
	return nil
}

func applyResubmit(app *model.MemberApplication, req *dto.ResubmitApplicationRequest) {
	if req.RealName != "" {
		app.RealName = req.RealName
	}
	if req.StudentNo != "" {
		app.StudentNo = req.StudentNo
	}
	if req.DepartmentID != "" {
		app.DepartmentID = parseOptionalUUID(req.DepartmentID)
	}
	if req.Reason != "" {
		app.Reason = req.Reason
	}
	if req.Skills != nil {
		app.Skills = model.JSONStrings(req.Skills)
	}
	if req.Experience != "" {
		app.Experience = req.Experience
	}
	if req.ContactPhone != "" {
		app.ContactPhone = req.ContactPhone
	}
	if req.ContactEmail != "" {
		app.ContactEmail = req.ContactEmail
	}
}

func missingRequired(app *model.MemberApplication, req *dto.ResubmitApplicationRequest) []string {
	merged := *app
	applyResubmit(&merged, req)
	var missing []string
	for _, field := range app.RequiredFields {
		if !fieldFilled(&merged, field) {
			missing = append(missing, field)
		}
	}
	return missing
}

func fieldFilled(app *model.MemberApplication, field string) bool {
	switch field {
	case "real_name":
		return app.RealName != ""
	case "student_no":
		return app.StudentNo != ""
	case "reason":
		return app.Reason != ""
	case "skills":
		return len(app.Skills) > 0
	case "experience":
		return app.Experience != ""
	case "contact_phone":
		return app.ContactPhone != ""
	case "contact_email":
		return app.ContactEmail != ""
	default:
		return true
	}
}

func nonemptyStrings(in []string) model.JSONStrings {
	if in == nil {
		return model.JSONStrings{}
	}
	return model.JSONStrings(in)
}

func joinFields(fields []string) string {
	out := ""
	for i, f := range fields {
		if i > 0 {
			out += ", "
		}
		out += f
	}
	return out
}
