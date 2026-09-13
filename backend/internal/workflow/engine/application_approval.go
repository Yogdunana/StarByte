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
	role, _, err := e.ApprovalPolicy(ctx, instanceID, nodeID)
	if err != nil {
		return err
	}
	var selected *model.FlowTask
	if role == "standing_committee" {
		selected, err = e.selectAssignedApprovalTask(ctx, instanceID, nodeID, reviewer)
	} else {
		selected, err = e.selectApprovalTask(ctx, instanceID, nodeID, reviewer)
	}
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

// TransferApplicationApproval reassigns the current membership approval task
// inside an already-authorized business transaction.
func (e *FlowEngine) TransferApplicationApproval(ctx context.Context, instanceID, from, to uuid.UUID, comment string) error {
	if !e.businessTransaction {
		return response.NewError(response.CodeForbidden, "入会转交需要业务事务")
	}
	if to == uuid.Nil || to == from {
		return response.NewError(response.CodeBadRequest, "请选择其他有效处理人")
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
		return response.NewError(response.CodeConflict, "关联入会流程不可转交")
	}
	nodeID, err := e.currentApprovalNodeID(ctx, inst)
	if err != nil {
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
	if !nodeAllowsTransfer(graph.GetNode(nodeID)) {
		return response.NewError(response.CodeForbidden, "当前环节不允许转交")
	}
	selected, err := e.selectExistingApprovalTask(ctx, instanceID, nodeID, from)
	if err != nil {
		return err
	}
	return e.transferPendingTask(ctx, selected, inst, from, to, comment)
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
	own, reference, err := e.findPendingApprovalTasks(ctx, instanceID, nodeID, reviewer)
	if err != nil {
		return nil, err
	}
	if own != nil {
		return own, nil
	}
	now := time.Now()
	selected := &model.FlowTask{
		ID: uuid.New(), InstanceID: instanceID, NodeID: nodeID, NodeName: reference.NodeName,
		TaskType: "approval", ActivationID: reference.ActivationID, AssigneeID: &reviewer,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := e.taskRepo.CreateTask(ctx, nil, selected); err != nil {
		return nil, err
	}
	return selected, nil
}

// selectExistingApprovalTask never mints a parallel vote. Transfer must hand
// off an already-open task so or-sign / countersign counts stay intact.
func (e *FlowEngine) selectExistingApprovalTask(ctx context.Context, instanceID uuid.UUID, nodeID string, reviewer uuid.UUID) (*model.FlowTask, error) {
	own, reference, err := e.findPendingApprovalTasks(ctx, instanceID, nodeID, reviewer)
	if err != nil {
		return nil, err
	}
	if own != nil {
		return own, nil
	}
	return reference, nil
}

func (e *FlowEngine) findPendingApprovalTasks(ctx context.Context, instanceID uuid.UUID, nodeID string, reviewer uuid.UUID) (own, reference *model.FlowTask, err error) {
	tasks, err := e.taskRepo.ListTasksByInstance(ctx, instanceID)
	if err != nil {
		return nil, nil, err
	}
	for i := range tasks {
		task := &tasks[i]
		if task.NodeID != nodeID || task.Status != 0 {
			continue
		}
		if reference == nil {
			reference = task
		}
		if task.AssigneeID != nil && *task.AssigneeID == reviewer {
			own = task
		}
	}
	if reference == nil {
		return nil, nil, response.NewError(response.CodeConflict, "当前审批环节没有待办")
	}
	return own, reference, nil
}

// ApprovalRoleCode returns the published roleCode for an approval node,
// falling back to the default officer/minister/president node IDs.
func (e *FlowEngine) ApprovalRoleCode(ctx context.Context, instanceID uuid.UUID, nodeID string) (string, error) {
	role, _, err := e.ApprovalPolicy(ctx, instanceID, nodeID)
	return role, err
}

// ApprovalPolicy returns the published roleCode and departmentScope for a node.
func (e *FlowEngine) ApprovalPolicy(ctx context.Context, instanceID uuid.UUID, nodeID string) (string, bool, error) {
	inst, err := e.instRepo.GetByID(ctx, instanceID)
	if err != nil {
		return "", false, err
	}
	if inst == nil {
		return "", false, response.NewError(response.CodeWorkflowInstNotFound, "流程实例不存在")
	}
	version, err := e.defRepo.GetVersionByID(ctx, inst.DefinitionVersionID)
	if err != nil {
		return "", false, err
	}
	if version == nil {
		return "", false, response.NewError(response.CodeWorkflowVerNotFound, "审批流程版本不存在")
	}
	graph, err := ParseGraph(version.BpmnData)
	if err != nil {
		return "", false, err
	}
	node := graph.GetNode(nodeID)
	return ResolveApprovalRole(nodeID, approvalRoleCode(node)), approvalDepartmentScope(node), nil
}

// IsLastApplicationApproval reports whether nodeID is the last human
// approval on the instance graph (no further approval is reachable).
func (e *FlowEngine) IsLastApplicationApproval(ctx context.Context, instanceID uuid.UUID, nodeID string) (bool, error) {
	inst, err := e.instRepo.GetByID(ctx, instanceID)
	if err != nil {
		return false, err
	}
	if inst == nil {
		return false, response.NewError(response.CodeWorkflowInstNotFound, "流程实例不存在")
	}
	version, err := e.defRepo.GetVersionByID(ctx, inst.DefinitionVersionID)
	if err != nil {
		return false, err
	}
	if version == nil {
		return false, response.NewError(response.CodeWorkflowVerNotFound, "审批流程版本不存在")
	}
	graph, err := ParseGraph(version.BpmnData)
	if err != nil {
		return false, err
	}
	return lastApplicationApproval(graph, GetCurrentNodeIDs(inst.CurrentNodeIDs), nodeID), nil
}

// ApplicationApprovalRoles lists roleCode values on the instance graph.
func (e *FlowEngine) ApplicationApprovalRoles(ctx context.Context, instanceID uuid.UUID) ([]string, error) {
	inst, err := e.instRepo.GetByID(ctx, instanceID)
	if err != nil {
		return nil, err
	}
	if inst == nil {
		return nil, response.NewError(response.CodeWorkflowInstNotFound, "流程实例不存在")
	}
	version, err := e.defRepo.GetVersionByID(ctx, inst.DefinitionVersionID)
	if err != nil {
		return nil, err
	}
	if version == nil {
		return nil, response.NewError(response.CodeWorkflowVerNotFound, "审批流程版本不存在")
	}
	graph, err := ParseGraph(version.BpmnData)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	out := make([]string, 0, 4)
	for _, node := range graph.Nodes {
		if node == nil || node.Type != "approval" {
			continue
		}
		role := ResolveApprovalRole(node.ID, approvalRoleCode(node))
		if role == "" || seen[role] {
			continue
		}
		seen[role] = true
		out = append(out, role)
	}
	return out, nil
}

// ApplicationRolePolicy is one approval node's published role and department fence.
type ApplicationRolePolicy struct {
	Role            string
	DepartmentScope bool
}

// ApplicationApprovalPolicies lists roleCode + departmentScope for each approval node.
func (e *FlowEngine) ApplicationApprovalPolicies(ctx context.Context, instanceID uuid.UUID) ([]ApplicationRolePolicy, error) {
	inst, err := e.instRepo.GetByID(ctx, instanceID)
	if err != nil {
		return nil, err
	}
	if inst == nil {
		return nil, response.NewError(response.CodeWorkflowInstNotFound, "流程实例不存在")
	}
	version, err := e.defRepo.GetVersionByID(ctx, inst.DefinitionVersionID)
	if err != nil {
		return nil, err
	}
	if version == nil {
		return nil, response.NewError(response.CodeWorkflowVerNotFound, "审批流程版本不存在")
	}
	graph, err := ParseGraph(version.BpmnData)
	if err != nil {
		return nil, err
	}
	out := make([]ApplicationRolePolicy, 0, 4)
	for _, node := range graph.Nodes {
		if node == nil || node.Type != "approval" {
			continue
		}
		role := ResolveApprovalRole(node.ID, approvalRoleCode(node))
		if role == "" {
			continue
		}
		out = append(out, ApplicationRolePolicy{Role: role, DepartmentScope: approvalDepartmentScope(node)})
	}
	return out, nil
}

func ResolveApprovalRole(nodeID, roleCode string) string {
	if code := strings.TrimSpace(roleCode); code != "" {
		return code
	}
	switch nodeID {
	case "officer", "minister", "president":
		return nodeID
	default:
		return ""
	}
}
