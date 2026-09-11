package service

import (
	"context"
	"testing"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/activity/dto"
	"github.com/Yogdunana/StarByte/backend/internal/activity/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
)

func TestCheckin_QR_Success(t *testing.T) {
	svc, _, _, _, _ := newTestSvc()
	resp := mustActivity(t, svc, uuid.New(), 10)
	activityID, _ := uuid.Parse(resp.ID)
	user := uuid.New()

	if _, err := svc.Register(context.Background(), activityID, user); err != nil {
		t.Fatal(err)
	}
	token := mustIssueToken(t, svc, activityID)

	reg, err := svc.Checkin(context.Background(), activityID, user, &dto.CheckinRequest{
		Method: model.CheckinMethodQR,
		Token:  token,
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

func TestCheckin_QR_MissingToken(t *testing.T) {
	svc, _, _, _, _ := newTestSvc()
	resp := mustActivity(t, svc, uuid.New(), 10)
	activityID, _ := uuid.Parse(resp.ID)
	user := uuid.New()
	if _, err := svc.Register(context.Background(), activityID, user); err != nil {
		t.Fatal(err)
	}
	_, err := svc.Checkin(context.Background(), activityID, user, &dto.CheckinRequest{Method: model.CheckinMethodQR})
	if err == nil {
		t.Fatal("expected missing token error")
	}
	if appErr, ok := err.(*response.AppError); ok && appErr.Code != response.CodeCheckinTokenInvalid {
		t.Errorf("code = %d, want %d", appErr.Code, response.CodeCheckinTokenInvalid)
	}
}

func TestCheckin_QR_InvalidToken(t *testing.T) {
	svc, _, _, _, _ := newTestSvc()
	resp := mustActivity(t, svc, uuid.New(), 10)
	activityID, _ := uuid.Parse(resp.ID)
	user := uuid.New()
	if _, err := svc.Register(context.Background(), activityID, user); err != nil {
		t.Fatal(err)
	}
	_, err := svc.Checkin(context.Background(), activityID, user, &dto.CheckinRequest{
		Method: model.CheckinMethodQR,
		Token:  "v1.1.9999999999.deadbeef",
	})
	if err == nil {
		t.Fatal("expected invalid token")
	}
	if appErr, ok := err.(*response.AppError); ok && appErr.Code != response.CodeCheckinTokenInvalid {
		t.Errorf("code = %d, want %d", appErr.Code, response.CodeCheckinTokenInvalid)
	}
}

func TestCheckin_QR_ExpiredToken(t *testing.T) {
	svc, _, _, _, _ := newTestSvc()
	frozen := time.Date(2026, 9, 11, 12, 0, 0, 0, time.UTC)
	svc.now = func() time.Time { return frozen }
	resp := mustActivity(t, svc, uuid.New(), 10)
	activityID, _ := uuid.Parse(resp.ID)
	user := uuid.New()
	if _, err := svc.Register(context.Background(), activityID, user); err != nil {
		t.Fatal(err)
	}
	token := mustIssueToken(t, svc, activityID)
	svc.now = func() time.Time { return frozen.Add(20 * time.Minute) }
	_, err := svc.Checkin(context.Background(), activityID, user, &dto.CheckinRequest{
		Method: model.CheckinMethodQR,
		Token:  token,
	})
	if err == nil {
		t.Fatal("expected expired token")
	}
	if appErr, ok := err.(*response.AppError); ok && appErr.Code != response.CodeCheckinTokenInvalid {
		t.Errorf("code = %d, want %d", appErr.Code, response.CodeCheckinTokenInvalid)
	}
}

func TestCheckin_QR_RotatedNonce(t *testing.T) {
	svc, _, _, _, _ := newTestSvc()
	resp := mustActivity(t, svc, uuid.New(), 10)
	activityID, _ := uuid.Parse(resp.ID)
	user := uuid.New()
	if _, err := svc.Register(context.Background(), activityID, user); err != nil {
		t.Fatal(err)
	}
	old := mustIssueToken(t, svc, activityID)
	_ = mustIssueToken(t, svc, activityID)
	_, err := svc.Checkin(context.Background(), activityID, user, &dto.CheckinRequest{
		Method: model.CheckinMethodQR,
		Token:  old,
	})
	if err == nil {
		t.Fatal("expected rotated token rejection")
	}
	if appErr, ok := err.(*response.AppError); ok && appErr.Code != response.CodeCheckinTokenInvalid {
		t.Errorf("code = %d, want %d", appErr.Code, response.CodeCheckinTokenInvalid)
	}
}

func TestCheckin_GPS_NotConfigured(t *testing.T) {
	svc, _, _, _, _ := newTestSvc()
	resp := mustActivity(t, svc, uuid.New(), 10)
	activityID, _ := uuid.Parse(resp.ID)
	user := uuid.New()
	if _, err := svc.Register(context.Background(), activityID, user); err != nil {
		t.Fatal(err)
	}
	lat, lng := 31.2304, 121.4737
	_, err := svc.Checkin(context.Background(), activityID, user, &dto.CheckinRequest{
		Method:    model.CheckinMethodGPS,
		Latitude:  &lat,
		Longitude: &lng,
	})
	if err == nil {
		t.Fatal("expected GPS not configured")
	}
	if appErr, ok := err.(*response.AppError); ok && appErr.Code != response.CodeCheckinGPSNotConfigured {
		t.Errorf("code = %d, want %d", appErr.Code, response.CodeCheckinGPSNotConfigured)
	}
}

func TestCheckin_GPS_SpoofRejected(t *testing.T) {
	svc, _, _, _, _ := newTestSvc()
	lat, lng, radius := 31.2304, 121.4737, 100
	start := time.Now().Add(time.Hour)
	resp, err := svc.CreateActivity(context.Background(), uuid.New(), &dto.CreateActivityRequest{
		Title:           "围栏活动",
		StartTime:       start,
		EndTime:         start.Add(time.Hour),
		Location:        "教学楼",
		Latitude:        &lat,
		Longitude:       &lng,
		CheckinRadiusM:  &radius,
		MaxParticipants: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	activityID, _ := uuid.Parse(resp.ID)
	user := uuid.New()
	if _, err := svc.Register(context.Background(), activityID, user); err != nil {
		t.Fatal(err)
	}
	spoofLat := 31.2404 // ~1.1km north
	_, err = svc.Checkin(context.Background(), activityID, user, &dto.CheckinRequest{
		Method:    model.CheckinMethodGPS,
		Latitude:  &spoofLat,
		Longitude: &lng,
	})
	if err == nil {
		t.Fatal("expected spoof rejection")
	}
	if appErr, ok := err.(*response.AppError); ok && appErr.Code != response.CodeCheckinGPSRejected {
		t.Errorf("code = %d, want %d", appErr.Code, response.CodeCheckinGPSRejected)
	}
}

func TestCheckin_GPS_InsideFence(t *testing.T) {
	svc, _, rr, _, _ := newTestSvc()
	lat, lng, radius := 31.2304, 121.4737, 200
	start := time.Now().Add(time.Hour)
	resp, err := svc.CreateActivity(context.Background(), uuid.New(), &dto.CreateActivityRequest{
		Title:           "围栏活动",
		StartTime:       start,
		EndTime:         start.Add(time.Hour),
		Location:        "教学楼",
		Latitude:        &lat,
		Longitude:       &lng,
		CheckinRadiusM:  &radius,
		MaxParticipants: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	activityID, _ := uuid.Parse(resp.ID)
	user := uuid.New()
	if _, err := svc.Register(context.Background(), activityID, user); err != nil {
		t.Fatal(err)
	}
	_, err = svc.Checkin(context.Background(), activityID, user, &dto.CheckinRequest{
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

func TestCheckin_Duplicate(t *testing.T) {
	svc, _, _, _, _ := newTestSvc()
	resp := mustActivity(t, svc, uuid.New(), 10)
	activityID, _ := uuid.Parse(resp.ID)
	user := uuid.New()

	if _, err := svc.Register(context.Background(), activityID, user); err != nil {
		t.Fatal(err)
	}
	token := mustIssueToken(t, svc, activityID)
	if _, err := svc.Checkin(context.Background(), activityID, user, &dto.CheckinRequest{Method: 1, Token: token}); err != nil {
		t.Fatal(err)
	}
	_, err := svc.Checkin(context.Background(), activityID, user, &dto.CheckinRequest{Method: 1, Token: token})
	if err == nil {
		t.Fatal("expected duplicate checkin error")
	}
	if appErr, ok := err.(*response.AppError); ok && appErr.Code != response.CodeCheckinAlreadyDone {
		t.Errorf("code = %d, want %d", appErr.Code, response.CodeCheckinAlreadyDone)
	}
}

func TestCheckin_NotApproved(t *testing.T) {
	svc, _, _, _, _ := newTestSvc()
	resp := mustActivity(t, svc, uuid.New(), 1)
	activityID, _ := uuid.Parse(resp.ID)
	user1 := uuid.New()
	user2 := uuid.New()

	if _, err := svc.Register(context.Background(), activityID, user1); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Register(context.Background(), activityID, user2); err != nil {
		t.Fatal(err)
	}
	token := mustIssueToken(t, svc, activityID)

	_, err := svc.Checkin(context.Background(), activityID, user2, &dto.CheckinRequest{Method: 1, Token: token})
	if err == nil {
		t.Fatal("expected not-approved error")
	}
	if appErr, ok := err.(*response.AppError); ok && appErr.Code != response.CodeCheckinNotApproved {
		t.Errorf("code = %d, want %d", appErr.Code, response.CodeCheckinNotApproved)
	}
}

func TestCheckin_NotRegistered(t *testing.T) {
	svc, _, _, _, _ := newTestSvc()
	resp := mustActivity(t, svc, uuid.New(), 10)
	activityID, _ := uuid.Parse(resp.ID)
	token := mustIssueToken(t, svc, activityID)

	_, err := svc.Checkin(context.Background(), activityID, uuid.New(), &dto.CheckinRequest{Method: 1, Token: token})
	if err == nil {
		t.Fatal("expected not-found error")
	}
	if appErr, ok := err.(*response.AppError); ok && appErr.Code != response.CodeRegistrationNotFound {
		t.Errorf("code = %d, want %d", appErr.Code, response.CodeRegistrationNotFound)
	}
}

func TestCheckin_AfterCancel_DoesNotRestoreApproved(t *testing.T) {
	svc, _, rr, _, _ := newTestSvc()
	resp := mustActivity(t, svc, uuid.New(), 1)
	activityID, _ := uuid.Parse(resp.ID)
	user1 := uuid.New()
	user2 := uuid.New()

	if _, err := svc.Register(context.Background(), activityID, user1); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Register(context.Background(), activityID, user2); err != nil {
		t.Fatal(err)
	}
	token := mustIssueToken(t, svc, activityID)
	if err := svc.CancelRegistration(context.Background(), activityID, user1); err != nil {
		t.Fatal(err)
	}

	_, err := svc.Checkin(context.Background(), activityID, user1, &dto.CheckinRequest{Method: 1, Token: token})
	if err == nil {
		t.Fatal("cancelled user must not check in")
	}
	if appErr, ok := err.(*response.AppError); ok && appErr.Code != response.CodeCheckinNotApproved {
		t.Errorf("code = %d, want %d", appErr.Code, response.CodeCheckinNotApproved)
	}

	cancelled, err := rr.GetByActivityAndUser(context.Background(), activityID, user1)
	if err != nil || cancelled == nil {
		t.Fatalf("cancelled row: %v %v", cancelled, err)
	}
	if cancelled.Status != model.RegCancelled {
		t.Errorf("status = %d, want cancelled", cancelled.Status)
	}
	if cancelled.CheckinStatus != model.CheckinPending {
		t.Errorf("checkin_status = %d, want pending", cancelled.CheckinStatus)
	}

	promoted, err := rr.GetByActivityAndUser(context.Background(), activityID, user2)
	if err != nil || promoted == nil {
		t.Fatalf("promoted row: %v %v", promoted, err)
	}
	if promoted.Status != model.RegApproved {
		t.Errorf("waitlist user status = %d, want approved", promoted.Status)
	}
}

func TestMarkCheckedIn_IgnoresCancelledRow(t *testing.T) {
	svc, _, rr, _, _ := newTestSvc()
	resp := mustActivity(t, svc, uuid.New(), 10)
	activityID, _ := uuid.Parse(resp.ID)
	user := uuid.New()
	if _, err := svc.Register(context.Background(), activityID, user); err != nil {
		t.Fatal(err)
	}
	if err := svc.CancelRegistration(context.Background(), activityID, user); err != nil {
		t.Fatal(err)
	}
	row, err := rr.GetByActivityAndUser(context.Background(), activityID, user)
	if err != nil || row == nil {
		t.Fatal(err)
	}
	affected, err := rr.MarkCheckedIn(context.Background(), row.ID, time.Now(), model.CheckinMethodQR, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if affected != 0 {
		t.Errorf("affected = %d, want 0", affected)
	}
	saved, _ := rr.GetByActivityAndUser(context.Background(), activityID, user)
	if saved.Status != model.RegCancelled {
		t.Errorf("status = %d, want cancelled (stale checkin must not restore approved)", saved.Status)
	}
	if saved.CheckinStatus != model.CheckinPending {
		t.Errorf("checkin_status = %d, want pending", saved.CheckinStatus)
	}
}

func TestHaversineNearby(t *testing.T) {
	d := haversineMeters(31.2304, 121.4737, 31.2304, 121.4737)
	if d > 1 {
		t.Errorf("same point distance = %f", d)
	}
	far := haversineMeters(31.2304, 121.4737, 31.2404, 121.4737)
	if far < 800 {
		t.Errorf("expected spoof point >800m, got %f", far)
	}
}
