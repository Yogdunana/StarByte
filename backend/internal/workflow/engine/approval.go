package engine

import (
	"context"
	"math"
	"time"

	"github.com/google/uuid"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

func ValidateApprovalConfig(config map[string]interface{}) error {
	kind, _ := config["approvalType"].(string)
	switch kind {
	case "", "single", "all", "any":
	case "ratio":
		ratio, ok := config["passRatio"].(float64)
		if !ok || ratio <= 0 || ratio > 100 || math.IsNaN(ratio) {
			return response.NewAppError(response.CodeWorkflowInvalidNode, "审批比例必须大于 0 且不超过 100")
		}
	default:
		return response.NewAppError(response.CodeWorkflowInvalidNode, "不支持的审批方式")
	}
	return nil
}

// approvalDecision returns pass, impossible. A rejection in all/single is a veto;
// any/ratio waits while the remaining undecided votes can still meet the threshold.
func approvalDecision(config map[string]interface{}, approved, rejected, pending int) (bool, bool) {
	total := approved + rejected + pending
	if total == 0 {
		return false, false
	}
	required := total
	switch config["approvalType"] {
	case "any":
		required = 1
	case "ratio":
		ratio, _ := config["passRatio"].(float64)
		required = int(math.Ceil(float64(total) * ratio / 100))
	}
	if required < 1 {
		required = total
	}
	return approved >= required, approved+pending < required
}

func (e *FlowEngine) resolveApproval(ctx context.Context, task *model.FlowTask, node *FlowNode, user uuid.UUID) (bool, bool, error) {
	if err := ValidateApprovalConfig(node.Config); err != nil {
		return false, false, err
	}
	tasks, err := e.taskRepo.ListTasksByInstance(ctx, task.InstanceID)
	if err != nil {
		return false, false, err
	}
	group := []model.FlowTask{}
	for _, item := range tasks {
		if item.NodeID != task.NodeID {
			continue
		}
		if task.ActivationID != nil {
			if item.ActivationID == nil || *item.ActivationID != *task.ActivationID {
				continue
			}
		} else if item.ID != task.ID {
			// Historical multi-task activations cannot be safely inferred.
			if item.Status == 0 {
				return false, false, response.NewAppError(response.CodeWorkflowTaskStatus, "历史多审批任务缺少轮次记录，请由管理员核验后重新发起")
			}
			continue
		}
		group = append(group, item)
	}
	if len(group) == 0 {
		group = append(group, *task)
	}
	approved, rejected, pending := 0, 0, 0
	for _, item := range group {
		switch item.Status {
		case 0:
			pending++
		case 1:
			approved++
		case 2:
			rejected++
		}
	}
	pass, impossible := approvalDecision(node.Config, approved, rejected, pending)
	if !pass {
		return false, impossible, nil
	}
	now := time.Now()
	for _, item := range group {
		if item.Status != 0 {
			continue
		}
		item.Status, item.Action, item.Comment = 5, "cancel", "审批条件已满足，关闭剩余待办"
		item.CompletedAt, item.UpdatedAt = &now, now
		if err := e.taskRepo.UpdateTask(ctx, nil, &item); err != nil {
			return false, false, err
		}
		if err := e.taskRepo.CreateHistory(ctx, nil, &model.FlowHistory{ID: uuid.New(), InstanceID: task.InstanceID, TaskID: &item.ID, NodeID: node.ID, NodeName: node.Label, NodeType: node.Type, OperatorID: &user, Action: "cancel", Comment: item.Comment, CreatedAt: now}); err != nil {
			return false, false, err
		}
	}
	return true, false, nil
}
