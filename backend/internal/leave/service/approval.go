package service

import (
	"context"

	"github.com/Yogdunana/StarByte/backend/internal/leave/model"
	"github.com/Yogdunana/StarByte/backend/internal/leave/repo"
	"github.com/google/uuid"
)

func (s *leaveService) Approve(ctx context.Context, viewer Viewer, id uuid.UUID, remark string) error {
	if !viewer.CanApprove {
		return noAccess("无权审批请假申请")
	}
	return s.rows.WithTx(ctx, func(tx repo.Repository) error {
		app, err := tx.GetLeaveApplicationByID(ctx, id)
		if err != nil {
			return err
		}
		if app == nil {
			return notFound()
		}
		if app.ApplicantID == viewer.UserID {
			return noAccess("不能审批自己的请假申请")
		}
		if app.Status != model.ApprovalStatusPending {
			return invalidState("该请假申请已审批，不可重复审批")
		}
		return tx.UpdateApprovalStatus(ctx, id, viewer.UserID, model.ApprovalStatusApproved, remark, s.clock())
	})
}

func (s *leaveService) Reject(ctx context.Context, viewer Viewer, id uuid.UUID, remark string) error {
	if !viewer.CanApprove {
		return noAccess("无权审批请假申请")
	}
	return s.rows.WithTx(ctx, func(tx repo.Repository) error {
		app, err := tx.GetLeaveApplicationByID(ctx, id)
		if err != nil {
			return err
		}
		if app == nil {
			return notFound()
		}
		if app.ApplicantID == viewer.UserID {
			return noAccess("不能审批自己的请假申请")
		}
		if app.Status != model.ApprovalStatusPending {
			return invalidState("该请假申请已审批，不可重复审批")
		}
		if err := tx.UpdateApprovalStatus(ctx, id, viewer.UserID, model.ApprovalStatusRejected, remark, s.clock()); err != nil {
			return err
		}
		leaveType, err := tx.GetLeaveTypeByID(ctx, app.LeaveTypeID)
		if err != nil {
			return err
		}
		if leaveType == nil {
			return typeNotFound()
		}
		if leaveType.Deductible {
			year := app.StartTime.Year()
			if err := tx.DeductLeaveBalance(ctx, app.ApplicantID, year, app.LeaveTypeID, -app.DurationDays); err != nil {
				return err
			}
		}
		return nil
	})
}
