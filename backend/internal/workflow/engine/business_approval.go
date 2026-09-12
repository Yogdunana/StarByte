package engine

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/model"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

// CompleteBusinessApproval is internal to an already-authorized business
// transaction. The business service must persist the real signature first.
// Public workflow handlers never receive a transaction-bound engine.
func (e *FlowEngine) CompleteBusinessApproval(ctx context.Context, instanceID uuid.UUID, stage, role string, signer uuid.UUID, signatureID uuid.UUID) error {
	return e.completeBusinessSignature(ctx, instanceID, "member_application", "admissionStage", "admissionRole", stage, role, signer, signatureID, "已正式签字")
}
func (e *FlowEngine) CompleteTaskTransferApproval(ctx context.Context, instanceID uuid.UUID, role string, signer, signatureID uuid.UUID, waived bool) error {
	comment := "转办正式签字"
	if waived {
		comment = "上级签字，明确豁免下级签字要求"
	}
	return e.completeBusinessSignature(ctx, instanceID, TaskTransferBusinessType, "transferStage", "transferRole", "handover", role, signer, signatureID, comment)
}
func (e *FlowEngine) completeBusinessSignature(ctx context.Context, instanceID uuid.UUID, kind, stageKey, roleKey, stage, role string, signer, signatureID uuid.UUID, comment string) error {
	if !e.businessTransaction || e.db == nil {
		return response.NewError(response.CodeForbidden, "正式签字需要业务事务")
	}
	if err := repo.NewRuntimeRepo(e.db).LockInstance(ctx, instanceID); err != nil {
		return err
	}
	inst, err := e.instRepo.GetByID(ctx, instanceID)
	if err != nil {
		return err
	}
	if inst == nil || inst.BusinessType != kind || inst.Status != 0 {
		return response.NewError(response.CodeConflict, "关联入会流程不可审批")
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
	nodeID := ""
	for _, id := range GetCurrentNodeIDs(inst.CurrentNodeIDs) {
		node := graph.GetNode(id)
		if node != nil && node.Config[stageKey] == stage && node.Config[roleKey] == role {
			nodeID = id
			break
		}
	}
	if nodeID == "" {
		return response.NewError(response.CodeConflict, "签字环节与流程待办不一致，请核验流程配置")
	}
	tasks, err := e.taskRepo.ListTasksByInstance(ctx, instanceID)
	if err != nil {
		return err
	}
	var selected, reference *model.FlowTask
	for i := range tasks {
		task := &tasks[i]
		if task.NodeID != nodeID || task.Status != 0 {
			continue
		}
		reference = task
		if task.AssigneeID != nil && *task.AssigneeID == signer {
			selected = task
		}
	}
	if reference == nil {
		return response.NewError(response.CodeConflict, "当前签字环节没有待办")
	}
	if selected == nil {
		// A legitimate role holder may change after activation. Preserve old tasks,
		// and record the current authorized human signer without impersonation.
		now := time.Now()
		selected = &model.FlowTask{ID: uuid.New(), InstanceID: instanceID, NodeID: nodeID, NodeName: reference.NodeName, TaskType: "approval", ActivationID: reference.ActivationID, AssigneeID: &signer, CreatedAt: now, UpdatedAt: now}
		if err := e.taskRepo.CreateTask(ctx, nil, selected); err != nil {
			return err
		}
	}
	// Store only the signature reference, not confidential interview notes.
	return e.CompleteTask(ctx, selected.ID, signer, ActionApprove, comment, map[string]interface{}{"last_signature_id": signatureID.String()})
}

// BusinessStage returns the active business stage after engine advancement.
// All active role branches must belong to the same stage.
func (e *FlowEngine) BusinessStage(ctx context.Context, id uuid.UUID) (string, bool, error) {
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
	version, err := e.defRepo.GetVersionByID(ctx, inst.DefinitionVersionID)
	if err != nil {
		return "", false, err
	}
	if version == nil {
		return "", false, response.NewError(response.CodeWorkflowVerNotFound, "流程版本不存在")
	}
	graph, err := ParseGraph(version.BpmnData)
	if err != nil {
		return "", false, err
	}
	stage := ""
	for _, id := range GetCurrentNodeIDs(inst.CurrentNodeIDs) {
		node := graph.GetNode(id)
		if node == nil {
			return "", false, response.NewError(response.CodeConflict, "流程节点缺失")
		}
		stageKey := "admissionStage"
		if inst.BusinessType == TaskBusinessType {
			stageKey = "taskStage"
		}
		if inst.BusinessType == TaskTransferBusinessType {
			stageKey = "transferStage"
		}
		if inst.BusinessType == LeaveBusinessType {
			stageKey = "leaveStage"
		}
		next, _ := node.Config[stageKey].(string)
		// A parallel join can wait alongside the other signature branch.
		if node.Type == "parallel_gateway" {
			continue
		}
		if next == "" || (stage != "" && stage != next) {
			return "", false, response.NewError(response.CodeConflict, "流程签字环节不一致")
		}
		stage = next
	}
	if stage == "" {
		return "", false, response.NewError(response.CodeConflict, "流程没有可处理的签字环节")
	}
	return stage, false, nil
}
