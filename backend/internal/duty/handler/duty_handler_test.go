package handler

import (
	"testing"

	"github.com/Yogdunana/StarByte/backend/internal/duty/dto"
	"github.com/Yogdunana/StarByte/backend/internal/duty/model"
)

func TestGetScheduleStatusText(t *testing.T) {
	tests := []struct {
		status int
		want   string
	}{
		{0, "待值班"},
		{1, "已到岗"},
		{2, "已完成"},
		{3, "缺勤"},
		{4, "调班"},
		{99, "未知"},
	}
	for _, tt := range tests {
		got := dto.GetScheduleStatusText(tt.status)
		if got != tt.want {
			t.Errorf("GetScheduleStatusText(%d) = %q, want %q", tt.status, got, tt.want)
		}
	}
}

func TestGetSwapStatusText(t *testing.T) {
	tests := []struct {
		status int
		want   string
	}{
		{0, "待审批"},
		{1, "已批准"},
		{2, "已拒绝"},
		{3, "已取消"},
		{99, "未知"},
	}
	for _, tt := range tests {
		got := dto.GetSwapStatusText(tt.status)
		if got != tt.want {
			t.Errorf("GetSwapStatusText(%d) = %q, want %q", tt.status, got, tt.want)
		}
	}
}

func TestScheduleStatusConstants(t *testing.T) {
	if model.ScheduleStatusPending != 0 {
		t.Error("expected ScheduleStatusPending = 0")
	}
	if model.ScheduleStatusOnDuty != 1 {
		t.Error("expected ScheduleStatusOnDuty = 1")
	}
	if model.ScheduleStatusCompleted != 2 {
		t.Error("expected ScheduleStatusCompleted = 2")
	}
	if model.ScheduleStatusAbsent != 3 {
		t.Error("expected ScheduleStatusAbsent = 3")
	}
	if model.ScheduleStatusSwapped != 4 {
		t.Error("expected ScheduleStatusSwapped = 4")
	}
}

func TestSwapStatusConstants(t *testing.T) {
	if model.SwapStatusPending != 0 {
		t.Error("expected SwapStatusPending = 0")
	}
	if model.SwapStatusApproved != 1 {
		t.Error("expected SwapStatusApproved = 1")
	}
	if model.SwapStatusRejected != 2 {
		t.Error("expected SwapStatusRejected = 2")
	}
	if model.SwapStatusCanceled != 3 {
		t.Error("expected SwapStatusCanceled = 3")
	}
}

func TestTimeSlotConstants(t *testing.T) {
	if model.TimeSlotMorning != "morning" {
		t.Error("expected TimeSlotMorning = morning")
	}
	if model.TimeSlotAfternoon != "afternoon" {
		t.Error("expected TimeSlotAfternoon = afternoon")
	}
	if model.TimeSlotEvening != "evening" {
		t.Error("expected TimeSlotEvening = evening")
	}
	if model.TimeSlotFullDay != "full_day" {
		t.Error("expected TimeSlotFullDay = full_day")
	}
}

func TestTableNames(t *testing.T) {
	if (model.Schedule{}).TableName() != "duty_schedules" {
		t.Error("wrong schedule table name")
	}
	if (model.SwapRequest{}).TableName() != "duty_swap_requests" {
		t.Error("wrong swap_request table name")
	}
}

func TestNewDutyHandler(t *testing.T) {
	h := NewDutyHandler(nil)
	if h == nil {
		t.Error("expected non-nil handler")
	}
}
