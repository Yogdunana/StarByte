package service

import (
	"context"
	"errors"

	"github.com/Yogdunana/StarByte/backend/internal/leave/model"
	"github.com/Yogdunana/StarByte/backend/internal/leave/repo"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/engine"
	"github.com/google/uuid"
)

func (s *leaveService) loadPending(ctx context.Context, tx repo.Repository, viewer Viewer, id uuid.UUID) (*model.ApplicationNamed, error) {
	app, err := tx.GetLeaveApplicationByIDForUpdate(ctx, id)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, notFound()
	}
	if app.ApplicantID == viewer.UserID {
		return nil, noAccess("不能审批自己的请假申请")
	}
	if !canAccessApplicant(viewer.Scope, app.ApplicantID, app.ApplicantDepartmentID, viewer.UserID) {
		return nil, noAccess("无权审批该请假申请")
	}
	if app.Status != model.ApprovalStatusPending {
		return nil, invalidState("该请假申请已审批，不可重复审批")
	}
	return app, nil
}

func mapApprovalErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, repo.ErrNotPending) {
		return invalidState("该请假申请已审批，不可重复审批")
	}
	return err
}

func (s *leaveService) Approve(ctx context.Context, viewer Viewer, id uuid.UUID, remark string) error {
	return s.decide(ctx, viewer, id, remark, true)
}

func (s *leaveService) Reject(ctx context.Context, viewer Viewer, id uuid.UUID, remark string) error {
	return s.decide(ctx, viewer, id, remark, false)
}

func (s *leaveService) decide(ctx context.Context, viewer Viewer, id uuid.UUID, remark string, approve bool) error {
	if !viewer.CanApprove {
		return noAccess("无权审批请假申请")
	}
	return s.withFlow(ctx, func(tx repo.Repository, flow LeaveFlow) error {
		app, err := s.loadPending(ctx, tx, viewer, id)
		if err != nil {
			return err
		}
		if app.WorkflowInstanceID != nil && s.engine != nil {
			return s.decideViaEngine(ctx, tx, flow, viewer, app, remark, approve)
		}
		return s.decideLegacy(ctx, tx, viewer, app, remark, approve)
	})
}

func (s *leaveService) decideViaEngine(ctx context.Context, tx repo.Repository, flow LeaveFlow, viewer Viewer, app *model.ApplicationNamed, remark string, approve bool) error {
	if flow == nil {
		return workflowUnavailable()
	}
	action := engine.ActionApprove
	if !approve {
		action = engine.ActionReject
	}
	if err := flow.CompleteLeaveApproval(ctx, *app.WorkflowInstanceID, viewer.UserID, action, remark); err != nil {
		return err
	}
	now := s.clock()
	app.ApproverID = &viewer.UserID
	app.ApproveRemark = remark
	app.UpdatedAt = now
	if !approve {
		app.Status = model.ApprovalStatusRejected
		app.ApprovedAt = &now
		app.WorkflowStage = ""
		if err := s.restoreBalance(ctx, tx, app); err != nil {
			return err
		}
		return tx.SaveApplication(ctx, &app.LeaveApplication)
	}
	nodeID, done, err := flow.RunningApprovalNode(ctx, *app.WorkflowInstanceID)
	if err != nil {
		return err
	}
	if done {
		app.Status = model.ApprovalStatusApproved
		app.ApprovedAt = &now
		app.WorkflowStage = ""
	} else {
		app.WorkflowStage = nodeID
	}
	return tx.SaveApplication(ctx, &app.LeaveApplication)
}

func (s *leaveService) decideLegacy(ctx context.Context, tx repo.Repository, viewer Viewer, app *model.ApplicationNamed, remark string, approve bool) error {
	status := model.ApprovalStatusApproved
	if !approve {
		status = model.ApprovalStatusRejected
		if err := s.restoreBalance(ctx, tx, app); err != nil {
			return err
		}
	}
	return mapApprovalErr(tx.UpdateApprovalStatus(ctx, app.ID, viewer.UserID, status, remark, s.clock(), app.BalanceDeducted))
}

func (s *leaveService) restoreBalance(ctx context.Context, tx repo.Repository, app *model.ApplicationNamed) error {
	if !app.BalanceDeducted {
		return nil
	}
	year := bizYear(app.StartTime)
	if err := tx.DeductLeaveBalance(ctx, app.ApplicantID, year, app.LeaveTypeID, -app.DurationDays); err != nil {
		return err
	}
	app.BalanceDeducted = false
	return nil
}
