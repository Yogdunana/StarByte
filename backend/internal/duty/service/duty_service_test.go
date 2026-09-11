package service

import (
	"testing"

	"github.com/Yogdunana/StarByte/backend/internal/duty/dto"
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

func TestParseRequiredUUID_Invalid(t *testing.T) {
	_, err := parseRequiredUUID("invalid", "测试ID")
	if err == nil {
		t.Error("expected error for invalid UUID")
	}
}

func TestParseRequiredUUID_Valid(t *testing.T) {
	_, err := parseRequiredUUID("550e8400-e29b-41d4-a716-446655440000", "测试ID")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestParseOptionalUUID_Empty(t *testing.T) {
	result, err := parseOptionalUUID("", "测试ID")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != nil {
		t.Error("expected nil for empty string")
	}
}

func TestParseOptionalUUID_Valid(t *testing.T) {
	result, err := parseOptionalUUID("550e8400-e29b-41d4-a716-446655440000", "测试ID")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result == nil {
		t.Error("expected non-nil for valid UUID")
	}
}

func TestParseOptionalUUID_Invalid(t *testing.T) {
	_, err := parseOptionalUUID("not-a-uuid", "测试ID")
	if err == nil {
		t.Error("expected error for invalid UUID")
	}
}

func TestDutyService_NewDutyService(t *testing.T) {
	svc := NewDutyService(nil, nil, nil, nil)
	if svc == nil {
		t.Error("expected non-nil service")
	}
}
