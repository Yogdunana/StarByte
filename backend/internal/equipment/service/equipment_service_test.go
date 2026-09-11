package service

import (
	"testing"

	"github.com/Yogdunana/StarByte/backend/internal/equipment/dto"
	"github.com/Yogdunana/StarByte/backend/internal/equipment/model"
)

func TestGetEquipmentStatusText(t *testing.T) {
	tests := []struct {
		status int
		want   string
	}{
		{0, "可用"},
		{1, "部分借用"},
		{2, "全部借出"},
		{3, "维修中"},
		{4, "已报废"},
		{99, "未知"},
	}
	for _, tt := range tests {
		got := dto.GetEquipmentStatusText(tt.status)
		if got != tt.want {
			t.Errorf("GetEquipmentStatusText(%d) = %q, want %q", tt.status, got, tt.want)
		}
	}
}

func TestGetBorrowStatusText(t *testing.T) {
	tests := []struct {
		status int
		want   string
	}{
		{0, "待审批"},
		{1, "已批准"},
		{2, "已领取"},
		{3, "已归还"},
		{4, "已拒绝"},
		{5, "逾期"},
		{99, "未知"},
	}
	for _, tt := range tests {
		got := dto.GetBorrowStatusText(tt.status)
		if got != tt.want {
			t.Errorf("GetBorrowStatusText(%d) = %q, want %q", tt.status, got, tt.want)
		}
	}
}

func TestGetMaintenanceTypeText(t *testing.T) {
	tests := []struct {
		val  int
		want string
	}{
		{0, "维修"},
		{1, "保养"},
		{2, "报废"},
		{99, "未知"},
	}
	for _, tt := range tests {
		got := dto.GetMaintenanceTypeText(tt.val)
		if got != tt.want {
			t.Errorf("GetMaintenanceTypeText(%d) = %q, want %q", tt.val, got, tt.want)
		}
	}
}

func TestGetMaintenanceStatusText(t *testing.T) {
	tests := []struct {
		val  int
		want string
	}{
		{0, "进行中"},
		{1, "已完成"},
		{2, "已取消"},
		{99, "未知"},
	}
	for _, tt := range tests {
		got := dto.GetMaintenanceStatusText(tt.val)
		if got != tt.want {
			t.Errorf("GetMaintenanceStatusText(%d) = %q, want %q", tt.val, got, tt.want)
		}
	}
}

func TestGetCategoryText(t *testing.T) {
	tests := []struct {
		cat  string
		want string
	}{
		{"general", "通用"},
		{"electronic", "电子设备"},
		{"book", "书籍"},
		{"tool", "工具"},
		{"other", "其他"},
		{"unknown", "其他"},
	}
	for _, tt := range tests {
		got := dto.GetCategoryText(tt.cat)
		if got != tt.want {
			t.Errorf("GetCategoryText(%q) = %q, want %q", tt.cat, got, tt.want)
		}
	}
}

func TestFormatTimePtr(t *testing.T) {
	if dto.FormatTimePtr(nil) != "" {
		t.Error("expected empty string for nil time")
	}
}

func TestFormatFloatPtr(t *testing.T) {
	if dto.FormatFloatPtr(nil) != 0 {
		t.Error("expected 0 for nil float")
	}
	val := 99.5
	if dto.FormatFloatPtr(&val) != 99.5 {
		t.Error("expected 99.5 for non-nil float")
	}
}

func TestFormatDatePtr(t *testing.T) {
	if dto.FormatDatePtr(nil) != "" {
		t.Error("expected empty string for nil date")
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

func TestEquipmentService_New(t *testing.T) {
	svc := NewEquipmentService(nil, nil, nil, nil, nil, nil)
	if svc == nil {
		t.Error("expected non-nil service")
	}
}

func TestEquipmentStatusConstants(t *testing.T) {
	if model.EquipmentStatusAvailable != 0 {
		t.Error("expected EquipmentStatusAvailable = 0")
	}
	if model.EquipmentStatusPartBorrowed != 1 {
		t.Error("expected EquipmentStatusPartBorrowed = 1")
	}
	if model.EquipmentStatusAllBorrowed != 2 {
		t.Error("expected EquipmentStatusAllBorrowed = 2")
	}
	if model.EquipmentStatusRepairing != 3 {
		t.Error("expected EquipmentStatusRepairing = 3")
	}
	if model.EquipmentStatusScrapped != 4 {
		t.Error("expected EquipmentStatusScrapped = 4")
	}
}

func TestBorrowStatusConstants(t *testing.T) {
	if model.BorrowStatusPending != 0 {
		t.Error("expected BorrowStatusPending = 0")
	}
	if model.BorrowStatusApproved != 1 {
		t.Error("expected BorrowStatusApproved = 1")
	}
	if model.BorrowStatusTaken != 2 {
		t.Error("expected BorrowStatusTaken = 2")
	}
	if model.BorrowStatusReturned != 3 {
		t.Error("expected BorrowStatusReturned = 3")
	}
	if model.BorrowStatusRejected != 4 {
		t.Error("expected BorrowStatusRejected = 4")
	}
	if model.BorrowStatusOverdue != 5 {
		t.Error("expected BorrowStatusOverdue = 5")
	}
}

func TestMaintenanceConstants(t *testing.T) {
	if model.MaintenanceTypeRepair != 0 {
		t.Error("expected MaintenanceTypeRepair = 0")
	}
	if model.MaintenanceTypeService != 1 {
		t.Error("expected MaintenanceTypeService = 1")
	}
	if model.MaintenanceTypeScrap != 2 {
		t.Error("expected MaintenanceTypeScrap = 2")
	}
	if model.MaintenanceStatusInProgress != 0 {
		t.Error("expected MaintenanceStatusInProgress = 0")
	}
	if model.MaintenanceStatusCompleted != 1 {
		t.Error("expected MaintenanceStatusCompleted = 1")
	}
	if model.MaintenanceStatusCanceled != 2 {
		t.Error("expected MaintenanceStatusCanceled = 2")
	}
}

func TestTableNames(t *testing.T) {
	if (model.Equipment{}).TableName() != "equipment_items" {
		t.Error("wrong equipment table name")
	}
	if (model.Borrow{}).TableName() != "equipment_borrows" {
		t.Error("wrong borrow table name")
	}
	if (model.Maintenance{}).TableName() != "equipment_maintenance" {
		t.Error("wrong maintenance table name")
	}
	if (model.Inventory{}).TableName() != "equipment_inventories" {
		t.Error("wrong inventory table name")
	}
}
