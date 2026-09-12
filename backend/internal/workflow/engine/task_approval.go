package engine

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/model"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

// CompleteTaskApproval completes the current collaboration-task node inside an
// already-authorized business transaction (same pattern as CompleteApplicationApproval).
// Reject / fail always terminates the instance so leftover pending todos cannot revive it.
func (e *FlowEngine) CompleteTaskApproval(ctx context.Context, instanceID, actor uuid.UUID, action TaskAction, comment string) error {
	if !e.businessTransaction {
		return response.NewError(response.CodeForbidden, "任务审批需要业务事务")
	}
	if action != ActionApprove && action != ActionReject {
		return response.NewError(response.CodeBadRequest, "任务审批只支持同意或拒绝")
	}
	if action == ActionReject && strings.TrimSpace(comment) == "" {
		return response.NewError(response.CodeBadRequest, "拒绝须填写原因")
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
	if inst == nil || inst.BusinessType != TaskBusinessType || inst.Status != 0 {
		return response.NewError(response.CodeConflict, "任务流程当前不可审批")
	}
	nodeID, err := e.currentApprovalNodeID(ctx, inst)
	if err != nil {
		return err
	}
	selected, err := e.selectApprovalTask(ctx, instanceID, nodeID, actor)
	if err != nil {
		return err
	}
	if err := e.CompleteTask(ctx, selected.ID, actor, action, comment, nil); err != nil {
		return err
	}
	if action != ActionReject {
		return nil
	}
	fresh, err := e.instRepo.GetByID(ctx, instanceID)
	if err != nil {
		return err
	}
	if fresh != nil && fresh.Status == 0 {
		return e.Terminate(ctx, instanceID, actor, "任务审批拒绝："+comment)
	}
	return nil
}

// ClaimTaskAssignment advances past the assignment node after the business task
// already has an assignee. The publisher's assignment todo is cancelled so it
// cannot later complete or revive the stage.
func (e *FlowEngine) ClaimTaskAssignment(ctx context.Context, instanceID, actor uuid.UUID, comment string) error {
	if !e.businessTransaction {
		return response.NewError(response.CodeForbidden, "任务认领需要业务事务")
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
	if inst == nil || inst.BusinessType != TaskBusinessType || inst.Status != 0 {
		return response.NewError(response.CodeConflict, "任务流程当前不可认领")
	}
	nodeID, err := e.currentApprovalNodeID(ctx, inst)
	if err != nil {
		return err
	}
	stage, err := e.nodeTaskStage(ctx, inst, nodeID)
	if err != nil {
		return err
	}
	if stage != "assignment" {
		return response.NewError(response.CodeConflict, "当前不是认领或分配环节")
	}
	reason := strings.TrimSpace(comment)
	if reason == "" {
		reason = "任务已被认领"
	}
	if err := e.cancelNodePendingTasks(ctx, inst.ID, nodeID, actor, reason); err != nil {
		return err
	}
	now := time.Now()
	hist := &model.FlowHistory{
		ID: uuid.New(), InstanceID: inst.ID, NodeID: nodeID, NodeName: nodeID,
		NodeType: "approval", Action: "claim", Comment: reason, CreatedAt: now,
	}
	if actor != uuid.Nil {
		hist.OperatorID = &actor
	}
	if err := e.taskRepo.CreateHistory(ctx, nil, hist); err != nil {
		return err
	}
	return e.advancePastApprovalNode(ctx, inst, nodeID, actor)
}

func (e *FlowEngine) nodeTaskStage(ctx context.Context, inst *model.FlowInstance, nodeID string) (string, error) {
	version, err := e.defRepo.GetVersionByID(ctx, inst.DefinitionVersionID)
	if err != nil {
		return "", err
	}
	if version == nil {
		return "", response.NewError(response.CodeWorkflowVerNotFound, "任务流程版本不存在")
	}
	graph, err := ParseGraph(version.BpmnData)
	if err != nil {
		return "", err
	}
	node := graph.GetNode(nodeID)
	if node == nil {
		return "", response.NewError(response.CodeWorkflowNodeNotFound, "流程节点不存在: "+nodeID)
	}
	stage, _ := node.Config["taskStage"].(string)
	return stage, nil
}

func (e *FlowEngine) advancePastApprovalNode(ctx context.Context, inst *model.FlowInstance, nodeID string, _ uuid.UUID) error {
	version, err := e.defRepo.GetVersionByID(ctx, inst.DefinitionVersionID)
	if err != nil {
		return err
	}
	if version == nil {
		return response.NewError(response.CodeWorkflowVerNotFound, "任务流程版本不存在")
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

// InstanceTerminated reports whether a collaboration instance was force-stopped.
func (e *FlowEngine) InstanceTerminated(ctx context.Context, id uuid.UUID) (bool, error) {
	inst, err := e.instRepo.GetByID(ctx, id)
	if err != nil {
		return false, err
	}
	if inst == nil {
		return false, response.NewError(response.CodeWorkflowInstNotFound, "流程实例不存在")
	}
	return inst.Status == 2, nil
}
