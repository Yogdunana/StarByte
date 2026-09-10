package engine

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/model"
	"github.com/Yogdunana/StarByte/backend/pkg/events"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

// Terminate terminates a running flow instance.
func (e *FlowEngine) terminate(ctx context.Context, instanceID uuid.UUID, operatorID uuid.UUID, reason string) error {
	inst, err := e.instRepo.GetByID(ctx, instanceID)
	if err != nil || inst == nil {
		return response.NewAppError(response.CodeWorkflowInstNotFound,
			"流程实例不存在")
	}
	if inst.Status != 0 && inst.Status != 3 {
		return response.NewAppError(response.CodeWorkflowInstStatus,
			"流程实例状态不允许操作")
	}

	return e.terminateInstance(ctx, inst, operatorID, reason)
}

// Suspend suspends a running flow instance.
func (e *FlowEngine) suspend(ctx context.Context, instanceID uuid.UUID, operatorID uuid.UUID, reason string) error {
	inst, err := e.instRepo.GetByID(ctx, instanceID)
	if err != nil || inst == nil {
		return response.NewAppError(response.CodeWorkflowInstNotFound,
			"流程实例不存在")
	}
	if inst.Status != 0 {
		return response.NewAppError(response.CodeWorkflowInstStatus,
			"流程实例状态不允许操作")
	}

	inst.Status = 3 // suspended
	inst.TerminateReason = reason
	inst.UpdatedAt = time.Now()

	return e.instRepo.Update(ctx, nil, inst)
}

// Resume resumes a suspended flow instance.
func (e *FlowEngine) resume(ctx context.Context, instanceID uuid.UUID, operatorID uuid.UUID) error {
	inst, err := e.instRepo.GetByID(ctx, instanceID)
	if err != nil || inst == nil {
		return response.NewAppError(response.CodeWorkflowInstNotFound,
			"流程实例不存在")
	}
	if inst.Status != 3 {
		return response.NewAppError(response.CodeWorkflowInstStatus,
			"流程实例状态不允许操作")
	}

	inst.Status = 0 // running
	inst.TerminateReason = ""
	inst.UpdatedAt = time.Now()

	return e.instRepo.Update(ctx, nil, inst)
}

// completeInstance marks an instance as completed.
func (e *FlowEngine) completeInstance(ctx context.Context, inst *model.FlowInstance) error {
	tasks, err := e.taskRepo.ListTasksByInstance(ctx, inst.ID)
	if err != nil {
		return err
	}
	for _, task := range tasks {
		if task.Status == 0 {
			return response.NewAppError(response.CodeWorkflowTaskStatus, "仍有审批待办，不能结束流程")
		}
	}
	now := time.Now()
	inst.Status = 1 // completed
	inst.EndedAt = &now
	inst.UpdatedAt = now

	if err := e.instRepo.Update(ctx, nil, inst); err != nil {
		return err
	}

	// Publish FlowCompletedEvent.
	e.eventBus.Publish(ctx, events.FlowCompletedEvent{
		InstanceID: inst.ID,
		EndedAt:    now,
	})

	return nil
}

// terminateInstance marks an instance as terminated.
func (e *FlowEngine) terminateInstance(ctx context.Context, inst *model.FlowInstance, operatorID uuid.UUID, reason string) error {
	if err := e.cancelPendingTasks(ctx, inst.ID, operatorID, reason); err != nil {
		return err
	}
	inst.CurrentNodeIDs = []byte("[]")
	now := time.Now()
	inst.Status = 2 // terminated
	inst.EndedAt = &now
	inst.TerminateReason = reason
	inst.UpdatedAt = now

	if err := e.instRepo.Update(ctx, nil, inst); err != nil {
		return err
	}

	// Publish FlowTerminatedEvent.
	e.eventBus.Publish(ctx, events.FlowTerminatedEvent{
		InstanceID: inst.ID,
		Reason:     reason,
		OperatorID: operatorID,
		EndedAt:    now,
	})

	return nil
}

// updateCurrentNodes persists the current node IDs to the instance.
func (e *FlowEngine) updateCurrentNodes(ctx context.Context, inst *model.FlowInstance, nodeIDs []string) error {
	nodeBytes, err := json.Marshal(nodeIDs)
	if err != nil {
		e.logger.Error("failed to marshal current node IDs", zap.Error(err))
		return err
	}
	inst.CurrentNodeIDs = nodeBytes
	if err := e.instRepo.Update(ctx, nil, inst); err != nil {
		e.logger.Error("failed to update instance current nodes",
			zap.String("instance_id", inst.ID.String()),
			zap.Error(err))
		return err
	}
	return nil
}

// cancelPendingTasks closes every remaining branch in the same transaction.
func (e *FlowEngine) cancelPendingTasks(ctx context.Context, instanceID, operatorID uuid.UUID, reason string) error {
	tasks, err := e.taskRepo.ListTasksByInstance(ctx, instanceID)
	if err != nil {
		return err
	}
	now := time.Now()
	for _, task := range tasks {
		if task.Status != 0 {
			continue
		}
		task.Status = 5
		task.Action = "cancel"
		task.Comment = reason
		task.CompletedAt = &now
		task.UpdatedAt = now
		if err := e.taskRepo.UpdateTask(ctx, nil, &task); err != nil {
			return err
		}
		hist := &model.FlowHistory{ID: uuid.New(), InstanceID: instanceID, TaskID: &task.ID, NodeID: task.NodeID, NodeName: task.NodeName, NodeType: task.TaskType, Action: "cancel", Comment: reason, CreatedAt: now}
		if operatorID != uuid.Nil {
			hist.OperatorID = &operatorID
		}
		if err := e.taskRepo.CreateHistory(ctx, nil, hist); err != nil {
			return err
		}
	}
	return nil
}
