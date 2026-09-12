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
	return e.CompleteTask(ctx, selected.ID, reviewer, action, comment, nil)
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
	return "", response.NewError(response.CodeConflict, "当前没有可处理的入会审批环节")
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
