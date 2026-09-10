package service

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Yogdunana/StarByte/backend/internal/task/repo"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/engine"
	wfmodel "github.com/Yogdunana/StarByte/backend/internal/workflow/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

type transferApprover struct{ db *gorm.DB }

func NewTransferApprover(db *gorm.DB) engine.BusinessApprover { return &transferApprover{db} }
func (r *transferApprover) ForTransaction(tx *gorm.DB) engine.BusinessApprover {
	return &transferApprover{tx}
}
func (r *transferApprover) Resolve(ctx context.Context, inst *wfmodel.FlowInstance, node *engine.FlowNode) ([]uuid.UUID, error) {
	if inst.BusinessType != engine.TaskTransferBusinessType {
		return nil, response.NewError(response.CodeForbidden, "转办业务类型不匹配")
	}
	id, err := uuid.Parse(inst.BusinessKey)
	if err != nil {
		return nil, err
	}
	store := repo.NewTransferRepo(r.db)
	request, err := store.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if request == nil || request.InitiatorID != inst.InitiatorID || request.Status != "pending" {
		return nil, response.NewError(response.CodeConflict, "转办请求与流程不一致")
	}
	actors, err := store.Actors(ctx)
	if err != nil {
		return nil, err
	}
	role, _ := node.Config["transferRole"].(string)
	ids := []uuid.UUID{}
	for _, actor := range actors {
		if actual, _ := transferAuthority(&actor, request, role); actual != "" {
			ids = append(ids, actor.ID)
		}
	}
	if len(ids) == 0 {
		return nil, response.NewError(response.CodeConflict, "转办缺少可用负责人，请配置双方部门及中心的有效职务人员")
	}
	return ids, nil
}
