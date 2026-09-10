package service

import (
	"context"
	"testing"

	"github.com/Yogdunana/StarByte/backend/internal/activity/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
)

// 报名：未满直接通过
func TestRegister_AutoApprove(t *testing.T) {
	svc, _, _, _, nn := newTestSvc()
	resp := mustActivity(t, svc, uuid.New(), 2)
	activityID, _ := uuid.Parse(resp.ID)
	user := uuid.New()

	reg, err := svc.Register(context.Background(), activityID, user)
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if reg.Status != int16(model.RegApproved) {
		t.Errorf("status = %d, want approved(%d)", reg.Status, model.RegApproved)
	}
	if len(nn.sent) == 0 || nn.sent[0] != tplActivityRegistered {
		t.Errorf("notify template = %v, want %s", nn.sent, tplActivityRegistered)
	}
}

// 报名：满了进候补
func TestRegister_Waitlist(t *testing.T) {
	svc, _, _, _, _ := newTestSvc()
	resp := mustActivity(t, svc, uuid.New(), 1)
	activityID, _ := uuid.Parse(resp.ID)

	user1 := uuid.New()
	user2 := uuid.New()

	if _, err := svc.Register(context.Background(), activityID, user1); err != nil {
		t.Fatalf("register user1: %v", err)
	}
	reg2, err := svc.Register(context.Background(), activityID, user2)
	if err != nil {
		t.Fatalf("register user2: %v", err)
	}
	if reg2.Status != int16(model.RegWaitlist) {
		t.Errorf("user2 status = %d, want waitlist(%d)", reg2.Status, model.RegWaitlist)
	}
}

// 重复报名
func TestRegister_Duplicate(t *testing.T) {
	svc, _, _, _, _ := newTestSvc()
	resp := mustActivity(t, svc, uuid.New(), 10)
	activityID, _ := uuid.Parse(resp.ID)
	user := uuid.New()

	if _, err := svc.Register(context.Background(), activityID, user); err != nil {
		t.Fatal(err)
	}
	_, err := svc.Register(context.Background(), activityID, user)
	if err == nil {
		t.Fatal("expected duplicate error")
	}
	if appErr, ok := err.(*response.AppError); ok && appErr.Code != response.CodeRegistrationExists {
		t.Errorf("code = %d, want %d", appErr.Code, response.CodeRegistrationExists)
	}
}

// 活动状态不对不能报名
func TestRegister_WrongStatus(t *testing.T) {
	svc, _, _, _, _ := newTestSvc()
	resp := mustActivity(t, svc, uuid.New(), 10)
	activityID, _ := uuid.Parse(resp.ID)

	if _, err := svc.StartActivity(context.Background(), activityID); err != nil {
		t.Fatal(err)
	}
	_, err := svc.Register(context.Background(), activityID, uuid.New())
	if err == nil {
		t.Fatal("expected error registering ongoing activity")
	}
	if appErr, ok := err.(*response.AppError); ok && appErr.Code != response.CodeActivityInvalidState {
		t.Errorf("code = %d, want %d", appErr.Code, response.CodeActivityInvalidState)
	}
}

// 取消报名后，候补自动递补
func TestCancelRegistration_AutoPromote(t *testing.T) {
	svc, _, _, _, nn := newTestSvc()
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

	// user1 取消，user2 应自动递补
	if err := svc.CancelRegistration(context.Background(), activityID, user1); err != nil {
		t.Fatalf("CancelRegistration: %v", err)
	}

	reg2, err := svc.regs.GetByActivityAndUser(context.Background(), activityID, user2)
	if err != nil {
		t.Fatal(err)
	}
	if reg2 == nil || reg2.Status != model.RegApproved {
		t.Errorf("user2 status = %v, want approved", reg2)
	}

	// 应发送 approved 通知给 user2
	found := false
	for i, tpl := range nn.sent {
		if tpl == tplActivityApproved {
			for _, u := range nn.targets[i] {
				if u == user2 {
					found = true
				}
			}
		}
	}
	if !found {
		t.Error("user2 should receive approved notification")
	}
}

// 管理员审批通过
func TestApproveRegistration(t *testing.T) {
	svc, _, _, _, _ := newTestSvc()
	resp := mustActivity(t, svc, uuid.New(), 10)
	activityID, _ := uuid.Parse(resp.ID)
	user := uuid.New()

	// 用户报名（自动通过）
	if _, err := svc.Register(context.Background(), activityID, user); err != nil {
		t.Fatal(err)
	}

	// 管理员拒绝
	regRej, err := svc.ApproveRegistration(context.Background(), activityID, user, false, "不合适")
	if err != nil {
		t.Fatalf("reject: %v", err)
	}
	if regRej.Status != int16(model.RegRejected) {
		t.Errorf("status = %d, want rejected", regRej.Status)
	}

	// 管理员再次通过
	regApp, err := svc.ApproveRegistration(context.Background(), activityID, user, true, "")
	if err != nil {
		t.Fatalf("approve: %v", err)
	}
	if regApp.Status != int16(model.RegApproved) {
		t.Errorf("status = %d, want approved", regApp.Status)
	}
}

// 审批时名额已满
func TestApproveRegistration_Full(t *testing.T) {
	svc, _, _, _, _ := newTestSvc()
	resp := mustActivity(t, svc, uuid.New(), 1)
	activityID, _ := uuid.Parse(resp.ID)
	user1 := uuid.New()
	user2 := uuid.New()

	if _, err := svc.Register(context.Background(), activityID, user1); err != nil {
		t.Fatal(err)
	}
	// user2 在候补
	if _, err := svc.Register(context.Background(), activityID, user2); err != nil {
		t.Fatal(err)
	}
	// 强制把 user2 改为待审批，然后尝试通过（名额仍满）
	reg2, _ := svc.regs.GetByActivityAndUser(context.Background(), activityID, user2)
	reg2.Status = model.RegPending
	_ = svc.regs.Update(context.Background(), reg2)

	_, err := svc.ApproveRegistration(context.Background(), activityID, user2, true, "")
	if err == nil {
		t.Fatal("expected full error")
	}
	if appErr, ok := err.(*response.AppError); ok && appErr.Code != response.CodeActivityFull {
		t.Errorf("code = %d, want %d", appErr.Code, response.CodeActivityFull)
	}
}

func TestListRegistrations(t *testing.T) {
	svc, _, _, _, _ := newTestSvc()
	resp := mustActivity(t, svc, uuid.New(), 10)
	activityID, _ := uuid.Parse(resp.ID)

	for i := 0; i < 3; i++ {
		if _, err := svc.Register(context.Background(), activityID, uuid.New()); err != nil {
			t.Fatal(err)
		}
	}
	list, err := svc.ListRegistrations(context.Background(), activityID)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 3 {
		t.Errorf("list len = %d, want 3", len(list))
	}
}
