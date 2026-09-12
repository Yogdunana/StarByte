package service

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Yogdunana/StarByte/backend/internal/leave/repo"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/engine"
	wfmodel "github.com/Yogdunana/StarByte/backend/internal/workflow/model"
	wfrepo "github.com/Yogdunana/StarByte/backend/internal/workflow/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

type leaveApprover struct{ db *gorm.DB }

func NewLeaveApprover(db *gorm.DB) engine.BusinessApprover { return &leaveApprover{db} }
func (r *leaveApprover) ForTransaction(tx *gorm.DB) engine.BusinessApprover {
	return &leaveApprover{tx}
}

func (r *leaveApprover) Resolve(ctx context.Context, inst *wfmodel.FlowInstance, node *engine.FlowNode) ([]uuid.UUID, error) {
	if inst.BusinessType != engine.LeaveBusinessType {
		return nil, response.NewError(response.CodeForbidden, "请假审批类型不匹配")
	}
	id, err := uuid.Parse(inst.BusinessKey)
	if err != nil {
		return nil, err
	}
	app, err := repo.New(r.db).GetLeaveApplicationByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if app == nil || app.ApplicantID != inst.InitiatorID {
		return nil, response.NewError(response.CodeConflict, "请假申请与流程发起人不一致")
	}
	stage, _ := node.Config["leaveStage"].(string)
	approvers := wfrepo.NewApproverRepo(r.db)
	var ids []uuid.UUID
	switch stage {
	case "department":
		ids, err = approvers.DepartmentLeaders(ctx, app.ApplicantID)
	case "org":
		ids, err = approvers.ByRoleCode(ctx, "president", nil)
		if err != nil {
			return nil, err
		}
		extra, extraErr := approvers.ByRoleCode(ctx, "vice_president", nil)
		if extraErr != nil {
			return nil, extraErr
		}
		ids = append(ids, extra...)
	default:
		return nil, response.NewError(response.CodeBadRequest, "未知请假审批环节")
	}
	if err != nil {
		return nil, err
	}
	out := uniqueExcept(ids, app.ApplicantID)
	if len(out) == 0 {
		return nil, response.NewError(response.CodeLeaveWorkflow, "缺少可用的请假审批人，请先配置部长或社长")
	}
	return out, nil
}

func uniqueExcept(ids []uuid.UUID, skip uuid.UUID) []uuid.UUID {
	seen := map[uuid.UUID]struct{}{}
	out := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if id == skip || id == uuid.Nil {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}
