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

func mustActivity(t *testing.T, svc ActivityService, organizer uuid.UUID, maxP int) *dto.ActivityResponse {
	t.Helper()
	start := time.Now().Add(24 * time.Hour)
	end := start.Add(2 * time.Hour)
	resp, err := svc.CreateActivity(context.Background(), organizer, &dto.CreateActivityRequest{
		Title:           "测试活动",
		Description:     "描述",
		Category:        "讲座",
		Tags:            []string{"Go", "后端"},
		StartTime:       start,
		EndTime:         end,
		Location:        "教学楼101",
		MaxParticipants: maxP,
	})
	if err != nil {
		t.Fatalf("CreateActivity failed: %v", err)
	}
	return resp
}

func mustIssueToken(t *testing.T, svc ActivityService, activityID uuid.UUID) string {
	t.Helper()
	qr, err := svc.IssueCheckinQR(context.Background(), activityID)
	if err != nil {
		t.Fatalf("IssueCheckinQR: %v", err)
	}
	return qr.Token
}

func TestCreateActivity_Success(t *testing.T) {
	svc, _, _, _, _ := newTestSvc()
	organizer := uuid.New()
	resp := mustActivity(t, svc, organizer, 50)

	if resp.Title != "测试活动" {
		t.Errorf("title = %q, want 测试活动", resp.Title)
	}
	if resp.Status != int16(model.ActivityOpen) {
		t.Errorf("status = %d, want %d", resp.Status, model.ActivityOpen)
	}
	if resp.Organizer.ID != organizer.String() {
		t.Errorf("organizer mismatch")
	}
	if len(resp.Tags) != 2 {
		t.Errorf("tags len = %d, want 2", len(resp.Tags))
	}
}

func TestCreateActivity_EndBeforeStart(t *testing.T) {
	svc, _, _, _, _ := newTestSvc()
	start := time.Now().Add(24 * time.Hour)
	_, err := svc.CreateActivity(context.Background(), uuid.New(), &dto.CreateActivityRequest{
		Title:     "时间错误",
		StartTime: start,
		EndTime:   start.Add(-time.Hour),
	})
	if err == nil {
		t.Fatal("expected error for end before start")
	}
	if appErr, ok := err.(*response.AppError); ok {
		if appErr.Code != response.CodeBadRequest {
			t.Errorf("code = %d, want %d", appErr.Code, response.CodeBadRequest)
		}
	}
}

func TestGetActivity_NotFound(t *testing.T) {
	svc, _, _, _, _ := newTestSvc()
	_, err := svc.GetActivity(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected not found error")
	}
	if appErr, ok := err.(*response.AppError); ok {
		if appErr.Code != response.CodeActivityNotFound {
			t.Errorf("code = %d, want %d", appErr.Code, response.CodeActivityNotFound)
		}
	}
}

func TestUpdateActivity_Success(t *testing.T) {
	svc, _, _, _, _ := newTestSvc()
	resp := mustActivity(t, svc, uuid.New(), 50)
	id, _ := uuid.Parse(resp.ID)

	newTitle := "更新后的标题"
	newMax := 100
	updated, err := svc.UpdateActivity(context.Background(), id, &dto.UpdateActivityRequest{
		Title:           &newTitle,
		MaxParticipants: &newMax,
	})
	if err != nil {
		t.Fatalf("UpdateActivity failed: %v", err)
	}
	if updated.Title != newTitle {
		t.Errorf("title = %q, want %q", updated.Title, newTitle)
	}
	if updated.MaxParticipants != newMax {
		t.Errorf("max = %d, want %d", updated.MaxParticipants, newMax)
	}
}

func TestDeleteActivity_HiddenFromListAndDetail(t *testing.T) {
	svc, _, _, _, _ := newTestSvc()
	resp := mustActivity(t, svc, uuid.New(), 50)
	id, _ := uuid.Parse(resp.ID)

	if err := svc.DeleteActivity(context.Background(), id); err != nil {
		t.Fatalf("DeleteActivity: %v", err)
	}
	if _, err := svc.GetActivity(context.Background(), id); err == nil {
		t.Fatal("deleted activity should not be visible via GetActivity")
	}
	list, total, err := svc.ListActivities(context.Background(), &dto.ListActivityRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if total != 0 || len(list) != 0 {
		t.Errorf("list after delete = %d items total %d, want 0", len(list), total)
	}
}

func TestUpdateActivity_ClearGeo(t *testing.T) {
	svc, _, _, _, _ := newTestSvc()
	lat, lng, radius := 31.23, 121.47, 80
	start := time.Now().Add(time.Hour)
	resp, err := svc.CreateActivity(context.Background(), uuid.New(), &dto.CreateActivityRequest{
		Title:           "围栏活动",
		StartTime:       start,
		EndTime:         start.Add(time.Hour),
		Latitude:        &lat,
		Longitude:       &lng,
		CheckinRadiusM:  &radius,
		MaxParticipants: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !resp.GPSEnabled {
		t.Fatal("expected gps_enabled after create")
	}
	id, _ := uuid.Parse(resp.ID)
	updated, err := svc.UpdateActivity(context.Background(), id, &dto.UpdateActivityRequest{ClearGeo: true})
	if err != nil {
		t.Fatalf("ClearGeo: %v", err)
	}
	if updated.GPSEnabled || updated.Latitude != nil || updated.Longitude != nil || updated.CheckinRadiusM != nil {
		t.Errorf("geo should be cleared, got lat=%v lng=%v r=%v enabled=%v",
			updated.Latitude, updated.Longitude, updated.CheckinRadiusM, updated.GPSEnabled)
	}
}

func TestFormatTime_RFC3339(t *testing.T) {
	ts := time.Date(2026, 9, 11, 12, 0, 0, 0, time.UTC)
	got := formatTime(ts)
	if got != "2026-09-11T12:00:00Z" {
		t.Errorf("formatTime = %q, want RFC3339", got)
	}
	if _, err := time.Parse(time.RFC3339, got); err != nil {
		t.Errorf("formatTime output not RFC3339: %v", err)
	}
}

func TestDeleteActivity_OngoingForbidden(t *testing.T) {
	svc, _, _, _, _ := newTestSvc()
	resp := mustActivity(t, svc, uuid.New(), 50)
	id, _ := uuid.Parse(resp.ID)

	if _, err := svc.StartActivity(context.Background(), id); err != nil {
		t.Fatalf("StartActivity failed: %v", err)
	}
	err := svc.DeleteActivity(context.Background(), id)
	if err == nil {
		t.Fatal("expected error deleting ongoing activity")
	}
	if appErr, ok := err.(*response.AppError); ok && appErr.Code != response.CodeActivityInvalidState {
		t.Errorf("code = %d, want %d", appErr.Code, response.CodeActivityInvalidState)
	}
}

func TestStartEndCancelFlow(t *testing.T) {
	svc, _, _, _, _ := newTestSvc()
	resp := mustActivity(t, svc, uuid.New(), 50)
	id, _ := uuid.Parse(resp.ID)

	started, err := svc.StartActivity(context.Background(), id)
	if err != nil {
		t.Fatalf("StartActivity: %v", err)
	}
	if started.Status != int16(model.ActivityOngoing) {
		t.Errorf("after start status = %d", started.Status)
	}

	if _, err := svc.StartActivity(context.Background(), id); err == nil {
		t.Error("expected error starting twice")
	}

	ended, err := svc.EndActivity(context.Background(), id)
	if err != nil {
		t.Fatalf("EndActivity: %v", err)
	}
	if ended.Status != int16(model.ActivityEnded) {
		t.Errorf("after end status = %d", ended.Status)
	}

	if _, err := svc.CancelActivity(context.Background(), id, ""); err == nil {
		t.Error("expected error cancelling ended activity")
	}
}

func TestCancelActivity_Success(t *testing.T) {
	svc, _, _, _, _ := newTestSvc()
	resp := mustActivity(t, svc, uuid.New(), 50)
	id, _ := uuid.Parse(resp.ID)

	cancelled, err := svc.CancelActivity(context.Background(), id, "天气原因")
	if err != nil {
		t.Fatalf("CancelActivity: %v", err)
	}
	if cancelled.Status != int16(model.ActivityCancelled) {
		t.Errorf("status = %d, want cancelled", cancelled.Status)
	}
}

func TestActivityStatusConstants(t *testing.T) {
	if model.ActivityOpen != 1 {
		t.Errorf("ActivityOpen = %d, want 1", model.ActivityOpen)
	}
	if model.RegWaitlist != 3 {
		t.Errorf("RegWaitlist = %d, want 3", model.RegWaitlist)
	}
	if model.CheckinMethodQR != 1 {
		t.Errorf("CheckinMethodQR = %d, want 1", model.CheckinMethodQR)
	}
}

func TestErrorCodesInActivityRange(t *testing.T) {
	r := response.ModuleRanges["activity"]
	codes := []int{
		response.CodeActivityNotFound,
		response.CodeActivityInvalidState,
		response.CodeActivityFull,
		response.CodeRegistrationExists,
		response.CodeRegistrationNotFound,
		response.CodeCheckinFailed,
		response.CodeCheckinAlreadyDone,
		response.CodeCheckinNotApproved,
		response.CodeSurveyAlreadySubmitted,
		response.CodeSurveyNotEnded,
		response.CodeCheckinTokenInvalid,
		response.CodeCheckinGPSRejected,
		response.CodeCheckinGPSNotConfigured,
	}
	for _, c := range codes {
		if c < r[0] || c > r[1] {
			t.Errorf("code %d not in activity range %d-%d", c, r[0], r[1])
		}
	}
	duty := response.ModuleRanges["duty"]
	if duty[0] != 26000 || duty[1] != 26999 {
		t.Errorf("duty range = %v, want 26000-26999", duty)
	}
}
