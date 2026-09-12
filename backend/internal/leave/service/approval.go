package service

import (
	"context"
	"errors"

	"github.com/Yogdunana/StarByte/backend/internal/leave/model"
	"github.com/Yogdunana/StarByte/backend/internal/leave/repo"
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
	if !viewer.CanApprove {
		return noAccess("无权审批请假申请")
	}
	return s.rows.WithTx(ctx, func(tx repo.Repository) error {
		if _, err := s.loadPending(ctx, tx, viewer, id); err != nil {
			return err
		}
		return mapApprovalErr(tx.UpdateApprovalStatus(ctx, id, viewer.UserID, model.ApprovalStatusApproved, remark, s.clock()))
	})
}

func (s *leaveService) Reject(ctx context.Context, viewer Viewer, id uuid.UUID, remark string) error {
	if !viewer.CanApprove {
		return noAccess("无权审批请假申请")
	}
	return s.rows.WithTx(ctx, func(tx repo.Repository) error {
		app, err := s.loadPending(ctx, tx, viewer, id)
		if err != nil {
			return err
		}
		if err := mapApprovalErr(tx.UpdateApprovalStatus(ctx, id, viewer.UserID, model.ApprovalStatusRejected, remark, s.clock())); err != nil {
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
			year := bizYear(app.StartTime)
			if err := tx.DeductLeaveBalance(ctx, app.ApplicantID, year, app.LeaveTypeID, -app.DurationDays); err != nil {
				return err
			}
		}
		return nil
	})
}
