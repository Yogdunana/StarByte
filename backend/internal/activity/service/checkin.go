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

// Checkin 签到
func (s *activityService) Checkin(ctx context.Context, activityID, userID uuid.UUID, req *dto.CheckinRequest) (*dto.RegistrationResponse, error) {
	a, err := s.activities.GetByID(ctx, activityID)
	if err != nil {
		return nil, fmt.Errorf("get activity: %w", err)
	}
	if a == nil {
		return nil, response.NewError(CodeActivityNotFound, "活动不存在")
	}
	if a.Status != model.ActivityOpen && a.Status != model.ActivityOngoing {
		return nil, response.NewError(CodeActivityInvalidState, "活动不在签到时间")
	}

	reg, err := s.regs.GetByActivityAndUser(ctx, activityID, userID)
	if err != nil {
		return nil, fmt.Errorf("get registration: %w", err)
	}
	if reg == nil {
		return nil, response.NewError(CodeRegistrationNotFound, "未找到报名记录")
	}
	if reg.Status != model.RegApproved {
		return nil, response.NewError(CodeCheckinNotApproved, "报名未通过，无法签到")
	}
	if reg.CheckinStatus == model.CheckinDone {
		return nil, response.NewError(response.CodeCheckinAlreadyDone, "已签到，请勿重复签到")
	}

	now := time.Now()
	reg.CheckinStatus = model.CheckinDone
	reg.CheckedInAt = &now
	reg.CheckinMethod = &req.Method

	if req.Method == model.CheckinMethodGPS {
		if req.Latitude != nil {
			lat := *req.Latitude
			reg.GPSLatitude = &lat
		}
		if req.Longitude != nil {
			lng := *req.Longitude
			reg.GPSLongitude = &lng
		}
	}

	if err := s.regs.Update(ctx, reg); err != nil {
		return nil, fmt.Errorf("checkin: %w", err)
	}

	return &dto.RegistrationResponse{
		ID:            reg.ID.String(),
		ActivityID:    reg.ActivityID.String(),
		Status:        reg.Status,
		CheckinStatus: reg.CheckinStatus,
		CheckedInAt:   formatTime(now),
		CheckinMethod: reg.CheckinMethod,
		CreatedAt:     formatTime(reg.CreatedAt),
	}, nil
}
