package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LeaveType 请假类型（事假/病假/年假/调休）。
// 字段语义来自 SMB-Star PR #162；default_days 仅用于余额自动开户。
type LeaveType struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Name        string         `gorm:"type:varchar(50);not null" json:"name"`
	Code        string         `gorm:"type:varchar(20);not null;uniqueIndex" json:"code"`
	Deductible  bool           `gorm:"not null;default:true" json:"deductible"`
	DefaultDays float64        `gorm:"type:numeric(6,1);not null;default:0" json:"default_days"`
	Description string         `gorm:"type:varchar(255)" json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (LeaveType) TableName() string { return "leave_types" }

// LeaveBalance 按用户+年份+类型保存额度。算术字段与 #162 一致。
type LeaveBalance struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	UserID        uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:idx_leave_bal_user_year_type" json:"user_id"`
	Year          int            `gorm:"not null;uniqueIndex:idx_leave_bal_user_year_type" json:"year"`
	LeaveTypeID   uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:idx_leave_bal_user_year_type" json:"leave_type_id"`
	TotalDays     float64        `gorm:"type:numeric(6,1);not null;default:0" json:"total_days"`
	UsedDays      float64        `gorm:"type:numeric(6,1);not null;default:0" json:"used_days"`
	RemainingDays float64        `gorm:"type:numeric(6,1);not null;default:0" json:"remaining_days"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
	LeaveType     LeaveType      `gorm:"foreignKey:LeaveTypeID" json:"leave_type,omitempty"`
}

func (LeaveBalance) TableName() string { return "leave_balances" }

// LeaveApplication 请假申请。提交即扣余额、驳回返还，与 #162 一致。
type LeaveApplication struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	ApplicantID   uuid.UUID      `gorm:"type:uuid;not null;index" json:"applicant_id"`
	LeaveTypeID   uuid.UUID      `gorm:"type:uuid;not null" json:"leave_type_id"`
	StartTime     time.Time      `gorm:"not null" json:"start_time"`
	EndTime       time.Time      `gorm:"not null" json:"end_time"`
	DurationDays  float64        `gorm:"type:numeric(6,1);not null" json:"duration_days"`
	Reason        string         `gorm:"type:varchar(500)" json:"reason"`
	Status        string         `gorm:"type:varchar(20);not null;default:pending;index" json:"status"`
	ApproverID    *uuid.UUID     `gorm:"type:uuid" json:"approver_id,omitempty"`
	ApproveRemark string         `gorm:"type:varchar(500)" json:"approve_remark"`
	ApprovedAt    *time.Time     `json:"approved_at"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
	LeaveType     LeaveType      `gorm:"foreignKey:LeaveTypeID" json:"leave_type,omitempty"`
}

func (LeaveApplication) TableName() string { return "leave_applications" }

// ApplicationNamed 列表/详情附带申请人与审批人姓名。
type ApplicationNamed struct {
	LeaveApplication
	ApplicantName         string     `gorm:"column:applicant_name" json:"applicant_name"`
	ApproverName          string     `gorm:"column:approver_name" json:"approver_name"`
	ApplicantDepartmentID *uuid.UUID `gorm:"column:applicant_department_id" json:"-"`
}
