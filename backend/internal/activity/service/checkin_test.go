package service

import (
	"context"
	"testing"

	"github.com/Yogdunana/StarByte/backend/internal/activity/dto"
	"github.com/Yogdunana/StarByte/backend/internal/activity/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
)

// 二维码签到成功
func TestCheckin_QR_Success(t *testing.T) {
	svc, _, _, _, _ := newTestSvc()
	resp := mustActivity(t, svc, uuid.New(), 10)
	activityID, _ := uuid.Parse(resp.ID)
	user := uuid.New()

	if _, err := svc.Register(context.Background(), activityID, user); err != nil {
		t.Fatal(err)
	}

	reg, err := svc.Checkin(context.Background(), activityID, user, &dto.CheckinRequest{
		Method: model.CheckinMethodQR,
	})
	if err != nil {
		t.Fatalf("Checkin: %v", err)
	}
	if reg.CheckinStatus != int16(model.CheckinDone) {
		t.Errorf("checkin = %d, want done", reg.CheckinStatus)
	}
	if reg.CheckinMethod == nil || *reg.CheckinMethod != model.CheckinMethodQR {
		t.Errorf("checkin method mismatch")
	}
}

// GPS 签到记录坐标
func TestCheckin_GPS_RecordsCoords(t *testing.T) {
	svc, _, rr, _, _ := newTestSvc()
	resp := mustActivity(t, svc, uuid.New(), 10)
	activityID, _ := uuid.Parse(resp.ID)
	user := uuid.New()

	if _, err := svc.Register(context.Background(), activityID, user); err != nil {
		t.Fatal(err)
	}

	lat := 31.2304
	lng := 121.4737
	_, err := svc.Checkin(context.Background(), activityID, user, &dto.CheckinRequest{
		Method:    model.CheckinMethodGPS,
		Latitude:  &lat,
		Longitude: &lng,
	})
	if err != nil {
		t.Fatalf("Checkin GPS: %v", err)
	}
	saved, _ := rr.GetByActivityAndUser(context.Background(), activityID, user)
	if saved.GPSLatitude == nil || *saved.GPSLatitude != lat {
		t.Errorf("lat = %v, want %v", saved.GPSLatitude, lat)
	}
}

// 重复签到
func TestCheckin_Duplicate(t *testing.T) {
	svc, _, _, _, _ := newTestSvc()
	resp := mustActivity(t, svc, uuid.New(), 10)
	activityID, _ := uuid.Parse(resp.ID)
	user := uuid.New()

	if _, err := svc.Register(context.Background(), activityID, user); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Checkin(context.Background(), activityID, user, &dto.CheckinRequest{Method: 1}); err != nil {
		t.Fatal(err)
	}
	_, err := svc.Checkin(context.Background(), activityID, user, &dto.CheckinRequest{Method: 1})
	if err == nil {
		t.Fatal("expected duplicate checkin error")
	}
	if appErr, ok := err.(*response.AppError); ok && appErr.Code != response.CodeCheckinAlreadyDone {
		t.Errorf("code = %d, want %d", appErr.Code, response.CodeCheckinAlreadyDone)
	}
}

// 未通过报名不能签到
func TestCheckin_NotApproved(t *testing.T) {
	svc, _, _, _, _ := newTestSvc()
	resp := mustActivity(t, svc, uuid.New(), 1)
	activityID, _ := uuid.Parse(resp.ID)
	user1 := uuid.New()
	user2 := uuid.New()

	// user1 占名额，user2 候补
	if _, err := svc.Register(context.Background(), activityID, user1); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Register(context.Background(), activityID, user2); err != nil {
		t.Fatal(err)
	}

	_, err := svc.Checkin(context.Background(), activityID, user2, &dto.CheckinRequest{Method: 1})
	if err == nil {
		t.Fatal("expected not-approved error")
	}
	if appErr, ok := err.(*response.AppError); ok && appErr.Code != response.CodeCheckinNotApproved {
		t.Errorf("code = %d, want %d", appErr.Code, response.CodeCheckinNotApproved)
	}
}

// 未报名不能签到
func TestCheckin_NotRegistered(t *testing.T) {
	svc, _, _, _, _ := newTestSvc()
	resp := mustActivity(t, svc, uuid.New(), 10)
	activityID, _ := uuid.Parse(resp.ID)

	_, err := svc.Checkin(context.Background(), activityID, uuid.New(), &dto.CheckinRequest{Method: 1})
	if err == nil {
		t.Fatal("expected not-found error")
	}
	if appErr, ok := err.(*response.AppError); ok && appErr.Code != response.CodeRegistrationNotFound {
		t.Errorf("code = %d, want %d", appErr.Code, response.CodeRegistrationNotFound)
	}
}
