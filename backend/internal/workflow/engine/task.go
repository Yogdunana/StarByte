package engine

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/model"
	"github.com/Yogdunana/StarByte/backend/pkg/events"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

// CompleteTask processes a task completion and continues the flow.
func (e *FlowEngine) completeTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID, action TaskAction, comment string, formData map[string]interface{}) error {
	if err := validateInputVariables(formData); err != nil {
		return response.NewAppError(response.CodeBadRequest, err.Error())
	}
	// 1. Get the task.
	task, err := e.taskRepo.GetTaskByID(ctx, taskID)
	if err != nil {
		return response.NewAppErrorf(response.CodeWorkflowTaskNotFnd,
			"failed to query task: %v", err)
	}
	if task == nil {
		return response.NewAppError(response.CodeWorkflowTaskNotFnd,
			"流程任务不存在")
	}
	if task.Status != 0 {
		return response.NewAppError(response.CodeWorkflowTaskStatus,
			"流程任务状态不允许操作")
	}

	// 2. Check permission (assignee or initiator for withdraw).
	if action != ActionWithdraw {
		if task.AssigneeID == nil || *task.AssigneeID != userID {
			return response.NewAppError(response.CodeWorkflowTaskNoAccess,
				"无权操作流程任务")
		}
	}

	// 3. Get the instance.
	inst, err := e.instRepo.GetByID(ctx, task.InstanceID)
	if err != nil {
		return response.NewAppErrorf(response.CodeWorkflowInstNotFound,
			"failed to query instance: %v", err)
	}
	if inst == nil {
		return response.NewAppError(response.CodeWorkflowInstNotFound,
			"流程实例不存在")
	}
	if inst.Status != 0 {
		return response.NewAppError(response.CodeWorkflowInstStatus,
			"流程实例状态不允许操作")
	}

	if action == ActionWithdraw && inst.InitiatorID != userID {
		return response.NewAppError(response.CodeWorkflowTaskNoAccess, "只有发起人可以撤回流程")
	}
	if action != ActionApprove && action != ActionReject && action != ActionWithdraw {
		return response.NewAppError(response.CodeBadRequest, "转办或回退请使用专用操作接口")
	}
	// 4. Get the definition version to parse the graph.
	version, err := e.defRepo.GetVersionByID(ctx, inst.DefinitionVersionID)
	if err != nil || version == nil {
		return response.NewAppError(response.CodeWorkflowVerNotFound,
			"流程版本不存在")
	}

	graph, err := ParseGraph(version.BpmnData)
	if err != nil {
		return response.NewAppErrorf(response.CodeWorkflowInvalidNode,
			"failed to parse graph: %v", err)
	}

	node := graph.GetNode(task.NodeID)
	if node == nil {
		return response.NewAppError(response.CodeWorkflowNodeNotFound,
			"流程节点不存在: "+task.NodeID)
	}

	// 5. Update task status within a transaction.
	now := time.Now()
	switch action {
	case ActionApprove:
		task.Status = 1
	case ActionReject:
		task.Status = 2
	case ActionTransfer:
		task.Status = 3
	case ActionRollback:
		task.Status = 4
	case ActionWithdraw:
		task.Status = 4
	}
	task.Action = string(action)
	task.Comment = comment
	task.CompletedAt = &now
	task.UpdatedAt = now

	// Merge formData into task.
	if len(formData) > 0 {
		formBytes, err := json.Marshal(formData)
		if err != nil {
			return response.NewAppError(response.CodeBadRequest, "审批表单数据无法序列化")
		}
		task.FormData = formBytes
	}

	txErr := e.withTransaction(ctx, func(tx *gorm.DB) error {
		if err := e.taskRepo.UpdateTask(ctx, tx, task); err != nil {
			return err
		}

		// Record history.
		hist := &model.FlowHistory{
			ID:         uuid.New(),
			InstanceID: inst.ID,
			TaskID:     &task.ID,
			NodeID:     node.ID,
			NodeName:   node.Label,
			NodeType:   node.Type,
			OperatorID: &userID,
			Action:     string(action),
			Comment:    comment,
			CreatedAt:  now,
		}
		if err := e.taskRepo.CreateHistory(ctx, tx, hist); err != nil {
			return err
		}

		// Merge form data into variables.
		if len(formData) > 0 {
			if err := e.varRepo.SetMap(ctx, tx, inst.ID, formData); err != nil {
				return err
			}
		}

		return nil
	})
	if txErr != nil {
		return txErr
	}

	// 6. Publish TaskCompletedEvent.
	e.eventBus.Publish(ctx, events.TaskCompletedEvent{
		InstanceID: inst.ID,
		TaskID:     task.ID,
		OperatorID: userID,
		Action:     string(action),
		Comment:    comment,
	})

	if action == ActionWithdraw {
		return e.terminateInstance(ctx, inst, userID, "发起人撤回")
	}
	pass, impossible, err := e.resolveApproval(ctx, task, node, userID)
	if err != nil {
		return err
	}
	if impossible {
		return e.terminateInstance(ctx, inst, userID, "审批未通过")
	}
	if !pass {
		return nil
	}

	// 8. Continue execution from the next node.
	if pass {
		vars, err := e.varRepo.GetMap(ctx, inst.ID)
		if err != nil {
			return err
		}
		nextNodes, err := e.executeNode(ctx, inst, graph, node, vars)
		if err != nil {
			return err
		}
		if len(nextNodes) > 0 {
			current := []string{}
			for _, id := range GetCurrentNodeIDs(inst.CurrentNodeIDs) {
				if id != node.ID {
					current = append(current, id)
				}
			}
			if err := e.updateCurrentNodes(ctx, inst, current); err != nil {
				return err
			}
			return e.executeFromNodes(ctx, inst, graph, nextNodes, vars, node.ID)
		}
	}

	return nil
}
