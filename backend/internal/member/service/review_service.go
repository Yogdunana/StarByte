package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/Yogdunana/StarByte/backend/internal/member/dto"
	"github.com/Yogdunana/StarByte/backend/internal/member/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

func (s *memberService) Approve(ctx context.Context, reviewer, id uuid.UUID, comment string) (*dto.ApplicationResponse, error) {
	if s.admission != nil {
		if err := s.admission.ReviewMaterials(ctx, reviewer, id, actionApprove, comment); err != nil {
			return nil, err
		}
		return s.applicationResponse(ctx, id)
	}
	return s.transit(ctx, reviewer, id, actionApprove, comment, nil)
}

func (s *memberService) Reject(ctx context.Context, reviewer, id uuid.UUID, comment string) (*dto.ApplicationResponse, error) {
	if s.admission != nil {
		if err := s.admission.ReviewMaterials(ctx, reviewer, id, actionReject, comment); err != nil {
			return nil, err
		}
		return s.applicationResponse(ctx, id)
	}
	return s.transit(ctx, reviewer, id, actionReject, comment, nil)
}

func (s *memberService) Supplement(ctx context.Context, reviewer, id uuid.UUID, req *dto.SupplementRequest) (*dto.ApplicationResponse, error) {
	if s.admission != nil {
		if err := s.admission.RequestSupplement(ctx, reviewer, id, req); err != nil {
			return nil, err
		}
		return s.applicationResponse(ctx, id)
	}
	return s.transit(ctx, reviewer, id, actionSupplement, req.Comment, req.RequiredFields)
}

func (s *memberService) transit(ctx context.Context, reviewer, id uuid.UUID, action, comment string, required []string) (*dto.ApplicationResponse, error) {
	app, err := s.apps.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get application: %w", err)
	}
	if app == nil {
		return nil, response.NewError(response.CodeMemberAppNotFound, "申请不存在")
	}
	next, ok := nextStatus(app.Type, app.Status, action)
	if !ok {
		return nil, response.NewError(response.CodeMemberAppInvalid, "当前状态不允许该操作")
	}

	from := app.Status
	now := time.Now()
	app.Status = next
	app.CurrentStage = stageLabel(next)
	app.ReviewerID = &reviewer
	app.ReviewComment = comment
	app.ReviewedAt = &now
	app.UpdatedAt = now
	if action == actionSupplement {
		app.RequiredFields = nonemptyStrings(required)
	}

	if next == model.AppInterviewing {
		if err := s.startInterview(ctx, app, reviewer); err != nil {
			return nil, err
		}
	}
	if next == model.AppApproved {
		if err := s.ensureProfile(ctx, app); err != nil {
			return nil, err
		}
	}
	if err := s.apps.Update(ctx, app); err != nil {
		return nil, fmt.Errorf("update application: %w", err)
	}
	extra := map[string]any{"action": action}
	if len(required) > 0 {
		extra["required_fields"] = required
	}
	if err := s.recordAppHistory(ctx, app.ID, from, next, &reviewer, comment, extra); err != nil {
		return nil, err
	}
	return s.applicationResponse(ctx, app.ID)
}

// SyncFromInterview records a public progress event only. Interview scores and
// recommendations can never approve, reject or sign an admission decision.
func (s *memberService) SyncFromInterview(ctx context.Context, operator, applicationID uuid.UUID, result int16, comment string) error {
	app, err := s.apps.GetByID(ctx, applicationID)
	if err != nil {
		return fmt.Errorf("get application: %w", err)
	}
	if app == nil || app.Status != model.AppInterviewing {
		return nil
	}
	return s.recordAppHistory(ctx, app.ID, app.Status, app.Status, &operator, "面试结果已记录，等待正式签字审批", map[string]interface{}{"event": "interview_result_recorded"})
}

func (s *memberService) applicationResponse(ctx context.Context, id uuid.UUID) (*dto.ApplicationResponse, error) {
	row, err := s.apps.GetByIDWithNames(ctx, id)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, response.NewError(response.CodeMemberAppNotFound, "申请不存在")
	}
	return mapApplication(row), nil
}

func (s *memberService) startInterview(ctx context.Context, app *model.MemberApplication, reviewer uuid.UUID) error {
	if s.starter == nil {
		return response.NewError(response.CodeMemberAppInvalid, "面试流程引擎未配置")
	}
	vars := map[string]interface{}{
		"application_id": app.ID.String(),
		"applicant_type": int(app.Type),
		"real_name":      app.RealName,
		"student_no":     app.StudentNo,
	}
	instID, err := s.starter.StartOfficerInterview(ctx, app.ID, reviewer, vars)
	if err != nil {
		return err
	}
	app.FlowInstanceID = &instID
	return nil
}

func (s *memberService) ensureProfile(ctx context.Context, app *model.MemberApplication) error {
	exist, err := s.profs.GetByUserID(ctx, app.UserID)
	if err != nil {
		return fmt.Errorf("get profile by user: %w", err)
	}
	now := time.Now()
	memberType := model.MemberTypeMember
	if app.Type == model.ApplicantOfficer {
		memberType = model.MemberTypeOfficer
	}
	if exist != nil {
		if exist.Status == model.ProfileDisabled {
			return admissionDenied("档案已停用，不能通过录用恢复权限")
		}
		if exist.Status == model.ProfileActive && exist.MemberType >= memberType {
			return response.NewError(response.CodeMemberAppDuplicate, "已具有该成员身份，不能重复录用或覆盖现有职务")
		}
		exist.Status = model.ProfileActive
		exist.LeaveDate = nil
		exist.MemberType = memberType
		if app.AdmissionStage == model.AdmissionProbation {
			exist.Status = model.ProfileProbation
		}
		exist.DepartmentID = app.DepartmentID
		if exist.RealName == "" {
			exist.RealName = app.RealName
		}
		if exist.StudentNo == "" {
			exist.StudentNo = app.StudentNo
		}
		exist.UpdatedAt = now
		return s.profs.Update(ctx, exist)
	}
	dup, err := s.profs.GetByStudentNo(ctx, app.StudentNo, nil)
	if err != nil {
		return fmt.Errorf("check student no: %w", err)
	}
	if dup != nil {
		return response.NewError(response.CodeMemberStudentExists, "学号已存在")
	}
	join := now
	profile := &model.MemberProfile{
		ID:           uuid.New(),
		UserID:       app.UserID,
		RealName:     app.RealName,
		StudentNo:    app.StudentNo,
		DepartmentID: app.DepartmentID,
		MemberType:   memberType,
		Status:       model.ProfileActive,
		JoinDate:     &join,
		Skills:       nonemptyStrings(app.Skills),
		Projects:     model.JSONProjects{},
		ContactPhone: app.ContactPhone,
		ContactEmail: app.ContactEmail,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if app.AdmissionStage == model.AdmissionProbation {
		profile.Status = model.ProfileProbation
	}
	if err := s.profs.Create(ctx, profile); err != nil {
		return fmt.Errorf("create profile: %w", err)
	}
	return nil
}
