package service

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Yogdunana/StarByte/backend/internal/member/dto"
	"github.com/Yogdunana/StarByte/backend/internal/member/model"
	"github.com/Yogdunana/StarByte/backend/internal/member/repo"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/engine"
	wfmodel "github.com/Yogdunana/StarByte/backend/internal/workflow/model"
	wfrepo "github.com/Yogdunana/StarByte/backend/internal/workflow/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

func (s *admissionService) TransferReview(ctx context.Context, viewer, id, target uuid.UUID, comment string) error {
	if s.flow == nil {
		return response.NewError(response.CodeBadRequest, "入会流程引擎未配置")
	}
	if target == uuid.Nil || target == viewer {
		return response.NewError(response.CodeBadRequest, "请选择其他有效处理人")
	}
	var deliver func(context.Context)
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		store := repo.NewAdmissionRepo(tx)
		app, actor, parent, err := loadAdmission(ctx, store, viewer, id)
		if err != nil {
			return err
		}
		ok, err := s.isEngineChain(ctx, store, app)
		if err != nil {
			return err
		}
		if !ok {
			return response.NewError(response.CodeMemberAppInvalid, "当前申请未接入可转交的流程引擎")
		}
		deliver, err = s.runEngineTransfer(ctx, tx, store, app, actor, parent, viewer, target, comment)
		return err
	})
	if err == nil && deliver != nil {
		deliver(ctx)
	}
	return err
}

func (s *admissionService) runEngineTransfer(
	ctx context.Context, tx *gorm.DB, store repo.AdmissionRepo,
	app *model.MemberApplication, actor *model.AdmissionActor, parent *uuid.UUID,
	viewer, target uuid.UUID, comment string,
) (func(context.Context), error) {
	if app.HistoricalReviewRequired {
		return nil, response.NewError(response.CodeMemberAppInvalid, "历史申请须先核验，不能转交")
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
	role, scoped, err := flow.ApprovalPolicy(ctx, *app.FlowInstanceID, nodeID)
	if err != nil {
		return nil, err
	}
	if err := engineReviewAccess(actor, app, parent, role, scoped, comment, s.now(), true); err != nil {
		return nil, err
	}
	if err := flow.TransferApplicationApproval(ctx, *app.FlowInstanceID, viewer, target, comment); err != nil {
		return nil, err
	}
	from := app.Status
	app.ReviewerID = &viewer
	app.ReviewComment = strings.TrimSpace(comment)
	app.UpdatedAt = s.now()
	if err := store.SaveApplication(ctx, app); err != nil {
		return nil, err
	}
	members := &memberService{apps: repo.NewApplicationRepo(tx), profs: repo.NewProfileRepo(tx)}
	if err := members.recordAppHistory(ctx, app.ID, from, app.Status, &viewer, comment, map[string]interface{}{"action": "transfer", "engine": true, "node": nodeID, "target": target.String()}); err != nil {
		return nil, err
	}
	return deliver, nil
}

func (s *admissionService) ApplicationProgress(ctx context.Context, viewer, id uuid.UUID, scope *rbacModel.DataScopeCondition) (*dto.ApplicationProgressResponse, error) {
	if s.flow == nil {
		return nil, response.NewError(response.CodeBadRequest, "入会流程引擎未配置")
	}
	store := repo.NewAdmissionRepo(s.db)
	app, err := repo.NewApplicationRepo(s.db).GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, response.NewError(response.CodeMemberAppNotFound, "申请不存在")
	}
	if err := s.canViewEngineProgress(ctx, store, app, viewer, scope); err != nil {
		return nil, err
	}
	ok, err := s.isEngineChain(ctx, store, app)
	if err != nil {
		return nil, err
	}
	if !ok || app.FlowInstanceID == nil {
		return nil, response.NewError(response.CodeMemberAppInvalid, "当前申请没有流程引擎进度")
	}
	progress, err := s.flow.ApplicationProgress(ctx, *app.FlowInstanceID)
	if err != nil {
		return nil, err
	}
	out := &dto.ApplicationProgressResponse{
		ApplicationID:  app.ID.String(),
		InstanceID:     progress.InstanceID.String(),
		Status:         progress.Status,
		CurrentNodeIDs: progress.CurrentNodeIDs,
		Completed:      progress.Completed,
		Terminated:     progress.Terminated,
		Editable:       !engineReviewClosed(app.Status),
	}
	for _, step := range progress.Steps {
		out.Steps = append(out.Steps, dto.ApplicationProgressStep{
			ID: step.ID, Label: step.Label, Type: step.Type, Role: step.Role,
			State: step.State, ApprovalType: step.ApprovalType, AllowTransfer: step.AllowTransfer,
		})
	}
	return out, nil
}

func (s *admissionService) TransferCandidates(ctx context.Context, viewer, id uuid.UUID, keyword string) ([]dto.TransferCandidate, error) {
	if s.flow == nil {
		return nil, response.NewError(response.CodeBadRequest, "入会流程引擎未配置")
	}
	store := repo.NewAdmissionRepo(s.db)
	app, actor, parent, err := loadAdmission(ctx, store, viewer, id)
	if err != nil {
		return nil, err
	}
	ok, err := s.isEngineChain(ctx, store, app)
	if err != nil {
		return nil, err
	}
	if !ok || app.FlowInstanceID == nil {
		return nil, response.NewError(response.CodeMemberAppInvalid, "当前申请不可转交")
	}
	nodeID, completed, err := s.flow.RunningApprovalNode(ctx, *app.FlowInstanceID)
	if err != nil {
		return nil, err
	}
	if completed {
		return nil, response.NewError(response.CodeConflict, "入会流程已结束")
	}
	role, scoped, err := s.flow.ApprovalPolicy(ctx, *app.FlowInstanceID, nodeID)
	if err != nil {
		return nil, err
	}
	if err := engineTransferPickerAccess(actor, app, parent, role, scoped, s.now()); err != nil {
		return nil, err
	}
	rows, err := wfrepo.NewApproverRepo(s.db).Search(ctx, keyword, viewer)
	if err != nil {
		return nil, err
	}
	out := make([]dto.TransferCandidate, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapTransferCandidate(row))
	}
	return out, nil
}

func (s *admissionService) canViewEngineProgress(ctx context.Context, store repo.AdmissionRepo, app *model.MemberApplication, viewer uuid.UUID, scope *rbacModel.DataScopeCondition) error {
	if app.UserID == viewer {
		return nil
	}
	// scope 来自进度路由上的可选 member 数据范围中间件（不强制 member:read）。
	if canAccessRecord(scope, app.UserID, app.DepartmentID, viewer) {
		return nil
	}
	actor, err := store.Actor(ctx, viewer)
	if err != nil {
		return err
	}
	if actor == nil || len(actor.Roles) == 0 {
		return admissionDenied("无权查看该申请审批进度")
	}
	var parent *uuid.UUID
	if app.DepartmentID != nil {
		parent, err = store.ParentDepartment(ctx, *app.DepartmentID)
		if err != nil {
			return err
		}
	}
	for _, policy := range progressReviewPolicies(ctx, s.flow, app) {
		if allowed, _ := admissionRoleAuthority(actor, app, parent, policy.Role, policy.DepartmentScope); allowed {
			return nil
		}
	}
	return admissionDenied("无权查看该申请审批进度")
}

func progressReviewPolicies(ctx context.Context, flow *engine.FlowEngine, app *model.MemberApplication) []engine.ApplicationRolePolicy {
	policies := []engine.ApplicationRolePolicy{
		{Role: "officer"}, {Role: "minister"}, {Role: "president"},
	}
	if flow == nil || app == nil || app.FlowInstanceID == nil {
		return policies
	}
	extra, err := flow.ApplicationApprovalPolicies(ctx, *app.FlowInstanceID)
	if err != nil {
		return policies
	}
	index := map[string]int{"officer": 0, "minister": 1, "president": 2}
	for _, policy := range extra {
		if policy.Role == "" {
			continue
		}
		if i, ok := index[policy.Role]; ok {
			if i >= 3 {
				policies[i].DepartmentScope = policies[i].DepartmentScope || policy.DepartmentScope
			}
			continue
		}
		index[policy.Role] = len(policies)
		policies = append(policies, policy)
	}
	return policies
}

func mapTransferCandidate(row wfmodel.ApproverOption) dto.TransferCandidate {
	return dto.TransferCandidate{ID: row.ID.String(), Name: row.Name, DepartmentName: row.DepartmentName}
}

func engineReviewClosed(status int16) bool {
	return status == model.AppApproved || status == model.AppRejected || status == model.AppSupplement
}
