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

type taskApprover struct{ db *gorm.DB }

func NewTaskApprover(db *gorm.DB) engine.BusinessApprover                  { return &taskApprover{db} }
func (r *taskApprover) ForTransaction(tx *gorm.DB) engine.BusinessApprover { return &taskApprover{tx} }
func (r *taskApprover) Resolve(ctx context.Context, inst *wfmodel.FlowInstance, node *engine.FlowNode) ([]uuid.UUID, error) {
	if inst.BusinessType != engine.TaskBusinessType {
		return nil, response.NewError(response.CodeForbidden, "任务审批类型不匹配")
	}
	id, err := uuid.Parse(inst.BusinessKey)
	if err != nil {
		return nil, err
	}
	tasks := repo.NewTaskRepo(r.db)
	task, err := tasks.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if task == nil || task.CreatorID != inst.InitiatorID {
		return nil, response.NewError(response.CodeConflict, "任务与流程不一致")
	}
	actor := &task.CreatorID
	switch node.Config["taskStage"] {
	case "assignment":
	case "execution":
		actor = task.AssigneeID
	case "review":
		actor = task.ReviewerID
	case "acceptance":
		actor = task.AcceptorID
	default:
		return nil, response.NewError(response.CodeBadRequest, "未知任务审批环节")
	}
	if actor == nil {
		return nil, response.NewError(response.CodeTaskTargetGone, "任务环节缺少负责人")
	}
	user, err := tasks.GetUser(ctx, *actor)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, response.NewError(response.CodeTaskTargetGone, "任务负责人不存在或已停用")
	}
	return []uuid.UUID{*actor}, nil
}
