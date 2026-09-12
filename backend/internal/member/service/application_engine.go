package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Yogdunana/StarByte/backend/internal/member/model"
	"github.com/Yogdunana/StarByte/backend/internal/member/repo"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/engine"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

const (
	engineStageOfficer   = "干事审批"
	engineStageMinister  = "部长审批"
	engineStagePresident = "社长审批"
)

func applicationVariables(app *model.MemberApplication) map[string]interface{} {
	vars := map[string]interface{}{
		"application_id": app.ID.String(),
		"applicant":      app.UserID.String(),
		"applicant_type": app.Type,
		"apply_type":     app.Type,
		"real_name":      app.RealName,
		"student_no":     app.StudentNo,
		"revision":       app.AdmissionRevision,
	}
	if app.Type != model.ApplicantOfficer {
		vars[engine.SkipOfficerVariable] = true
	}
	if app.DepartmentID != nil {
		vars["department_id"] = app.DepartmentID.String()
		vars["department"] = app.DepartmentID.String()
	} else {
		vars[engine.SkipMinisterVariable] = true
	}
	return vars
}

func engineStageLabel(nodeID string) string {
	switch nodeID {
	case "officer":
		return engineStageOfficer
	case "president":
		return engineStagePresident
	case "minister":
		return engineStageMinister
	case "":
		return ""
	default:
		return nodeID
	}
}

func applyEngineOutcome(app *model.MemberApplication, action, nodeID string, completed bool, now time.Time) {
	app.UpdatedAt = now
	app.ReviewedAt = &now
	if action == actionReject {
		app.Status = model.AppRejected
		app.AdmissionStage = model.AdmissionRejected
		app.CurrentStage = stageLabel(model.AppRejected)
		return
	}
	if action == actionSupplement {
		app.Status = model.AppSupplement
		app.AdmissionStage = model.AdmissionSupplement
		app.CurrentStage = stageLabel(model.AppSupplement)
		return
	}
	if completed {
		app.Status = model.AppApproved
		app.AdmissionStage = model.AdmissionApproved
		app.CurrentStage = stageLabel(model.AppApproved)
		return
	}
	next := engineStageLabel(nodeID)
	if next != "" && app.CurrentStage != next {
		app.StageEnteredAt = now
	}
	app.AdmissionStage = model.AdmissionEngine
	app.CurrentStage = next
	if nodeID == "president" {
		app.Status = model.AppReviewing
		if app.CurrentStage == "" {
			app.CurrentStage = engineStagePresident
		}
		return
	}
	app.Status = model.AppPending
	if app.CurrentStage == "" {
		app.CurrentStage = engineStageMinister
	}
}

func skipMinisterNode(app *model.MemberApplication) bool {
	return app != nil && app.DepartmentID == nil
}

func engineReviewPermission(actor *model.AdmissionActor, app *model.MemberApplication, parent *uuid.UUID, role, comment string, now time.Time) error {
	if strings.TrimSpace(role) == "" {
		return admissionDenied("当前环节不可审批")
	}
	allowed, delegated := admissionAuthority(actor, app, parent, role)
	if !allowed {
		return admissionDenied("无权审批该入会申请")
	}
	// 部长节点只允许意向部门部长立即处理；社长/中心主任须走既有 24h 代签规则。
	if delegated && (strings.TrimSpace(comment) == "" || now.Before(app.StageEnteredAt.Add(24*time.Hour))) {
		return admissionDenied("超时24小时后上级才可代签，并须填写原因")
	}
	return nil
}

func engineNodeRole(nodeID string) string {
	return engine.ResolveApprovalRole(nodeID, "")
}

func (s *admissionService) tryEngineReview(ctx context.Context, viewer, id uuid.UUID, action, comment string, required []string) (bool, error) {
	if s.flow == nil {
		return false, nil
	}
	var handled bool
	var deliver func(context.Context)
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		store := repo.NewAdmissionRepo(tx)
		app, actor, parent, err := loadAdmission(ctx, store, viewer, id)
		if err != nil {
			return err
		}
		ok, err := s.isEngineChain(ctx, store, app)
		if err != nil || !ok {
			return err
		}
		handled = true
		deliver, err = s.runEngineReview(ctx, tx, store, app, actor, parent, viewer, action, comment, required)
		return err
	})
	if err == nil && deliver != nil {
		deliver(ctx)
		_ = s.refreshPermissions(ctx)
	}
	return handled, err
}

func (s *admissionService) isEngineChain(ctx context.Context, store repo.AdmissionRepo, app *model.MemberApplication) (bool, error) {
	if app.FlowInstanceID == nil {
		return false, nil
	}
	key, err := store.WorkflowDefinitionKey(ctx, *app.FlowInstanceID)
	if err != nil {
		return false, err
	}
	return key == engine.MemberApplicationDefinitionKey, nil
}

func (s *admissionService) runEngineReview(
	ctx context.Context, tx *gorm.DB, store repo.AdmissionRepo,
	app *model.MemberApplication, actor *model.AdmissionActor, parent *uuid.UUID,
	viewer uuid.UUID, action, comment string, required []string,
) (func(context.Context), error) {
	if app.HistoricalReviewRequired {
		return nil, response.NewError(response.CodeMemberAppInvalid, "历史申请须先核验，不能自动补签")
	}
	if action != actionApprove && action != actionReject && action != actionSupplement {
		return nil, response.NewError(response.CodeBadRequest, "无效的审批决定")
	}
	if action != actionApprove && comment == "" {
		return nil, response.NewError(response.CodeBadRequest, "拒绝或补充材料须填写原因")
	}
	flow, deliver, err := s.flow.BindTransaction(tx)
	if err != nil {
		return nil, err
	}
	nodeID, completed, err := flow.RunningApprovalNode(ctx, *app.FlowInstanceID)
	if err != nil {
		return nil, err
	}
	if completed {
		return nil, response.NewError(response.CodeConflict, "入会流程已结束")
	}
	role, err := flow.ApprovalRoleCode(ctx, *app.FlowInstanceID, nodeID)
	if err != nil {
		return nil, err
	}
	if err := engineReviewPermission(actor, app, parent, role, comment, s.now()); err != nil {
		return nil, err
	}
	from := app.Status
	now := s.now()
	if action == actionSupplement {
		if err := flow.Terminate(ctx, *app.FlowInstanceID, viewer, "正式审批：supplement"); err != nil {
			return nil, err
		}
		app.RequiredFields = nonemptyStrings(required)
		applyEngineOutcome(app, action, nodeID, false, now)
	} else {
		if err := flow.CompleteApplicationApproval(ctx, *app.FlowInstanceID, viewer, engine.TaskAction(action), comment); err != nil {
			return nil, err
		}
		nodeID, completed, err = flow.RunningApprovalNode(ctx, *app.FlowInstanceID)
		if err != nil && action != actionReject {
			return nil, err
		}
		if action == actionReject {
			completed = false
			nodeID = ""
		}
		applyEngineOutcome(app, action, nodeID, completed, now)
	}
	app.ReviewerID = &viewer
	app.ReviewComment = comment
	if err := store.SaveApplication(ctx, app); err != nil {
		return nil, err
	}
	if app.Status == model.AppApproved {
		if err := s.admitFromEngine(ctx, tx, app); err != nil {
			return nil, err
		}
	}
	members := &memberService{apps: repo.NewApplicationRepo(tx), profs: repo.NewProfileRepo(tx)}
	if err := members.recordAppHistory(ctx, app.ID, from, app.Status, &viewer, comment, map[string]interface{}{"action": action, "engine": true, "node": nodeID}); err != nil {
		return nil, err
	}
	return deliver, nil
}

func membershipRole(app *model.MemberApplication) string {
	if app.Type == model.ApplicantOfficer {
		return "officer"
	}
	return "member"
}

func (s *admissionService) admitFromEngine(ctx context.Context, tx *gorm.DB, app *model.MemberApplication) error {
	if err := ensureActiveApplicant(ctx, tx, app.UserID); err != nil {
		return err
	}
	members := &memberService{apps: repo.NewApplicationRepo(tx), profs: repo.NewProfileRepo(tx)}
	if err := members.ensureProfile(ctx, app); err != nil {
		return err
	}
	if err := repo.NewAdmissionMaintenanceRepo(tx).GrantRole(ctx, app.UserID, membershipRole(app), app.DepartmentID); err != nil {
		return err
	}
	return repo.NewAdmissionJobsRepo(tx).QueuePermissionRefresh(ctx, app.UserID)
}

func ensureActiveApplicant(ctx context.Context, tx *gorm.DB, user uuid.UUID) error {
	var status int
	err := tx.WithContext(ctx).Table("users").Where("id = ? AND deleted_at IS NULL", user).Select("status").Take(&status).Error
	if err == gorm.ErrRecordNotFound {
		return response.NewError(response.CodeMemberAppNotFound, "申请人账号不存在，无法完成录用")
	}
	if err != nil {
		return err
	}
	if status != 0 {
		return admissionDenied("申请人账号已停用，不能自动开通成员身份")
	}
	return nil
}
