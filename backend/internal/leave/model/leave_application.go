package model

import (
	"time"

	"gorm.io/gorm"
)

// 审批状态常量
const (
	ApprovalStatusPending  = "pending"  // 待审批
	ApprovalStatusApproved = "approved" // 已批准
	ApprovalStatusRejected = "rejected" // 已驳回
)

// LeaveApplication 请假申请模型
// 对应数据库表 leave_applications，记录每条请假申请的详细信息
type LeaveApplication struct {
	gorm.Model
	ApplicantID   uint       `gorm:"column:applicant_id;not null;index;comment:申请人用户ID" json:"applicant_id"`  // 申请人用户ID
	LeaveTypeID   uint       `gorm:"column:leave_type_id;not null;comment:请假类型ID" json:"leave_type_id"`       // 请假类型ID
	StartTime     time.Time  `gorm:"column:start_time;not null;comment:请假开始时间" json:"start_time"`            // 请假开始时间
	EndTime       time.Time  `gorm:"column:end_time;not null;comment:请假结束时间" json:"end_time"`               // 请假结束时间
	DurationDays  float64    `gorm:"column:duration_days;not null;comment:请假时长（天）" json:"duration_days"`     // 请假时长（天）
	Reason        string     `gorm:"column:reason;size:500;comment:请假事由" json:"reason"`                     // 请假事由
	Status        string     `gorm:"column:status;size:20;default:pending;index;comment:审批状态" json:"status"` // 审批状态（pending/approved/rejected）
	ApproverID    uint       `gorm:"column:approver_id;comment:审批人用户ID" json:"approver_id"`                  // 审批人用户ID
	ApproveRemark string     `gorm:"column:approve_remark;size:500;comment:审批备注" json:"approve_remark"`        // 审批备注
	ApprovedAt    *time.Time `gorm:"column:approved_at;comment:审批时间" json:"approved_at"`                       // 审批时间

	// 关联请假类型
	LeaveType LeaveType `gorm:"foreignKey:LeaveTypeID" json:"leave_type,omitempty"`
}

// TableName 指定数据库表名
func (LeaveApplication) TableName() string {
	return "leave_applications"
}
