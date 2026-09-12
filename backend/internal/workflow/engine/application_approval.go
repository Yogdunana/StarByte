package engine

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/model"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

// CompleteApplicationApproval completes the current role-based membership task
// inside an already-authorized business transaction.
func (e *FlowEngine) CompleteApplicationApproval(ctx context.Context, instanceID, reviewer uuid.UUID, action TaskAction, comment string) error {
	if !e.businessTransaction {
		return response.NewError(response.CodeForbidden, "入会审批需要业务事务")
	}
	if action != ActionApprove && action != ActionReject {
		return response.NewError(response.CodeBadRequest, "入会审批只支持同意或拒绝")
	}
	if e.db != nil {
		if err := repo.NewRuntimeRepo(e.db).LockInstance(ctx, instanceID); err != nil {
			return err
		}
	}
	inst, err := e.instRepo.GetByID(ctx, instanceID)
	if err != nil {
		return err
	}
	if inst == nil || inst.BusinessType != "member_application" || inst.Status != 0 {
		return response.NewError(response.CodeConflict, "关联入会流程不可审批")
	}
	nodeID, err := e.currentApprovalNodeID(ctx, inst)
	if err != nil {
		return err
	}
	selected, err := e.selectApprovalTask(ctx, instanceID, nodeID, reviewer)
	if err != nil {
		return err
	}
	if err := e.CompleteTask(ctx, selected.ID, reviewer, action, comment, nil); err != nil {
		return err
	}
	if action != ActionReject {
		return nil
	}
	// approvalType=any 在仍有 pending 时不会因单票拒绝结束实例；业务拒绝必须强制终止。
	fresh, err := e.instRepo.GetByID(ctx, instanceID)
	if err != nil {
		return err
	}
	if fresh != nil && fresh.Status == 0 {
		return e.Terminate(ctx, instanceID, reviewer, "入会审批拒绝："+comment)
	}
	return nil
}

// SkipApplicationApproval 跳过当前审批节点并进入下一环节。
// 会员申请无意向部门时没有对口部长，部长节点不能指派，因此直达社长。
func (e *FlowEngine) SkipApplicationApproval(ctx context.Context, instanceID uuid.UUID, nodeID string, operator uuid.UUID, reason string) error {
	if !e.businessTransaction {
		return response.NewError(response.CodeForbidden, "入会审批需要业务事务")
	}
	if e.db != nil {
		if err := repo.NewRuntimeRepo(e.db).LockInstance(ctx, instanceID); err != nil {
			return err
		}
	}
	inst, err := e.instRepo.GetByID(ctx, instanceID)
	if err != nil {
		return err
	}
	if inst == nil || inst.BusinessType != "member_application" || inst.Status != 0 {
		return response.NewError(response.CodeConflict, "关联入会流程不可跳过")
	}
	current, err := e.currentApprovalNodeID(ctx, inst)
	if err != nil {
		return err
	}
	if current != nodeID {
		return response.NewError(response.CodeConflict, "当前环节不是可跳过的审批节点")
	}
	if err := e.cancelNodePendingTasks(ctx, inst.ID, nodeID, operator, reason); err != nil {
		return err
	}
	now := time.Now()
	hist := &model.FlowHistory{
		ID: uuid.New(), InstanceID: inst.ID, NodeID: nodeID, NodeName: nodeID,
		NodeType: "approval", Action: "skip", Comment: reason, CreatedAt: now,
	}
	if operator != uuid.Nil {
		hist.OperatorID = &operator
	}
	if err := e.taskRepo.CreateHistory(ctx, nil, hist); err != nil {
		return err
	}
	version, err := e.defRepo.GetVersionByID(ctx, inst.DefinitionVersionID)
	if err != nil {
		return err
	}
	if version == nil {
		return response.NewError(response.CodeWorkflowVerNotFound, "审批流程版本不存在")
	}
	graph, err := ParseGraph(version.BpmnData)
	if err != nil {
		return err
	}
	node := graph.GetNode(nodeID)
	if node == nil {
		return response.NewError(response.CodeWorkflowNodeNotFound, "流程节点不存在: "+nodeID)
	}
	vars, err := e.varRepo.GetMap(ctx, inst.ID)
	if err != nil {
		return err
	}
	next, err := e.executeNode(ctx, inst, graph, node, vars)
	if err != nil {
		return err
	}
	currentIDs := []string{}
	for _, id := range GetCurrentNodeIDs(inst.CurrentNodeIDs) {
		if id != nodeID {
			currentIDs = append(currentIDs, id)
		}
	}
	if err := e.updateCurrentNodes(ctx, inst, currentIDs); err != nil {
		return err
	}
	if len(next) == 0 {
		return nil
	}
	return e.executeFromNodes(ctx, inst, graph, next, vars, node.ID)
}

func (e *FlowEngine) cancelNodePendingTasks(ctx context.Context, instanceID uuid.UUID, nodeID string, operator uuid.UUID, reason string) error {
	tasks, err := e.taskRepo.ListTasksByInstance(ctx, instanceID)
	if err != nil {
		return err
	}
	now := time.Now()
	for _, task := range tasks {
		if task.Status != 0 || task.NodeID != nodeID {
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
		if operator != uuid.Nil {
			hist.OperatorID = &operator
		}
		if err := e.taskRepo.CreateHistory(ctx, nil, hist); err != nil {
			return err
		}
	}
	return nil
}

// RunningApprovalNode returns the active approval node, or completed=true when the instance ended.
func (e *FlowEngine) RunningApprovalNode(ctx context.Context, id uuid.UUID) (string, bool, error) {
	inst, err := e.instRepo.GetByID(ctx, id)
	if err != nil {
		return "", false, err
	}
	if inst == nil {
		return "", false, response.NewError(response.CodeWorkflowInstNotFound, "流程实例不存在")
	}
	if inst.Status == 1 {
		return "", true, nil
	}
	if inst.Status != 0 {
		return "", false, response.NewError(response.CodeConflict, "关联流程已中止")
	}
	nodeID, err := e.currentApprovalNodeID(ctx, inst)
	return nodeID, false, err
}

func (e *FlowEngine) currentApprovalNodeID(ctx context.Context, inst *model.FlowInstance) (string, error) {
	version, err := e.defRepo.GetVersionByID(ctx, inst.DefinitionVersionID)
	if err != nil {
		return "", err
	}
	if version == nil {
		return "", response.NewError(response.CodeWorkflowVerNotFound, "审批流程版本不存在")
	}
	graph, err := ParseGraph(version.BpmnData)
	if err != nil {
		return "", err
	}
	for _, id := range GetCurrentNodeIDs(inst.CurrentNodeIDs) {
		node := graph.GetNode(id)
		if node != nil && node.Type == "approval" {
			return id, nil
		}
	}
	return "", response.NewError(response.CodeConflict, "当前没有可处理的审批环节")
}

func (e *FlowEngine) selectApprovalTask(ctx context.Context, instanceID uuid.UUID, nodeID string, reviewer uuid.UUID) (*model.FlowTask, error) {
	tasks, err := e.taskRepo.ListTasksByInstance(ctx, instanceID)
	if err != nil {
		return nil, err
	}
	var selected, reference *model.FlowTask
	for i := range tasks {
		task := &tasks[i]
		if task.NodeID != nodeID || task.Status != 0 {
			continue
		}
		reference = task
		if task.AssigneeID != nil && *task.AssigneeID == reviewer {
			selected = task
		}
	}
	if reference == nil {
		return nil, response.NewError(response.CodeConflict, "当前审批环节没有待办")
	}
	if selected != nil {
		return selected, nil
	}
	now := time.Now()
	selected = &model.FlowTask{
		ID: uuid.New(), InstanceID: instanceID, NodeID: nodeID, NodeName: reference.NodeName,
		TaskType: "approval", ActivationID: reference.ActivationID, AssigneeID: &reviewer,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := e.taskRepo.CreateTask(ctx, nil, selected); err != nil {
		return nil, err
	}
	return selected, nil
}
