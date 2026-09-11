package service

import (
	"context"
	"fmt"

	"github.com/Yogdunana/StarByte/backend/internal/activity/dto"
	"github.com/Yogdunana/StarByte/backend/internal/activity/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
)

func (s *activityService) Register(ctx context.Context, activityID, userID uuid.UUID) (*dto.RegistrationResponse, error) {
	var out *dto.RegistrationResponse
	var n *pendingNotify
	err := s.withTx(ctx, func(tx *activityService) error {
		resp, pn, err := tx.registerInTx(ctx, activityID, userID)
		out = resp
		n = pn
		return err
	})
	if err == nil {
		s.flushNotifies(ctx, n)
	}
	return out, err
}

func (s *activityService) registerInTx(ctx context.Context, activityID, userID uuid.UUID) (*dto.RegistrationResponse, *pendingNotify, error) {
	a, err := s.activities.GetByIDForUpdate(ctx, activityID)
	if err != nil {
		return nil, nil, fmt.Errorf("get activity: %w", err)
	}
	if a == nil {
		return nil, nil, response.NewError(response.CodeActivityNotFound, "活动不存在")
	}
	if a.Status != model.ActivityOpen {
		return nil, nil, response.NewError(response.CodeActivityInvalidState, "活动不在报名中")
	}

	existing, err := s.regs.GetByActivityAndUser(ctx, activityID, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("check registration: %w", err)
	}
	if existing != nil && existing.Status != model.RegCancelled {
		return nil, nil, response.NewError(response.CodeRegistrationExists, "你已报名该活动")
	}

	approvedCount, err := s.regs.CountByActivityAndStatus(ctx, activityID, model.RegApproved)
	if err != nil {
		return nil, nil, fmt.Errorf("count approved: %w", err)
	}

	status := model.RegApproved
	if a.MaxParticipants > 0 && approvedCount >= int64(a.MaxParticipants) {
		status = model.RegWaitlist
	}

	now := s.clock()
	var reg *model.ActivityRegistration
	if existing != nil {
		existing.Status = status
		existing.CheckinStatus = model.CheckinPending
		existing.CheckedInAt = nil
		existing.CheckinMethod = nil
		existing.GPSLatitude = nil
		existing.GPSLongitude = nil
		existing.UpdatedAt = now
		if err := s.regs.Update(ctx, existing); err != nil {
			return nil, nil, fmt.Errorf("update registration: %w", err)
		}
		reg = existing
	} else {
		reg = &model.ActivityRegistration{
			ID:            uuid.New(),
			ActivityID:    activityID,
			UserID:        userID,
			Status:        status,
			CheckinStatus: model.CheckinPending,
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		if err := s.regs.Create(ctx, reg); err != nil {
			return nil, nil, fmt.Errorf("create registration: %w", err)
		}
	}

	tpl := tplActivityRegistered
	if status == model.RegWaitlist {
		tpl = tplActivityWaitlist
	}
	pn := &pendingNotify{users: []uuid.UUID{userID}, template: tpl, activity: a}

	named, err := s.regs.ListByActivity(ctx, activityID)
	if err == nil {
		for i := range named {
			if named[i].ID == reg.ID {
				resp := toRegistrationResponse(&named[i])
				return &resp, pn, nil
			}
		}
	}
	return &dto.RegistrationResponse{
		ID:            reg.ID.String(),
		ActivityID:    reg.ActivityID.String(),
		Status:        reg.Status,
		CheckinStatus: reg.CheckinStatus,
		CreatedAt:     formatTime(reg.CreatedAt),
	}, pn, nil
}

func (s *activityService) GetMyRegistration(ctx context.Context, activityID, userID uuid.UUID) (*dto.RegistrationResponse, error) {
	a, err := s.activities.GetByID(ctx, activityID)
	if err != nil {
		return nil, fmt.Errorf("get activity: %w", err)
	}
	if a == nil {
		return nil, response.NewError(response.CodeActivityNotFound, "活动不存在")
	}
	reg, err := s.regs.GetByActivityAndUser(ctx, activityID, userID)
	if err != nil {
		return nil, fmt.Errorf("get registration: %w", err)
	}
	if reg == nil {
		return nil, nil
	}
	named, err := s.regs.ListByActivity(ctx, activityID)
	if err == nil {
		for i := range named {
			if named[i].ID == reg.ID {
				resp := toRegistrationResponse(&named[i])
				return &resp, nil
			}
		}
	}
	return &dto.RegistrationResponse{
		ID:            reg.ID.String(),
		ActivityID:    reg.ActivityID.String(),
		User:          dto.Person{ID: userID.String()},
		Status:        reg.Status,
		CheckinStatus: reg.CheckinStatus,
		CreatedAt:     formatTime(reg.CreatedAt),
	}, nil
}

func (s *activityService) CancelRegistration(ctx context.Context, activityID, userID uuid.UUID) error {
	var n *pendingNotify
	err := s.withTx(ctx, func(tx *activityService) error {
		pn, err := tx.cancelInTx(ctx, activityID, userID)
		n = pn
		return err
	})
	if err == nil {
		s.flushNotifies(ctx, n)
	}
	return err
}

func (s *activityService) cancelInTx(ctx context.Context, activityID, userID uuid.UUID) (*pendingNotify, error) {
	a, err := s.activities.GetByIDForUpdate(ctx, activityID)
	if err != nil {
		return nil, fmt.Errorf("get activity: %w", err)
	}
	reg, err := s.regs.GetByActivityAndUserForUpdate(ctx, activityID, userID)
	if err != nil {
		return nil, fmt.Errorf("get registration: %w", err)
	}
	if reg == nil || reg.Status == model.RegCancelled {
		return nil, response.NewError(response.CodeRegistrationNotFound, "未找到报名记录")
	}
	wasApproved := reg.Status == model.RegApproved
	reg.Status = model.RegCancelled
	reg.UpdatedAt = s.clock()
	if err := s.regs.Update(ctx, reg); err != nil {
		return nil, fmt.Errorf("cancel registration: %w", err)
	}
	if !wasApproved {
		return nil, nil
	}
	return s.promoteWaitlistIfSeat(ctx, a)
}

// promoteWaitlistIfSeat 仅在活动仍开放且实际空出名额时递补最早候补。
func (s *activityService) promoteWaitlistIfSeat(ctx context.Context, a *model.Activity) (*pendingNotify, error) {
	if a == nil || a.MaxParticipants <= 0 {
		return nil, nil
	}
	if a.Status != model.ActivityOpen && a.Status != model.ActivityOngoing {
		return nil, nil
	}
	approvedCount, err := s.regs.CountByActivityAndStatus(ctx, a.ID, model.RegApproved)
	if err != nil {
		return nil, fmt.Errorf("count approved: %w", err)
	}
	if approvedCount >= int64(a.MaxParticipants) {
		return nil, nil
	}
	waitlist, err := s.regs.ListWaitlist(ctx, a.ID)
	if err != nil || len(waitlist) == 0 {
		return nil, err
	}
	next := waitlist[0]
	next.Status = model.RegApproved
	next.UpdatedAt = s.clock()
	if err := s.regs.Update(ctx, &next); err != nil {
		return nil, fmt.Errorf("promote waitlist: %w", err)
	}
	return &pendingNotify{users: []uuid.UUID{next.UserID}, template: tplActivityApproved, activity: a}, nil
}

func (s *activityService) ApproveRegistration(ctx context.Context, activityID, userID uuid.UUID, approve bool, reason string) (*dto.RegistrationResponse, error) {
	var out *dto.RegistrationResponse
	var notes []*pendingNotify
	err := s.withTx(ctx, func(tx *activityService) error {
		resp, pn, err := tx.approveInTx(ctx, activityID, userID, approve)
		out = resp
		notes = pn
		return err
	})
	if err == nil {
		s.flushNotifies(ctx, notes...)
	}
	return out, err
}

func (s *activityService) approveInTx(ctx context.Context, activityID, userID uuid.UUID, approve bool) (*dto.RegistrationResponse, []*pendingNotify, error) {
	a, err := s.activities.GetByIDForUpdate(ctx, activityID)
	if err != nil {
		return nil, nil, fmt.Errorf("get activity: %w", err)
	}
	if a == nil {
		return nil, nil, response.NewError(response.CodeActivityNotFound, "活动不存在")
	}

	reg, err := s.regs.GetByActivityAndUserForUpdate(ctx, activityID, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("get registration: %w", err)
	}
	if reg == nil {
		return nil, nil, response.NewError(response.CodeRegistrationNotFound, "报名记录不存在")
	}

	wasApproved := reg.Status == model.RegApproved
	tpl := tplActivityRejected
	if approve {
		approvedCount, err := s.regs.CountByActivityAndStatus(ctx, activityID, model.RegApproved)
		if err != nil {
			return nil, nil, fmt.Errorf("count approved: %w", err)
		}
		if a.MaxParticipants > 0 && approvedCount >= int64(a.MaxParticipants) && reg.Status != model.RegApproved {
			return nil, nil, response.NewError(response.CodeActivityFull, "名额已满")
		}
		reg.Status = model.RegApproved
		tpl = tplActivityApproved
	} else {
		reg.Status = model.RegRejected
	}
	reg.UpdatedAt = s.clock()
	if err := s.regs.Update(ctx, reg); err != nil {
		return nil, nil, fmt.Errorf("approve registration: %w", err)
	}
	notes := []*pendingNotify{{users: []uuid.UUID{userID}, template: tpl, activity: a}}
	if !approve && wasApproved {
		promoted, err := s.promoteWaitlistIfSeat(ctx, a)
		if err != nil {
			return nil, nil, err
		}
		if promoted != nil {
			notes = append(notes, promoted)
		}
	}

	named, err := s.regs.ListByActivity(ctx, activityID)
	if err == nil {
		for i := range named {
			if named[i].ID == reg.ID {
				resp := toRegistrationResponse(&named[i])
				return &resp, notes, nil
			}
		}
	}
	return &dto.RegistrationResponse{ID: reg.ID.String(), ActivityID: reg.ActivityID.String(), Status: reg.Status}, notes, nil
}

func (s *activityService) ListRegistrations(ctx context.Context, activityID uuid.UUID) ([]dto.RegistrationResponse, error) {
	rows, err := s.regs.ListByActivity(ctx, activityID)
	if err != nil {
		return nil, fmt.Errorf("list registrations: %w", err)
	}
	list := make([]dto.RegistrationResponse, 0, len(rows))
	for i := range rows {
		list = append(list, toRegistrationResponse(&rows[i]))
	}
	return list, nil
}
