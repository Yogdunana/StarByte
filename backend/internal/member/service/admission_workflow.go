package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Yogdunana/StarByte/backend/internal/member/model"
	"github.com/Yogdunana/StarByte/backend/internal/member/repo"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/engine"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

// NewAdmissionServiceWithWorkflow is used by production wiring. Existing
// historical applications remain readable without invented workflow history.
func NewAdmissionServiceWithWorkflow(db *gorm.DB, flow *engine.FlowEngine, cache AdmissionPermissionCache) AdmissionService {
	return &admissionService{db: db, now: time.Now, permissions: cache, flow: flow}
}
func (s *admissionService) startAdmissionWorkflow(ctx context.Context, tx *gorm.DB, id uuid.UUID) (func(context.Context), error) {
	if s.flow == nil {
		return nil, nil
	}
	store := repo.NewAdmissionRepo(tx)
	app, err := store.LockApplication(ctx, id)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, response.NewError(response.CodeMemberAppNotFound, "申请不存在")
	}
	flow, deliver, err := s.flow.BindTransaction(tx)
	if err != nil {
		return nil, err
	}
	inst, err := flow.Start(ctx, engine.MemberApplicationDefinitionKey, app.ID.String(), "member_application", app.UserID, applicationVariables(app))
	if err != nil {
		return nil, err
	}
	app.FlowInstanceID = &inst.ID
	app.AdmissionStage = model.AdmissionEngine
	app.StageEnteredAt = s.now()
	// skip_minister 在 Start 时已让引擎穿过部长节点，不会创建全员部长待办。
	if skipMinisterNode(app) {
		app.Status = model.AppReviewing
		app.CurrentStage = engineStagePresident
	} else {
		app.Status = model.AppPending
		app.CurrentStage = engineStageMinister
	}
	if err := store.SaveApplication(ctx, app); err != nil {
		return nil, err
	}
	return deliver, nil
}
func (s *admissionService) advanceWorkflow(ctx context.Context, tx *gorm.DB, app *model.MemberApplication, signature *model.AdmissionSignature) (func(context.Context), error) {
	if app.FlowInstanceID == nil || s.flow == nil {
		return nil, nil
	}
	flow, deliver, err := s.flow.BindTransaction(tx)
	if err != nil {
		return nil, err
	}
	if signature.Decision != "approve" {
		// A request for more material closes this revision; resubmission starts a new
		// instance. Earlier signatures/tasks stay immutable and separately traceable.
		err = flow.Terminate(ctx, *app.FlowInstanceID, signature.SignerID, "正式审批："+signature.Decision)
		return deliver, err
	}
	if err := flow.CompleteBusinessApproval(ctx, *app.FlowInstanceID, signature.Stage, signature.SignerRole, signature.SignerID, signature.ID); err != nil {
		return nil, err
	}
	stage, completed, err := flow.BusinessStage(ctx, *app.FlowInstanceID)
	if err != nil {
		return nil, err
	}
	// The engine advances first; the policy state is a checked projection, never
	// an independent approval path capable of bypassing the configured graph.
	if completed {
		if app.AdmissionStage != model.AdmissionApproved && app.AdmissionStage != model.AdmissionProbation {
			return nil, response.NewError(response.CodeConflict, "流程提前结束，缺少必需签字")
		}
	} else if stage != app.AdmissionStage {
		return nil, response.NewError(response.CodeConflict, "流程尚未完成当前环节的全部签字")
	}
	return deliver, nil
}
