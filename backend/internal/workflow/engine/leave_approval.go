package engine

import (
	"context"

	"github.com/google/uuid"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

// CompleteLeaveApproval completes the current leave-approval node inside an
// already-authorized business transaction. Reject always terminates leftover todos.
func (e *FlowEngine) CompleteLeaveApproval(ctx context.Context, instanceID, reviewer uuid.UUID, action TaskAction, comment string) error {
	if !e.businessTransaction {
		return response.NewError(response.CodeForbidden, "请假审批需要业务事务")
	}
	if action != ActionApprove && action != ActionReject {
		return response.NewError(response.CodeBadRequest, "请假审批只支持同意或拒绝")
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
	if inst == nil || inst.BusinessType != LeaveBusinessType || inst.Status != 0 {
		return response.NewError(response.CodeConflict, "关联请假流程不可审批")
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
	fresh, err := e.instRepo.GetByID(ctx, instanceID)
	if err != nil {
		return err
	}
	if fresh != nil && fresh.Status == 0 {
		return e.Terminate(ctx, instanceID, reviewer, "请假审批拒绝："+comment)
	}
	return nil
}
