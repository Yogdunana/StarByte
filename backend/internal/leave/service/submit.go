package service

import (
	"context"

	"github.com/Yogdunana/StarByte/backend/internal/leave/dto"
	"github.com/Yogdunana/StarByte/backend/internal/leave/model"
	"github.com/Yogdunana/StarByte/backend/internal/leave/repo"
	"github.com/google/uuid"
)

func (s *leaveService) Submit(ctx context.Context, viewer Viewer, req *dto.SubmitLeaveRequest) (*dto.LeaveApplicationResponse, error) {
	typeID, err := uuid.Parse(req.LeaveTypeID)
	if err != nil {
		return nil, typeNotFound()
	}
	var createdID uuid.UUID
	err = s.rows.WithTx(ctx, func(tx repo.Repository) error {
		id, txErr := s.submitInTx(ctx, tx, viewer.UserID, typeID, req)
		createdID = id
		return txErr
	})
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, viewer, createdID)
}

func (s *leaveService) submitInTx(ctx context.Context, tx repo.Repository, applicantID, typeID uuid.UUID, req *dto.SubmitLeaveRequest) (uuid.UUID, error) {
	if err := tx.LockApplicant(ctx, applicantID); err != nil {
		return uuid.Nil, err
	}
	leaveType, err := tx.GetLeaveTypeByID(ctx, typeID)
	if err != nil {
		return uuid.Nil, err
	}
	if leaveType == nil {
		return uuid.Nil, typeNotFound()
	}
	if !req.StartTime.Before(req.EndTime) {
		return uuid.Nil, invalidTime("开始时间必须早于结束时间")
	}
	if calendarDate(req.StartTime).Before(calendarDate(s.clock())) {
		return uuid.Nil, invalidTime("开始时间不能早于今天")
	}

	durationDays := CalculateDurationDays(req.StartTime, req.EndTime)
	if durationDays <= 0 {
		return uuid.Nil, invalidTime("请假时长必须大于0")
	}

	overlapRows, err := tx.GetLeaveApplicationsByUserAndTimeRange(ctx, applicantID, req.StartTime, req.EndTime)
	if err != nil {
		return uuid.Nil, err
	}
	if len(overlapRows) > 0 {
		return uuid.Nil, overlap()
	}

	year := bizYear(req.StartTime)
	if leaveType.Deductible {
		if err := s.ensureAndDeduct(ctx, tx, applicantID, year, leaveType, durationDays); err != nil {
			return uuid.Nil, err
		}
	}

	app := &model.LeaveApplication{
		ID:           uuid.New(),
		ApplicantID:  applicantID,
		LeaveTypeID:  typeID,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
		DurationDays: durationDays,
		Reason:       req.Reason,
		Status:       model.ApprovalStatusPending,
		CreatedAt:    s.clock(),
		UpdatedAt:    s.clock(),
	}
	if err := tx.CreateLeaveApplication(ctx, app); err != nil {
		return uuid.Nil, err
	}
	return app.ID, nil
}

func (s *leaveService) ensureAndDeduct(ctx context.Context, tx repo.Repository, userID uuid.UUID, year int, leaveType *model.LeaveType, days float64) error {
	balance, err := tx.GetLeaveBalanceForUpdate(ctx, userID, year, leaveType.ID)
	if err != nil {
		return err
	}
	if balance == nil {
		if err := s.createBalance(ctx, tx, userID, year, leaveType); err != nil {
			return err
		}
		balance, err = tx.GetLeaveBalanceForUpdate(ctx, userID, year, leaveType.ID)
		if err != nil {
			return err
		}
		if balance == nil {
			return balanceMissing()
		}
	}
	if balance.RemainingDays < days {
		return insufficient()
	}
	return tx.DeductLeaveBalance(ctx, userID, year, leaveType.ID, days)
}

func (s *leaveService) createBalance(ctx context.Context, tx repo.Repository, userID uuid.UUID, year int, leaveType *model.LeaveType) error {
	now := s.clock()
	return tx.CreateLeaveBalance(ctx, &model.LeaveBalance{
		ID:            uuid.New(),
		UserID:        userID,
		Year:          year,
		LeaveTypeID:   leaveType.ID,
		TotalDays:     leaveType.DefaultDays,
		UsedDays:      0,
		RemainingDays: leaveType.DefaultDays,
		CreatedAt:     now,
		UpdatedAt:     now,
	})
}
