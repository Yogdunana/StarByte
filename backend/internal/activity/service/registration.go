package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/activity/dto"
	"github.com/Yogdunana/StarByte/backend/internal/activity/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
)

// Register 报名活动
func (s *activityService) Register(ctx context.Context, activityID, userID uuid.UUID) (*dto.RegistrationResponse, error) {
	a, err := s.activities.GetByID(ctx, activityID)
	if err != nil {
		return nil, fmt.Errorf("get activity: %w", err)
	}
	if a == nil {
		return nil, response.NewError(response.CodeActivityNotFound, "活动不存在")
	}
	if a.Status != model.ActivityOpen {
		return nil, response.NewError(response.CodeActivityInvalidState, "活动不在报名中")
	}

	// 检查是否重复报名
	existing, err := s.regs.GetByActivityAndUser(ctx, activityID, userID)
	if err != nil {
		return nil, fmt.Errorf("check registration: %w", err)
	}
	if existing != nil && existing.Status != model.RegCancelled {
		return nil, response.NewError(response.CodeRegistrationExists, "你已报名该活动")
	}

	// 计算已通过人数
	approvedCount, err := s.regs.CountByActivityAndStatus(ctx, activityID, model.RegApproved)
	if err != nil {
		return nil, fmt.Errorf("count approved: %w", err)
	}

	status := model.RegApproved
	if a.MaxParticipants > 0 && approvedCount >= int64(a.MaxParticipants) {
		status = model.RegWaitlist
	}

	reg := &model.ActivityRegistration{
		ID:            uuid.New(),
		ActivityID:    activityID,
		UserID:        userID,
		Status:        status,
		CheckinStatus: model.CheckinPending,
	}
	if existing != nil {
		reg.ID = existing.ID
		reg.CreatedAt = existing.CreatedAt
	}

	if err := s.regs.Update(ctx, reg); err != nil {
		// 若 update 失败则尝试 create
		if err := s.regs.Create(ctx, reg); err != nil {
			return nil, fmt.Errorf("save registration: %w", err)
		}
	}

	tpl := tplActivityRegistered
	if status == model.RegWaitlist {
		tpl = tplActivityWaitlist
	}
	s.notifyActivity(ctx, []uuid.UUID{userID}, tpl, a)

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
		Status:        reg.Status,
		CheckinStatus: reg.CheckinStatus,
		CreatedAt:     formatTime(reg.CreatedAt),
	}, nil
}

// CancelRegistration 取消报名
func (s *activityService) CancelRegistration(ctx context.Context, activityID, userID uuid.UUID) error {
	reg, err := s.regs.GetByActivityAndUser(ctx, activityID, userID)
	if err != nil {
		return fmt.Errorf("get registration: %w", err)
	}
	if reg == nil || reg.Status == model.RegCancelled {
		return response.NewError(response.CodeRegistrationNotFound, "未找到报名记录")
	}
	reg.Status = model.RegCancelled
	if err := s.regs.Update(ctx, reg); err != nil {
		return fmt.Errorf("cancel registration: %w", err)
	}

	// 若取消的是已通过名额，自动递补候补中最早的一人
	a, _ := s.activities.GetByID(ctx, activityID)
	if a != nil && a.MaxParticipants > 0 {
		waitlist, err := s.regs.ListWaitlist(ctx, activityID)
		if err == nil && len(waitlist) > 0 {
			next := waitlist[0]
			next.Status = model.RegApproved
			if err := s.regs.Update(ctx, &next); err == nil {
				s.notifyActivity(ctx, []uuid.UUID{next.UserID}, tplActivityApproved, a)
			}
		}
	}
	return nil
}

// ApproveRegistration 审批报名
func (s *activityService) ApproveRegistration(ctx context.Context, activityID, userID uuid.UUID, approve bool, reason string) (*dto.RegistrationResponse, error) {
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
		return nil, response.NewError(response.CodeRegistrationNotFound, "报名记录不存在")
	}

	if approve {
		// 检查是否还有名额
		approvedCount, err := s.regs.CountByActivityAndStatus(ctx, activityID, model.RegApproved)
		if err != nil {
			return nil, fmt.Errorf("count approved: %w", err)
		}
		if a.MaxParticipants > 0 && approvedCount >= int64(a.MaxParticipants) && reg.Status != model.RegApproved {
			return nil, response.NewError(response.CodeActivityFull, "名额已满")
		}
		reg.Status = model.RegApproved
		s.notifyActivity(ctx, []uuid.UUID{userID}, tplActivityApproved, a)
	} else {
		reg.Status = model.RegRejected
		s.notifyActivity(ctx, []uuid.UUID{userID}, tplActivityRejected, a)
	}
	reg.UpdatedAt = time.Now()
	if err := s.regs.Update(ctx, reg); err != nil {
		return nil, fmt.Errorf("approve registration: %w", err)
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
	return &dto.RegistrationResponse{ID: reg.ID.String(), ActivityID: reg.ActivityID.String(), Status: reg.Status}, nil
}

// ListRegistrations 报名列表
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
