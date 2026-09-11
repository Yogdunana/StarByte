package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Yogdunana/StarByte/backend/internal/activity/dto"
	"github.com/Yogdunana/StarByte/backend/internal/activity/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
)

func (s *activityService) CreateActivity(ctx context.Context, operator uuid.UUID, req *dto.CreateActivityRequest) (*dto.ActivityResponse, error) {
	if req.EndTime.Before(req.StartTime) {
		return nil, response.NewError(response.CodeBadRequest, "结束时间不能早于开始时间")
	}
	if req.MaxParticipants < 0 {
		return nil, response.NewError(response.CodeBadRequest, "人数上限不能为负数")
	}
	if err := validateGeoInput(req.Latitude, req.Longitude, req.CheckinRadiusM, false); err != nil {
		return nil, err
	}

	secret, err := newCheckinSecret()
	if err != nil {
		return nil, fmt.Errorf("generate checkin secret: %w", err)
	}

	a := &model.Activity{
		ID:              uuid.New(),
		Title:           req.Title,
		Description:     req.Description,
		Category:        req.Category,
		StartTime:       req.StartTime,
		EndTime:         req.EndTime,
		Location:        req.Location,
		Latitude:        req.Latitude,
		Longitude:       req.Longitude,
		CheckinRadiusM:  req.CheckinRadiusM,
		CheckinSecret:   secret,
		MaxParticipants: req.MaxParticipants,
		Status:          model.ActivityOpen,
		OrganizerID:     operator,
	}
	if tags, err := json.Marshal(req.Tags); err == nil {
		a.Tags = tags
	} else {
		a.Tags = []byte("[]")
	}
	if req.CoverImageID != "" {
		id, err := uuid.Parse(req.CoverImageID)
		if err == nil {
			a.CoverImageID = &id
		}
	}

	if err := s.activities.Create(ctx, a); err != nil {
		return nil, fmt.Errorf("create activity: %w", err)
	}
	return s.getActivityResponse(ctx, a.ID)
}

func (s *activityService) UpdateActivity(ctx context.Context, id uuid.UUID, req *dto.UpdateActivityRequest) (*dto.ActivityResponse, error) {
	a, err := s.activities.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get activity: %w", err)
	}
	if a == nil {
		return nil, response.NewError(response.CodeActivityNotFound, "活动不存在")
	}
	if a.Status != model.ActivityDraft && a.Status != model.ActivityOpen {
		return nil, response.NewError(response.CodeActivityInvalidState, "当前状态不允许修改")
	}

	if req.Title != nil {
		a.Title = *req.Title
	}
	if req.Description != nil {
		a.Description = *req.Description
	}
	if req.Category != nil {
		a.Category = *req.Category
	}
	if req.Tags != nil {
		if tags, err := json.Marshal(req.Tags); err == nil {
			a.Tags = tags
		}
	}
	if req.StartTime != nil {
		a.StartTime = *req.StartTime
	}
	if req.EndTime != nil {
		a.EndTime = *req.EndTime
	}
	if req.Location != nil {
		a.Location = *req.Location
	}
	if req.MaxParticipants != nil {
		if *req.MaxParticipants < 0 {
			return nil, response.NewError(response.CodeBadRequest, "人数上限不能为负数")
		}
		a.MaxParticipants = *req.MaxParticipants
	}
	if req.CoverImageID != nil {
		if *req.CoverImageID == "" {
			a.CoverImageID = nil
		} else if parsed, err := uuid.Parse(*req.CoverImageID); err == nil {
			a.CoverImageID = &parsed
		}
	}
	if req.ClearGeo {
		a.Latitude = nil
		a.Longitude = nil
		a.CheckinRadiusM = nil
	} else {
		if req.Latitude != nil {
			a.Latitude = req.Latitude
		}
		if req.Longitude != nil {
			a.Longitude = req.Longitude
		}
		if req.CheckinRadiusM != nil {
			a.CheckinRadiusM = req.CheckinRadiusM
		}
		if err := validateGeoInput(a.Latitude, a.Longitude, a.CheckinRadiusM, true); err != nil {
			return nil, err
		}
	}
	if a.EndTime.Before(a.StartTime) {
		return nil, response.NewError(response.CodeBadRequest, "结束时间不能早于开始时间")
	}

	if err := s.activities.Update(ctx, a); err != nil {
		return nil, fmt.Errorf("update activity: %w", err)
	}
	return s.getActivityResponse(ctx, a.ID)
}

func (s *activityService) DeleteActivity(ctx context.Context, id uuid.UUID) error {
	a, err := s.activities.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get activity: %w", err)
	}
	if a == nil {
		return response.NewError(response.CodeActivityNotFound, "活动不存在")
	}
	if a.Status == model.ActivityOngoing {
		return response.NewError(response.CodeActivityInvalidState, "进行中的活动不能删除")
	}
	return s.activities.Delete(ctx, id)
}

func (s *activityService) GetActivity(ctx context.Context, id uuid.UUID) (*dto.ActivityResponse, error) {
	return s.getActivityResponse(ctx, id)
}

func (s *activityService) ListActivities(ctx context.Context, req *dto.ListActivityRequest) ([]*dto.ActivityResponse, int64, error) {
	rows, total, err := s.activities.List(ctx, req)
	if err != nil {
		return nil, 0, fmt.Errorf("list activities: %w", err)
	}
	list := make([]*dto.ActivityResponse, 0, len(rows))
	for i := range rows {
		list = append(list, toActivityResponse(&rows[i]))
	}
	return list, total, nil
}

func (s *activityService) StartActivity(ctx context.Context, id uuid.UUID) (*dto.ActivityResponse, error) {
	a, err := s.activities.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get activity: %w", err)
	}
	if a == nil {
		return nil, response.NewError(response.CodeActivityNotFound, "活动不存在")
	}
	if a.Status != model.ActivityOpen {
		return nil, response.NewError(response.CodeActivityInvalidState, "只有报名中的活动可以开始")
	}
	a.Status = model.ActivityOngoing
	if err := s.activities.Update(ctx, a); err != nil {
		return nil, fmt.Errorf("start activity: %w", err)
	}
	return s.getActivityResponse(ctx, id)
}

func (s *activityService) EndActivity(ctx context.Context, id uuid.UUID) (*dto.ActivityResponse, error) {
	a, err := s.activities.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get activity: %w", err)
	}
	if a == nil {
		return nil, response.NewError(response.CodeActivityNotFound, "活动不存在")
	}
	if a.Status != model.ActivityOngoing {
		return nil, response.NewError(response.CodeActivityInvalidState, "只有进行中的活动可以结束")
	}
	a.Status = model.ActivityEnded
	if err := s.activities.Update(ctx, a); err != nil {
		return nil, fmt.Errorf("end activity: %w", err)
	}
	return s.getActivityResponse(ctx, id)
}

func (s *activityService) CancelActivity(ctx context.Context, id uuid.UUID, reason string) (*dto.ActivityResponse, error) {
	a, err := s.activities.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get activity: %w", err)
	}
	if a == nil {
		return nil, response.NewError(response.CodeActivityNotFound, "活动不存在")
	}
	if a.Status == model.ActivityEnded || a.Status == model.ActivityCancelled {
		return nil, response.NewError(response.CodeActivityInvalidState, "活动已结束或已取消")
	}
	a.Status = model.ActivityCancelled
	if err := s.activities.Update(ctx, a); err != nil {
		return nil, fmt.Errorf("cancel activity: %w", err)
	}
	return s.getActivityResponse(ctx, id)
}

func (s *activityService) getActivityResponse(ctx context.Context, id uuid.UUID) (*dto.ActivityResponse, error) {
	row, err := s.activities.GetByIDWithNames(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get activity with names: %w", err)
	}
	if row == nil {
		return nil, response.NewError(response.CodeActivityNotFound, "活动不存在")
	}
	return toActivityResponse(row), nil
}

func validateGeoInput(lat, lng *float64, radius *int, allowPartial bool) error {
	if lat == nil && lng == nil && radius == nil {
		return nil
	}
	if allowPartial && (lat == nil || lng == nil || radius == nil) {
		return nil
	}
	if lat == nil || lng == nil {
		return response.NewError(response.CodeBadRequest, "GPS 围栏需同时提供经纬度")
	}
	if !validLatLng(*lat, *lng) {
		return response.NewError(response.CodeBadRequest, "经纬度不合法")
	}
	if radius != nil && *radius < 0 {
		return response.NewError(response.CodeBadRequest, "签到半径不能为负数")
	}
	return nil
}
