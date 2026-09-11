package service

import (
	"context"
	"testing"

	"github.com/Yogdunana/StarByte/backend/internal/activity/dto"
	"github.com/Yogdunana/StarByte/backend/internal/activity/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
)

func TestRegister_AutoApprove(t *testing.T) {
	svc, _, rr, _, nn := newTestSvc()
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
	if rr.createCount != 1 {
		t.Errorf("createCount = %d, want 1", rr.createCount)
	}
	if rr.updateCount != 0 {
		t.Errorf("first-time register must Create, updateCount = %d", rr.updateCount)
	}
}

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

func TestRegister_CancelledThenCreateNotUsed(t *testing.T) {
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
	creates := rr.createCount
	if _, err := svc.Register(context.Background(), activityID, user); err != nil {
		t.Fatalf("re-register: %v", err)
	}
	if rr.createCount != creates {
		t.Errorf("re-register after cancel should Update, createCount %d -> %d", creates, rr.createCount)
	}
	if rr.updateCount < 2 { // cancel + re-register
		t.Errorf("updateCount = %d, want >= 2", rr.updateCount)
	}
	mine, err := svc.GetMyRegistration(context.Background(), activityID, user)
	if err != nil || mine == nil {
		t.Fatalf("GetMyRegistration: %v %v", mine, err)
	}
	if mine.Status != model.RegApproved {
		t.Errorf("status = %d, want approved", mine.Status)
	}
}

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

func TestCancelRegistration_NoPromoteWhenEnded(t *testing.T) {
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
	if _, err := svc.StartActivity(context.Background(), activityID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.EndActivity(context.Background(), activityID); err != nil {
		t.Fatal(err)
	}

	if err := svc.CancelRegistration(context.Background(), activityID, user1); err != nil {
		t.Fatalf("CancelRegistration: %v", err)
	}
	reg2, err := svc.regs.GetByActivityAndUser(context.Background(), activityID, user2)
	if err != nil {
		t.Fatal(err)
	}
	if reg2 == nil || reg2.Status != model.RegWaitlist {
		t.Errorf("ended activity should not promote waitlist, got %+v", reg2)
	}
	for _, tpl := range nn.sent {
		if tpl == tplActivityApproved {
			t.Error("should not notify waitlist after activity ended")
		}
	}
}

func TestCancelRegistration_NoPromoteWhenOverCapacity(t *testing.T) {
	svc, _, _, _, _ := newTestSvc()
	resp := mustActivity(t, svc, uuid.New(), 2)
	activityID, _ := uuid.Parse(resp.ID)
	user1 := uuid.New()
	user2 := uuid.New()
	user3 := uuid.New()

	if _, err := svc.Register(context.Background(), activityID, user1); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Register(context.Background(), activityID, user2); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Register(context.Background(), activityID, user3); err != nil {
		t.Fatal(err)
	}
	newMax := 1
	if _, err := svc.UpdateActivity(context.Background(), activityID, &dto.UpdateActivityRequest{
		MaxParticipants: &newMax,
	}); err != nil {
		t.Fatal(err)
	}

	if err := svc.CancelRegistration(context.Background(), activityID, user1); err != nil {
		t.Fatalf("CancelRegistration: %v", err)
	}
	reg3, err := svc.regs.GetByActivityAndUser(context.Background(), activityID, user3)
	if err != nil {
		t.Fatal(err)
	}
	if reg3 == nil || reg3.Status != model.RegWaitlist {
		t.Errorf("over-capacity cancel should not promote, got %+v", reg3)
	}
}

func TestApproveRegistration_RejectPromotesWaitlist(t *testing.T) {
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

	if _, err := svc.ApproveRegistration(context.Background(), activityID, user1, false, "不合适"); err != nil {
		t.Fatalf("reject: %v", err)
	}
	reg2, err := svc.regs.GetByActivityAndUser(context.Background(), activityID, user2)
	if err != nil {
		t.Fatal(err)
	}
	if reg2 == nil || reg2.Status != model.RegApproved {
		t.Errorf("rejecting approved should promote waitlist, got %+v", reg2)
	}

	foundReject, foundPromote := false, false
	for i, tpl := range nn.sent {
		if tpl == tplActivityRejected {
			for _, u := range nn.targets[i] {
				if u == user1 {
					foundReject = true
				}
			}
		}
		if tpl == tplActivityApproved {
			for _, u := range nn.targets[i] {
				if u == user2 {
					foundPromote = true
				}
			}
		}
	}
	if !foundReject {
		t.Error("rejected user should be notified")
	}
	if !foundPromote {
		t.Error("promoted waitlist user should be notified")
	}
}

func TestApproveRegistration(t *testing.T) {
	svc, _, _, _, _ := newTestSvc()
	resp := mustActivity(t, svc, uuid.New(), 10)
	activityID, _ := uuid.Parse(resp.ID)
	user := uuid.New()

	if _, err := svc.Register(context.Background(), activityID, user); err != nil {
		t.Fatal(err)
	}

	regRej, err := svc.ApproveRegistration(context.Background(), activityID, user, false, "不合适")
	if err != nil {
		t.Fatalf("reject: %v", err)
	}
	if regRej.Status != int16(model.RegRejected) {
		t.Errorf("status = %d, want rejected", regRej.Status)
	}

	regApp, err := svc.ApproveRegistration(context.Background(), activityID, user, true, "")
	if err != nil {
		t.Fatalf("approve: %v", err)
	}
	if regApp.Status != int16(model.RegApproved) {
		t.Errorf("status = %d, want approved", regApp.Status)
	}
}

func TestApproveRegistration_Full(t *testing.T) {
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

func TestGetMyRegistration_None(t *testing.T) {
	svc, _, _, _, _ := newTestSvc()
	resp := mustActivity(t, svc, uuid.New(), 10)
	activityID, _ := uuid.Parse(resp.ID)
	mine, err := svc.GetMyRegistration(context.Background(), activityID, uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	if mine != nil {
		t.Errorf("expected nil registration, got %+v", mine)
	}
}
