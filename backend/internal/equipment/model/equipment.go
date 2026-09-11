package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Equipment 物资/设备
type Equipment struct {
	ID                uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Name              string         `gorm:"type:varchar(100);not null;index" json:"name"`
	Model             string         `gorm:"type:varchar(100)" json:"model"`
	Category          string         `gorm:"type:varchar(50);not null;default:general;index" json:"category"`
	TotalQuantity     int            `gorm:"not null;default:1" json:"total_quantity"`
	AvailableQuantity int            `gorm:"not null;default:1" json:"available_quantity"`
	Location          string         `gorm:"type:varchar(200)" json:"location"`
	ManagerID         *uuid.UUID     `gorm:"type:uuid;index" json:"manager_id"`
	PhotoFileID       *uuid.UUID     `gorm:"type:uuid" json:"photo_file_id"`
	Status            int            `gorm:"type:smallint;not null;default:0;index" json:"status"`
	QRCode            string         `gorm:"type:varchar(500)" json:"qr_code"`
	Description       string         `gorm:"type:text" json:"description"`
	PurchaseDate      *time.Time     `gorm:"type:date" json:"purchase_date"`
	PurchasePrice     *float64       `gorm:"type:numeric(12,2)" json:"purchase_price"`
	CreatedBy         *uuid.UUID     `gorm:"type:uuid" json:"created_by"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Equipment) TableName() string {
	return "equipment_items"
}

// Borrow 借用记录
type Borrow struct {
	ID               uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	EquipmentID      uuid.UUID      `gorm:"type:uuid;not null;index" json:"equipment_id"`
	BorrowerID       uuid.UUID      `gorm:"type:uuid;not null;index" json:"borrower_id"`
	Quantity         int            `gorm:"not null;default:1" json:"quantity"`
	BorrowAt         time.Time      `gorm:"not null" json:"borrow_at"`
	ExpectedReturnAt time.Time      `gorm:"not null" json:"expected_return_at"`
	ActualReturnAt   *time.Time     `json:"actual_return_at"`
	Status           int            `gorm:"type:smallint;not null;default:0;index" json:"status"`
	ApproverID       *uuid.UUID     `gorm:"type:uuid" json:"approver_id"`
	ApprovedAt       *time.Time     `json:"approved_at"`
	ApprovedRemark   string         `gorm:"type:varchar(500)" json:"approved_remark"`
	ReturnRemark     string         `gorm:"type:varchar(500)" json:"return_remark"`
	ReturnCheckerID  *uuid.UUID     `gorm:"type:uuid" json:"return_checker_id"`
	Remark           string         `gorm:"type:varchar(500)" json:"remark"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Borrow) TableName() string {
	return "equipment_borrows"
}

// Maintenance 维修记录
type Maintenance struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	EquipmentID  uuid.UUID      `gorm:"type:uuid;not null;index" json:"equipment_id"`
	Type         int            `gorm:"type:smallint;not null;default:0" json:"type"` // 0=维修 1=保养 2=报废
	Description string         `gorm:"type:varchar(500);not null" json:"description"`
	Cost         *float64       `gorm:"type:numeric(12,2)" json:"cost"`
	StartDate    time.Time      `gorm:"type:date;not null" json:"start_date"`
	EndDate      *time.Time     `gorm:"type:date" json:"end_date"`
	Status       int            `gorm:"type:smallint;not null;default:0;index" json:"status"` // 0=进行中 1=已完成 2=已取消
	OperatorID   *uuid.UUID     `gorm:"type:uuid" json:"operator_id"`
	Remark       string         `gorm:"type:varchar(500)" json:"remark"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Maintenance) TableName() string {
	return "equipment_maintenance"
}

// Inventory 库存盘点
type Inventory struct {
	ID               uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	EquipmentID      uuid.UUID      `gorm:"type:uuid;not null;index" json:"equipment_id"`
	ExpectedQuantity int            `gorm:"not null" json:"expected_quantity"`
	ActualQuantity   int            `gorm:"not null" json:"actual_quantity"`
	Difference       int            `gorm:"not null;default:0" json:"difference"`
	CheckerID        *uuid.UUID     `gorm:"type:uuid" json:"checker_id"`
	CheckDate        time.Time      `gorm:"type:date;not null;index" json:"check_date"`
	Remark           string         `gorm:"type:varchar(500)" json:"remark"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Inventory) TableName() string {
	return "equipment_inventories"
}

// 物资状态常量
const (
	EquipmentStatusAvailable    = 0 // 可用
	EquipmentStatusPartBorrowed = 1 // 部分借用
	EquipmentStatusAllBorrowed  = 2 // 全部借出
	EquipmentStatusRepairing    = 3 // 维修中
	EquipmentStatusScrapped     = 4 // 已报废
)

// 借用状态常量
const (
	BorrowStatusPending  = 0 // 待审批
	BorrowStatusApproved = 1 // 已批准
	BorrowStatusTaken    = 2 // 已领取
	BorrowStatusReturned = 3 // 已归还
	BorrowStatusRejected = 4 // 已拒绝
	BorrowStatusOverdue  = 5 // 逾期
)

// 维修类型常量
const (
	MaintenanceTypeRepair  = 0 // 维修
	MaintenanceTypeService = 1 // 保养
	MaintenanceTypeScrap   = 2 // 报废
)

// 维修状态常量
const (
	MaintenanceStatusInProgress = 0 // 进行中
	MaintenanceStatusCompleted  = 1 // 已完成
	MaintenanceStatusCanceled   = 2 // 已取消
)
