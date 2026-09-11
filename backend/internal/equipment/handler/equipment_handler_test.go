package handler

import (
	"testing"

	"github.com/Yogdunana/StarByte/backend/internal/equipment/dto"
	"github.com/Yogdunana/StarByte/backend/internal/equipment/model"
)

func TestGetEquipmentStatusText(t *testing.T) {
	for _, tt := range []struct {
		status int
		want   string
	}{
		{0, "可用"}, {1, "部分借用"}, {2, "全部借出"}, {3, "维修中"}, {4, "已报废"}, {99, "未知"},
	} {
		if got := dto.GetEquipmentStatusText(tt.status); got != tt.want {
			t.Errorf("GetEquipmentStatusText(%d) = %q, want %q", tt.status, got, tt.want)
		}
	}
}

func TestGetBorrowStatusText(t *testing.T) {
	for _, tt := range []struct {
		status int
		want   string
	}{
		{0, "待审批"}, {1, "已批准"}, {2, "已领取"}, {3, "已归还"}, {4, "已拒绝"}, {5, "逾期"}, {99, "未知"},
	} {
		if got := dto.GetBorrowStatusText(tt.status); got != tt.want {
			t.Errorf("GetBorrowStatusText(%d) = %q, want %q", tt.status, got, tt.want)
		}
	}
}

func TestGetCategoryText(t *testing.T) {
	for _, tt := range []struct {
		cat  string
		want string
	}{
		{"general", "通用"}, {"electronic", "电子设备"}, {"book", "书籍"}, {"tool", "工具"}, {"other", "其他"}, {"unknown", "其他"},
	} {
		if got := dto.GetCategoryText(tt.cat); got != tt.want {
			t.Errorf("GetCategoryText(%q) = %q, want %q", tt.cat, got, tt.want)
		}
	}
}

func TestEquipmentModelConstants(t *testing.T) {
	if model.EquipmentStatusAvailable != 0 {
		t.Error("expected EquipmentStatusAvailable = 0")
	}
	if model.BorrowStatusPending != 0 {
		t.Error("expected BorrowStatusPending = 0")
	}
	if model.MaintenanceTypeRepair != 0 {
		t.Error("expected MaintenanceTypeRepair = 0")
	}
}

func TestEquipmentTableNames(t *testing.T) {
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

func TestNewEquipmentHandler(t *testing.T) {
	h := NewEquipmentHandler(nil)
	if h == nil {
		t.Error("expected non-nil handler")
	}
}
