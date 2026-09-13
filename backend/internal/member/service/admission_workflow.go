package service

import (
	"context"
	"encoding/json"
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
	var definitionID uuid.UUID
	if err := tx.Raw("SELECT id FROM flow_definitions WHERE key=? FOR UPDATE", engine.MemberApplicationDefinitionKey).Row().Scan(&definitionID); err != nil {
		return nil, err
	}
	var graph []byte
	if err := tx.Table("flow_definition_versions").Where("definition_id=? AND status=1", definitionID).Select("bpmn_data").Row().Scan(&graph); err != nil {
		return nil, err
	}
	var policy struct {
		CharterVersion int `json:"charterVersion"`
	}
	if err := json.Unmarshal(graph, &policy); err != nil {
		return nil, err
	}
	app.CharterPolicy = policy.CharterVersion == 1
	vars := applicationVariables(app)
	if app.CharterPolicy {
		app.ReviewDepartmentID = app.DepartmentID
		if app.ReviewDepartmentID == nil {
			var id uuid.UUID
			if err := tx.Table("departments").Where("code=? AND status=0", "admin").Select("id").Row().Scan(&id); err != nil {
				return nil, err
			}
			app.ReviewDepartmentID = &id
		}
		parent, err := store.ParentDepartment(ctx, *app.ReviewDepartmentID)
		if err != nil {
			return nil, err
		}
		if parent == nil {
			return nil, admissionDenied("审批缺少部门或中心范围")
		}
		vars["department_id"] = app.ReviewDepartmentID.String()
		vars["center_department_id"] = parent.String()
	}
	flow, deliver, err := s.flow.BindTransaction(tx)
	if err != nil {
		return nil, err
	}
	inst, err := flow.Start(ctx, engine.MemberApplicationDefinitionKey, app.ID.String(), "member_application", app.UserID, vars)
	if err != nil {
		return nil, err
	}
	app.FlowInstanceID = &inst.ID
	app.AdmissionStage = model.AdmissionEngine
	app.StageEnteredAt = s.now()
	nodeID, completed, err := flow.RunningApprovalNode(ctx, inst.ID)
	if err != nil {
		return nil, err
	}
	if completed {
		applyEngineOutcome(app, actionApprove, nodeID, true, false, s.now())
		if err := s.admitFromEngine(ctx, tx, app); err != nil {
			return nil, err
		}
	} else {
		last, err := lastEngineApproval(ctx, flow, inst.ID, nodeID)
		if err != nil {
			return nil, err
		}
		applyEngineOutcome(app, actionApprove, nodeID, false, last, s.now())
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
