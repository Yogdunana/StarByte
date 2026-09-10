package service

import (
	"context"
	"testing"

	"github.com/Yogdunana/StarByte/backend/internal/activity/dto"
	"github.com/Yogdunana/StarByte/backend/internal/activity/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
)

func TestGetStats(t *testing.T) {
	svc, _, _, _, _ := newTestSvc()
	resp := mustActivity(t, svc, uuid.New(), 10)
	activityID, _ := uuid.Parse(resp.ID)

	user1 := uuid.New()
	user2 := uuid.New()
	if _, err := svc.Register(context.Background(), activityID, user1); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Register(context.Background(), activityID, user2); err != nil {
		t.Fatal(err)
	}
	// user1 签到
	if _, err := svc.Checkin(context.Background(), activityID, user1, &dto.CheckinRequest{Method: 1}); err != nil {
		t.Fatal(err)
	}

	stats, err := svc.GetStats(context.Background(), activityID)
	if err != nil {
		t.Fatalf("GetStats: %v", err)
	}
	if stats.ApprovedCount != 2 {
		t.Errorf("approved = %d, want 2", stats.ApprovedCount)
	}
	if stats.CheckedInCount != 1 {
		t.Errorf("checkedIn = %d, want 1", stats.CheckedInCount)
	}
	if stats.RegisterRate != 20.0 { // 2/10
		t.Errorf("registerRate = %f, want 20", stats.RegisterRate)
	}
	if stats.AttendRate != 50.0 { // 1/2
		t.Errorf("attendRate = %f, want 50", stats.AttendRate)
	}
}

func TestSubmitSurvey_Success(t *testing.T) {
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
	// 结束活动才能评价
	if _, err := svc.StartActivity(context.Background(), activityID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.EndActivity(context.Background(), activityID); err != nil {
		t.Fatal(err)
	}

	if err := svc.SubmitSurvey(context.Background(), activityID, user, &dto.SurveyRequest{
		Rating: 5, Comment: "很棒",
	}); err != nil {
		t.Fatalf("SubmitSurvey: %v", err)
	}

	stats, err := svc.GetStats(context.Background(), activityID)
	if err != nil {
		t.Fatal(err)
	}
	if stats.SurveyCount != 1 {
		t.Errorf("surveyCount = %d, want 1", stats.SurveyCount)
	}
	if stats.AvgRating != 5.0 {
		t.Errorf("avgRating = %f, want 5", stats.AvgRating)
	}
	if stats.RatingDist["5"] != 1 {
		t.Errorf("rating dist = %v", stats.RatingDist)
	}
}

func TestSubmitSurvey_NotEnded(t *testing.T) {
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
	// 活动未结束
	err := svc.SubmitSurvey(context.Background(), activityID, user, &dto.SurveyRequest{Rating: 4})
	if err == nil {
		t.Fatal("expected not-ended error")
	}
	if appErr, ok := err.(*response.AppError); ok && appErr.Code != response.CodeSurveyNotEnded {
		t.Errorf("code = %d, want %d", appErr.Code, response.CodeSurveyNotEnded)
	}
}

func TestSubmitSurvey_Duplicate(t *testing.T) {
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
	if _, err := svc.StartActivity(context.Background(), activityID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.EndActivity(context.Background(), activityID); err != nil {
		t.Fatal(err)
	}

	if err := svc.SubmitSurvey(context.Background(), activityID, user, &dto.SurveyRequest{Rating: 5}); err != nil {
		t.Fatal(err)
	}
	err := svc.SubmitSurvey(context.Background(), activityID, user, &dto.SurveyRequest{Rating: 3})
	if err == nil {
		t.Fatal("expected duplicate error")
	}
	if appErr, ok := err.(*response.AppError); ok && appErr.Code != response.CodeSurveyAlreadySubmitted {
		t.Errorf("code = %d, want %d", appErr.Code, response.CodeSurveyAlreadySubmitted)
	}
}

func TestSubmitSurvey_NotCheckedIn(t *testing.T) {
	svc, _, _, _, _ := newTestSvc()
	resp := mustActivity(t, svc, uuid.New(), 10)
	activityID, _ := uuid.Parse(resp.ID)
	user := uuid.New()

	if _, err := svc.Register(context.Background(), activityID, user); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.StartActivity(context.Background(), activityID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.EndActivity(context.Background(), activityID); err != nil {
		t.Fatal(err)
	}
	// 未签到
	err := svc.SubmitSurvey(context.Background(), activityID, user, &dto.SurveyRequest{Rating: 4})
	if err == nil {
		t.Fatal("expected error")
	}
	if appErr, ok := err.(*response.AppError); ok && appErr.Code != response.CodeCheckinNotApproved {
		t.Errorf("code = %d, want %d", appErr.Code, response.CodeCheckinNotApproved)
	}
}

// 确保状态常量正确
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
