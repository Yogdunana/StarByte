package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	qrcode "github.com/skip2/go-qrcode"

	"github.com/Yogdunana/StarByte/backend/internal/activity/dto"
	"github.com/Yogdunana/StarByte/backend/internal/activity/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
)

func (s *activityService) Checkin(ctx context.Context, activityID, userID uuid.UUID, req *dto.CheckinRequest) (*dto.RegistrationResponse, error) {
	a, err := s.activities.GetByID(ctx, activityID)
	if err != nil {
		return nil, fmt.Errorf("get activity: %w", err)
	}
	if a == nil {
		return nil, response.NewError(response.CodeActivityNotFound, "活动不存在")
	}
	if a.Status != model.ActivityOpen && a.Status != model.ActivityOngoing {
		return nil, response.NewError(response.CodeActivityInvalidState, "活动不在签到时间")
	}

	reg, err := s.regs.GetByActivityAndUser(ctx, activityID, userID)
	if err != nil {
		return nil, fmt.Errorf("get registration: %w", err)
	}
	if reg == nil {
		return nil, response.NewError(response.CodeRegistrationNotFound, "未找到报名记录")
	}
	if reg.Status != model.RegApproved {
		return nil, response.NewError(response.CodeCheckinNotApproved, "报名未通过，无法签到")
	}
	if reg.CheckinStatus == model.CheckinDone {
		return nil, response.NewError(response.CodeCheckinAlreadyDone, "已签到，请勿重复签到")
	}

	switch req.Method {
	case model.CheckinMethodQR:
		if err := verifyCheckinToken(a, req.Token, s.clock()); err != nil {
			return nil, response.NewError(response.CodeCheckinTokenInvalid, "签到二维码无效或已过期")
		}
	case model.CheckinMethodGPS:
		if err := s.validateGPSCheckin(a, req); err != nil {
			return nil, err
		}
	default:
		return nil, response.NewError(response.CodeBadRequest, "不支持的签到方式")
	}

	now := s.clock()
	reg.CheckinStatus = model.CheckinDone
	reg.CheckedInAt = &now
	reg.CheckinMethod = &req.Method
	if req.Method == model.CheckinMethodGPS {
		lat := *req.Latitude
		lng := *req.Longitude
		reg.GPSLatitude = &lat
		reg.GPSLongitude = &lng
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

func (s *activityService) validateGPSCheckin(a *model.Activity, req *dto.CheckinRequest) error {
	if !a.GeoConfigured() {
		return response.NewError(response.CodeCheckinGPSNotConfigured, "活动未配置签到地点与半径，暂不接受 GPS 签到")
	}
	if req.Latitude == nil || req.Longitude == nil {
		return response.NewError(response.CodeBadRequest, "GPS 签到需要经纬度")
	}
	if !validLatLng(*req.Latitude, *req.Longitude) {
		return response.NewError(response.CodeBadRequest, "经纬度不合法")
	}
	dist := haversineMeters(*a.Latitude, *a.Longitude, *req.Latitude, *req.Longitude)
	if dist > float64(*a.CheckinRadiusM) {
		return response.NewError(response.CodeCheckinGPSRejected, "签到位置超出活动围栏")
	}
	return nil
}

func (s *activityService) IssueCheckinQR(ctx context.Context, activityID uuid.UUID) (*dto.QRCodeResponse, error) {
	a, err := s.activities.GetByID(ctx, activityID)
	if err != nil {
		return nil, fmt.Errorf("get activity: %w", err)
	}
	if a == nil {
		return nil, response.NewError(response.CodeActivityNotFound, "活动不存在")
	}
	if a.Status != model.ActivityOpen && a.Status != model.ActivityOngoing {
		return nil, response.NewError(response.CodeActivityInvalidState, "当前状态不能签发签到码")
	}
	if a.CheckinSecret == "" {
		secret, err := newCheckinSecret()
		if err != nil {
			return nil, fmt.Errorf("generate checkin secret: %w", err)
		}
		a.CheckinSecret = secret
	}
	a.CheckinNonce++
	if err := s.activities.Update(ctx, a); err != nil {
		return nil, fmt.Errorf("rotate checkin nonce: %w", err)
	}
	exp := s.clock().Add(s.tokenTTL)
	token := signCheckinToken(a.CheckinSecret, a.ID, a.CheckinNonce, exp)
	path := "/activity/checkin?activity_id=" + a.ID.String() + "&token=" + token
	png, err := qrcode.Encode(path, qrcode.Medium, 256)
	if err != nil {
		return nil, fmt.Errorf("encode qr: %w", err)
	}
	return &dto.QRCodeResponse{
		ActivityID:  a.ID.String(),
		Token:       token,
		ExpiresAt:   formatTime(exp),
		CheckinPath: path,
		PNGBase64:   base64.StdEncoding.EncodeToString(png),
	}, nil
}

func formatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}
