package engine

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/model"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/events"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

// operateTask serializes every mutation with approval and instance lifecycle operations.
func (e *FlowEngine) operateTask(ctx context.Context, taskID, userID uuid.UUID, operation func(*FlowEngine, *model.FlowTask, *model.FlowInstance) error) error {
	task, err := e.taskRepo.GetTaskByID(ctx, taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return response.NewAppError(response.CodeWorkflowTaskNotFnd, "流程任务不存在")
	}
	return e.transaction(ctx, task.InstanceID, func(bound *FlowEngine) error {
		current, err := bound.taskRepo.GetTaskByID(ctx, taskID)
		if err != nil {
			return err
		}
		if current == nil || current.Status != 0 {
			return response.NewAppError(response.CodeWorkflowTaskStatus, "流程任务已处理")
		}
		if current.AssigneeID == nil || *current.AssigneeID != userID {
			return response.NewAppError(response.CodeWorkflowTaskNoAccess, "无权操作流程任务")
		}
		instance, err := bound.instRepo.GetByID(ctx, current.InstanceID)
		if err != nil {
			return err
		}
		if instance == nil || instance.Status != 0 {
			return response.NewAppError(response.CodeWorkflowInstStatus, "流程当前不可处理")
		}
		return operation(bound, current, instance)
	})
}

func (e *FlowEngine) TransferTask(ctx context.Context, taskID, from, to uuid.UUID, comment string) error {
	if to == uuid.Nil || to == from {
		return response.NewAppError(response.CodeBadRequest, "请选择其他有效处理人")
	}
	return e.operateTask(ctx, taskID, from, func(bound *FlowEngine, task *model.FlowTask, inst *model.FlowInstance) error {
		return bound.transferPendingTask(ctx, task, inst, from, to, comment)
	})
}

func (e *FlowEngine) transferPendingTask(ctx context.Context, task *model.FlowTask, inst *model.FlowInstance, from, to uuid.UUID, comment string) error {
	if e.db == nil {
		return response.NewAppError(response.CodeInternalError, "处理人查询不可用")
	}
	active, err := repo.NewRuntimeRepo(e.db).ActiveUser(ctx, to)
	if err != nil {
		return err
	}
	if !active {
		return response.NewAppError(response.CodeBadRequest, "接收人不存在或已停用")
	}
	tasks, err := e.taskRepo.ListTasksByInstance(ctx, inst.ID)
	if err != nil {
		return err
	}
	for _, item := range tasks {
		sameActivation := task.ActivationID != nil && item.ActivationID != nil && *item.ActivationID == *task.ActivationID
		if sameActivation && item.ID != task.ID && item.Status <= 2 && item.AssigneeID != nil && *item.AssigneeID == to {
			return response.NewAppError(response.CodeBadRequest, "接收人已参与本轮审批，不能重复计票")
		}
	}
	now := time.Now()
	next := *task
	next.ID = uuid.New()
	next.AssigneeID = &to
	next.Action, next.Comment = "", ""
	next.ClaimedAt, next.CompletedAt = nil, nil
	next.CreatedAt, next.UpdatedAt = now, now
	task.Status, task.Action, task.Comment = 3, "transfer", comment
	task.CompletedAt, task.UpdatedAt = &now, now
	if err := e.taskRepo.UpdateTask(ctx, nil, task); err != nil {
		return err
	}
	if err := e.taskRepo.CreateTask(ctx, nil, &next); err != nil {
		return err
	}
	if err := e.taskRepo.CreateHistory(ctx, nil, &model.FlowHistory{ID: uuid.New(), InstanceID: inst.ID, TaskID: &task.ID, NodeID: task.NodeID, NodeName: task.NodeName, NodeType: task.TaskType, OperatorID: &from, Action: "transfer", Comment: comment, CreatedAt: now}); err != nil {
		return err
	}
	return errors.Join(e.eventBus.Publish(ctx, events.TaskCreatedEvent{InstanceID: inst.ID, TaskID: next.ID, AssigneeID: to, NodeID: task.NodeID, NodeName: task.NodeName, TaskType: task.TaskType})...)
}

func (e *FlowEngine) RollbackTask(ctx context.Context, taskID, user uuid.UUID, target, comment string) error {
	return e.operateTask(ctx, taskID, user, func(bound *FlowEngine, task *model.FlowTask, inst *model.FlowInstance) error {
		version, err := bound.defRepo.GetVersionByID(ctx, inst.DefinitionVersionID)
		if err != nil {
			return err
		}
		if version == nil {
			return response.NewAppError(response.CodeWorkflowVerNotFound, "流程版本不存在")
		}
		graph, err := ParseGraph(version.BpmnData)
		if err != nil {
			return err
		}
		targetNode := graph.GetNode(target)
		if target == task.NodeID || targetNode == nil || targetNode.Type != "approval" || !graph.precedes(target, task.NodeID) {
			return response.NewAppError(response.CodeBadRequest, "只能退回已处理的前置审批节点")
		}
		active := GetCurrentNodeIDs(inst.CurrentNodeIDs)
		if len(active) != 1 || active[0] != task.NodeID {
			return response.NewAppError(response.CodeBadRequest, "并行分支尚未汇合，不能退回其他分支")
		}
		history, err := bound.taskRepo.ListHistory(ctx, inst.ID)
		if err != nil {
			return err
		}
		visited := false
		for _, item := range history {
			if item.NodeID == target && item.Action == "approve" {
				visited = true
			}
		}
		if !visited {
			return response.NewAppError(response.CodeBadRequest, "目标节点尚未审批，不能退回")
		}
		if err := bound.cancelPendingTasks(ctx, inst.ID, user, "退回重审："+comment); err != nil {
			return err
		}
		now := time.Now()
		if err := bound.taskRepo.CreateHistory(ctx, nil, &model.FlowHistory{ID: uuid.New(), InstanceID: inst.ID, TaskID: &task.ID, NodeID: task.NodeID, NodeName: task.NodeName, NodeType: task.TaskType, OperatorID: &user, Action: "rollback", Comment: comment, FromNodeID: task.NodeID, ToNodeID: target, CreatedAt: now}); err != nil {
			return err
		}
		if err := bound.updateCurrentNodes(ctx, inst, []string{}); err != nil {
			return err
		}
		variables, err := bound.varRepo.GetMap(ctx, inst.ID)
		if err != nil {
			return err
		}
		return bound.executeFromNodes(ctx, inst, graph, []string{target}, variables)
	})
}

func (g *FlowGraph) precedes(source, target string) bool {
	queue, seen := []string{source}, map[string]bool{}
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		if id == target {
			return true
		}
		if seen[id] {
			continue
		}
		seen[id] = true
		for _, edge := range g.GetNextNodes(id, "") {
			queue = append(queue, edge.Target)
		}
	}
	return false
}
