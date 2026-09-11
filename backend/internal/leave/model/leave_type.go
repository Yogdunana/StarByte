package model

import "gorm.io/gorm"

// LeaveType 请假类型模型
// 对应数据库表 leave_types，存储事假/病假/年假等请假类型定义
type LeaveType struct {
	gorm.Model
	Name        string `gorm:"column:name;size:50;not null;comment:请假类型名称" json:"name"`         // 请假类型名称（事假/病假/年假等）
	Code        string `gorm:"column:code;size:20;not null;uniqueIndex;comment:请假类型编码" json:"code"` // 请假类型编码（唯一）
	Deductible  bool   `gorm:"column:deductible;default:true;comment:是否扣减余额" json:"deductible"`     // 是否扣减假期余额
	Description string `gorm:"column:description;size:255;comment:类型描述说明" json:"description"`      // 类型描述说明
}

// TableName 指定数据库表名
func (LeaveType) TableName() string {
	return "leave_types"
}
