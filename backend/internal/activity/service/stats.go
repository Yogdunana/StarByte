package service

import (
	"context"
	"fmt"

	"github.com/Yogdunana/StarByte/backend/internal/activity/dto"
	"github.com/Yogdunana/StarByte/backend/internal/activity/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
)

// GetStats 活动统计
func (s *activityService) GetStats(ctx context.Context, activityID uuid.UUID) (*dto.ActivityStatsResponse, error) {
	a, err := s.activities.GetByID(ctx, activityID)
	if err != nil {
		return nil, fmt.Errorf("get activity: %w", err)
	}
	if a == nil {
		return nil, response.NewError(response.CodeActivityNotFound, "活动不存在")
	}

	approvedCount, err := s.regs.CountByActivityAndStatus(ctx, activityID, model.RegApproved)
	if err != nil {
		return nil, fmt.Errorf("count approved: %w", err)
	}
	waitlistCount, err := s.regs.CountByActivityAndStatus(ctx, activityID, model.RegWaitlist)
	if err != nil {
		return nil, fmt.Errorf("count waitlist: %w", err)
	}
	checkedInCount, err := s.regs.CountCheckedIn(ctx, activityID)
	if err != nil {
		return nil, fmt.Errorf("count checked in: %w", err)
	}
	totalReg, err := s.regs.CountByActivityAndStatus(ctx, activityID, model.RegApproved)
	if err != nil {
		return nil, fmt.Errorf("count total: %w", err)
	}

	surveyCount, avgRating, dist, err := s.surveys.StatsByActivity(ctx, activityID)
	if err != nil {
		return nil, fmt.Errorf("survey stats: %w", err)
	}

	registerRate := 0.0
	if a.MaxParticipants > 0 {
		registerRate = float64(approvedCount) / float64(a.MaxParticipants) * 100
	}
	attendRate := 0.0
	if approvedCount > 0 {
		attendRate = float64(checkedInCount) / float64(approvedCount) * 100
	}

	ratingDist := make(map[string]int64)
	for k, v := range dist {
		ratingDist[fmt.Sprintf("%d", k)] = v
	}

	return &dto.ActivityStatsResponse{
		ActivityID:      activityID.String(),
		MaxParticipants: a.MaxParticipants,
		RegisteredCount: totalReg,
		ApprovedCount:   approvedCount,
		WaitlistCount:   waitlistCount,
		CheckedInCount:  checkedInCount,
		RegisterRate:    registerRate,
		AttendRate:      attendRate,
		SurveyCount:     surveyCount,
		AvgRating:       avgRating,
		RatingDist:      ratingDist,
	}, nil
}

// SubmitSurvey 提交满意度调查
func (s *activityService) SubmitSurvey(ctx context.Context, activityID, userID uuid.UUID, req *dto.SurveyRequest) error {
	a, err := s.activities.GetByID(ctx, activityID)
	if err != nil {
		return fmt.Errorf("get activity: %w", err)
	}
	if a == nil {
		return response.NewError(response.CodeActivityNotFound, "活动不存在")
	}
	if a.Status != model.ActivityEnded {
		return response.NewError(response.CodeSurveyNotEnded, "活动未结束，暂不能评价")
	}

	// 必须已签到才能评价
	reg, err := s.regs.GetByActivityAndUser(ctx, activityID, userID)
	if err != nil {
		return fmt.Errorf("get registration: %w", err)
	}
	if reg == nil || reg.CheckinStatus != model.CheckinDone {
		return response.NewError(response.CodeCheckinNotApproved, "仅已签到参与者可评价")
	}

	existing, err := s.surveys.GetByActivityAndUser(ctx, activityID, userID)
	if err != nil {
		return fmt.Errorf("check survey: %w", err)
	}
	if existing != nil {
		return response.NewError(response.CodeSurveyAlreadySubmitted, "已提交过评价")
	}

	survey := &model.ActivitySurvey{
		ID:         uuid.New(),
		ActivityID: activityID,
		UserID:     userID,
		Rating:     req.Rating,
		Comment:    req.Comment,
	}
	return s.surveys.Create(ctx, survey)
}
