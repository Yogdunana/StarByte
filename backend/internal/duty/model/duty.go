package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Schedule 排班表模型
type Schedule struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	UserID        uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	DepartmentID  *uuid.UUID     `gorm:"type:uuid;index" json:"department_id"`
	DutyDate      time.Time      `gorm:"type:date;not null;index" json:"duty_date"`
	TimeSlot      string         `gorm:"type:varchar(20);not null;default:full_day" json:"time_slot"` // morning/afternoon/evening/full_day
	Location      string         `gorm:"type:varchar(100)" json:"location"`
	Remark        string         `gorm:"type:varchar(500)" json:"remark"`
	Status        int            `gorm:"type:smallint;not null;default:0;index" json:"status"` // 0=待值班 1=已到岗 2=已完成 3=缺勤 4=调班
	CreatedBy     *uuid.UUID     `gorm:"type:uuid" json:"created_by"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Schedule) TableName() string {
	return "duty_schedules"
}

// SwapRequest 调班申请模型
type SwapRequest struct {
	ID                   uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	RequesterID          uuid.UUID      `gorm:"type:uuid;not null;index" json:"requester_id"`
	TargetUserID         uuid.UUID      `gorm:"type:uuid;not null;index" json:"target_user_id"`
	RequesterScheduleID  uuid.UUID      `gorm:"type:uuid;not null" json:"requester_schedule_id"`
	TargetScheduleID     *uuid.UUID     `gorm:"type:uuid" json:"target_schedule_id"`
	Reason               string         `gorm:"type:varchar(500);not null" json:"reason"`
	Status               int            `gorm:"type:smallint;not null;default:0;index" json:"status"` // 0=待审批 1=已批准 2=已拒绝 3=已取消
	ApproverID           *uuid.UUID     `gorm:"type:uuid" json:"approver_id"`
	ApprovedAt           *time.Time     `json:"approved_at"`
	ApprovedRemark       string         `gorm:"type:varchar(500)" json:"approved_remark"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
	DeletedAt            gorm.DeletedAt `gorm:"index" json:"-"`
}

func (SwapRequest) TableName() string {
	return "duty_swap_requests"
}

// 排班状态常量
const (
	ScheduleStatusPending   = 0 // 待值班
	ScheduleStatusOnDuty    = 1 // 已到岗
	ScheduleStatusCompleted = 2 // 已完成
	ScheduleStatusAbsent    = 3 // 缺勤
	ScheduleStatusSwapped   = 4 // 调班
)

// 调班申请状态常量
const (
	SwapStatusPending  = 0 // 待审批
	SwapStatusApproved = 1 // 已批准
	SwapStatusRejected = 2 // 已拒绝
	SwapStatusCanceled = 3 // 已取消
)

// 时段常量
const (
	TimeSlotMorning   = "morning"
	TimeSlotAfternoon = "afternoon"
	TimeSlotEvening   = "evening"
	TimeSlotFullDay   = "full_day"
)
