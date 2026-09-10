package engine

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/model"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

// TaskCheckpoint can only run inside the task service's transaction. It never
// synthesizes an approval on behalf of another assigned person.
func (e *FlowEngine) TaskCheckpoint(ctx context.Context, id uuid.UUID, stage string, actor uuid.UUID, action, comment string) error {
	task, graph, err := e.taskCheckpoint(ctx, id, stage, actor)
	if err != nil {
		return err
	}
	switch action {
	case "approve":
		return e.CompleteTask(ctx, task.ID, actor, ActionApprove, "任务业务确认（"+stage+"）", nil)
	case "return":
		if stage != "review" && stage != "acceptance" {
			return response.NewError(response.CodeBadRequest, "该环节不能退回执行")
		}
		if strings.TrimSpace(comment) == "" {
			return response.NewError(response.CodeBadRequest, "请填写返工要求")
		}
		target := ""
		for id, node := range graph.Nodes {
			if node.Config["taskStage"] == "execution" {
				target = id
			}
		}
		return e.RollbackTask(ctx, task.ID, actor, target, "退回执行；具体要求见任务交付记录")
	default:
		return response.NewError(response.CodeBadRequest, "无效的任务流程操作")
	}
}
func (e *FlowEngine) TransferTaskExecution(ctx context.Context, id, from, to uuid.UUID, reason string) error {
	task, _, err := e.taskCheckpoint(ctx, id, "execution", from)
	if err != nil {
		return err
	}
	return e.TransferTask(ctx, task.ID, from, to, reason)
}
func (e *FlowEngine) taskCheckpoint(ctx context.Context, id uuid.UUID, stage string, actor uuid.UUID) (*model.FlowTask, *FlowGraph, error) {
	if !e.businessTransaction || e.db == nil {
		return nil, nil, response.NewError(response.CodeForbidden, "任务操作需要业务事务")
	}
	if err := repo.NewRuntimeRepo(e.db).LockInstance(ctx, id); err != nil {
		return nil, nil, err
	}
	inst, err := e.instRepo.GetByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	if inst == nil || inst.BusinessType != TaskBusinessType || inst.Status != 0 {
		return nil, nil, response.NewError(response.CodeConflict, "任务流程当前不可处理")
	}
	version, err := e.defRepo.GetVersionByID(ctx, inst.DefinitionVersionID)
	if err != nil {
		return nil, nil, err
	}
	if version == nil {
		return nil, nil, response.NewError(response.CodeWorkflowVerNotFound, "任务流程版本不存在")
	}
	graph, err := ParseGraph(version.BpmnData)
	if err != nil {
		return nil, nil, err
	}
	active := GetCurrentNodeIDs(inst.CurrentNodeIDs)
	if len(active) != 1 || graph.GetNode(active[0]) == nil || graph.GetNode(active[0]).Config["taskStage"] != stage {
		return nil, nil, response.NewError(response.CodeConflict, "任务进度与流程环节不一致")
	}
	tasks, err := e.taskRepo.ListTasksByInstance(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	for i := range tasks {
		task := &tasks[i]
		if task.NodeID == active[0] && task.Status == 0 && task.AssigneeID != nil && *task.AssigneeID == actor {
			return task, graph, nil
		}
	}
	return nil, nil, response.NewError(response.CodeForbidden, "只有该环节的指定处理人可以操作")
}

// ReassignTaskExecution is reserved for a completed, authorized handover. Audit
// the actual approving operator; never impersonate the previous executor.
func (e *FlowEngine) ReassignTaskExecution(ctx context.Context, id, previous, next, operator uuid.UUID, reason string) error {
	task, _, err := e.taskCheckpoint(ctx, id, "execution", previous)
	if err != nil {
		return err
	}
	inst, err := e.instRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	return e.transferPendingTask(ctx, task, inst, operator, next, reason)
}
