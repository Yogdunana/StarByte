package model

import "gorm.io/gorm"

// LeaveBalance 用户假期余额模型
// 对应数据库表 leave_balances，按用户、年份、请假类型保存剩余天数
type LeaveBalance struct {
	gorm.Model
	UserID        uint    `gorm:"column:user_id;not null;index:idx_user_year_type,unique;comment:用户ID" json:"user_id"`          // 用户ID
	Year          int     `gorm:"column:year;not null;index:idx_user_year_type,unique;comment:年份" json:"year"`                  // 年份
	LeaveTypeID   uint    `gorm:"column:leave_type_id;not null;index:idx_user_year_type,unique;comment:请假类型ID" json:"leave_type_id"` // 请假类型ID
	TotalDays     float64 `gorm:"column:total_days;default:0;comment:总天数" json:"total_days"`                              // 总天数
	UsedDays      float64 `gorm:"column:used_days;default:0;comment:已用天数" json:"used_days"`                               // 已用天数
	RemainingDays float64 `gorm:"column:remaining_days;default:0;comment:剩余天数" json:"remaining_days"`                        // 剩余天数

	// 关联请假类型
	LeaveType LeaveType `gorm:"foreignKey:LeaveTypeID" json:"leave_type,omitempty"`
}

// TableName 指定数据库表名
func (LeaveBalance) TableName() string {
	return "leave_balances"
}
