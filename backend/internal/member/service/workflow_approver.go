package service

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Yogdunana/StarByte/backend/internal/member/repo"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/engine"
	wfmodel "github.com/Yogdunana/StarByte/backend/internal/workflow/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

type admissionApprover struct{ db *gorm.DB }

func NewAdmissionApprover(db *gorm.DB) engine.BusinessApprover { return &admissionApprover{db} }
func (r *admissionApprover) ForTransaction(db *gorm.DB) engine.BusinessApprover {
	return &admissionApprover{db}
}
func (r *admissionApprover) Resolve(ctx context.Context, inst *wfmodel.FlowInstance, node *engine.FlowNode) ([]uuid.UUID, error) {
	if inst.BusinessType != "member_application" {
		return nil, admissionDenied("入会签字节点只能用于真实入会申请")
	}
	id, err := uuid.Parse(inst.BusinessKey)
	if err != nil {
		return nil, err
	}
	store := repo.NewAdmissionRepo(r.db)
	app, err := store.LockApplication(ctx, id)
	if err != nil {
		return nil, err
	}
	if app == nil || app.UserID != inst.InitiatorID {
		return nil, admissionDenied("入会申请与流程发起人不一致")
	}
	var parent *uuid.UUID
	if app.DepartmentID != nil {
		parent, err = store.ParentDepartment(ctx, *app.DepartmentID)
		if err != nil {
			return nil, err
		}
	}
	role, _ := node.Config["admissionRole"].(string)
	reviewers, err := repo.NewAdmissionJobsRepo(r.db).Reviewers(ctx)
	if err != nil {
		return nil, err
	}
	ids := []uuid.UUID{}
	for _, actor := range reviewers {
		// Superiors receive a visible task, but actual signing still enforces 24 hours,
		// interview completion and a delegation reason in AdmissionService.Sign.
		if allowed, _ := admissionAuthority(&actor, app, parent, role); allowed {
			ids = append(ids, actor.ID)
		}
	}
	if len(ids) == 0 {
		return nil, response.NewError(response.CodeMemberAppInvalid, "缺少可用的审批负责人，请先配置对应职务人员")
	}
	return ids, nil
}
