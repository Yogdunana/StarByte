package dto

import "time"

// ========== 请求 DTO ==========

// ListScheduleRequest 排班列表查询
type ListScheduleRequest struct {
	Page         int    `form:"page,default=1" binding:"min=1"`
	PageSize     int    `form:"page_size,default=20" binding:"min=1,max=100"`
	DepartmentID string `form:"department_id"`
	UserID       string `form:"user_id"`
	StartDate    string `form:"start_date" binding:"omitempty,datetime=2006-01-02"`
	EndDate      string `form:"end_date" binding:"omitempty,datetime=2006-01-02"`
	View         string `form:"view" binding:"omitempty,oneof=day week month"` // 视图类型
}

// CreateScheduleRequest 创建排班
type CreateScheduleRequest struct {
	UserID       string    `json:"user_id" binding:"required"`
	DepartmentID string    `json:"department_id"`
	DutyDate     string    `json:"duty_date" binding:"required"` // YYYY-MM-DD
	TimeSlot     string    `json:"time_slot" binding:"required,oneof=morning afternoon evening full_day"`
	Location     string    `json:"location" binding:"omitempty,max=100"`
	Remark       string    `json:"remark" binding:"omitempty,max=500"`
}

// BatchCreateScheduleRequest 批量排班（自动排班）
type BatchCreateScheduleRequest struct {
	UserIDs      []string `json:"user_ids" binding:"required,min=1"`
	DepartmentID string   `json:"department_id"`
	StartDate    string   `json:"start_date" binding:"required"` // YYYY-MM-DD
	EndDate      string   `json:"end_date" binding:"required"`   // YYYY-MM-DD
	TimeSlot     string   `json:"time_slot" binding:"required,oneof=morning afternoon evening full_day"`
	Rotation     bool     `json:"rotation"` // 是否轮换
	Location     string   `json:"location" binding:"omitempty,max=100"`
}

// UpdateScheduleRequest 调整排班（拖拽）
type UpdateScheduleRequest struct {
	UserID    string `json:"user_id"`
	DutyDate  string `json:"duty_date"`              // YYYY-MM-DD
	TimeSlot  string `json:"time_slot" binding:"omitempty,oneof=morning afternoon evening full_day"`
	Location  string `json:"location" binding:"omitempty,max=100"`
	Remark    string `json:"remark" binding:"omitempty,max=500"`
	Status    *int   `json:"status" binding:"omitempty,oneof=0 1 2 3 4"`
}

// SwapRequestDTO 调班申请
type SwapRequestDTO struct {
	TargetUserID       string `json:"target_user_id" binding:"required"`
	RequesterScheduleID string `json:"requester_schedule_id" binding:"required"`
	TargetScheduleID   string `json:"target_schedule_id"`
	Reason             string `json:"reason" binding:"required,min=5,max=500"`
}

// SwapActionRequest 审批调班
type SwapActionRequest struct {
	Status int    `json:"status" binding:"required,oneof=1 2"` // 1=批准 2=拒绝
	Remark string `json:"remark" binding:"omitempty,max=500"`
}

// DutyStatsRequest 值班统计查询
type DutyStatsRequest struct {
	DepartmentID string `form:"department_id"`
	UserID       string `form:"user_id"`
	StartDate    string `form:"start_date" binding:"omitempty,datetime=2006-01-02"`
	EndDate      string `form:"end_date" binding:"omitempty,datetime=2006-01-02"`
}

// ========== 响应 DTO ==========

// ScheduleResponse 排班响应
type ScheduleResponse struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	UserName      string    `json:"user_name"`
	DepartmentID string    `json:"department_id"`
	DeptName      string    `json:"dept_name"`
	DutyDate      string    `json:"duty_date"`
	TimeSlot      string    `json:"time_slot"`
	Location      string    `json:"location"`
	Remark        string    `json:"remark"`
	Status        int       `json:"status"`
	StatusText    string    `json:"status_text"`
	CreatedAt     string    `json:"created_at"`
	UpdatedAt     string    `json:"updated_at"`
}

// SwapResponse 调班申请响应
type SwapResponse struct {
	ID                  string `json:"id"`
	RequesterID         string `json:"requester_id"`
	RequesterName       string `json:"requester_name"`
	TargetUserID       string `json:"target_user_id"`
	TargetUserName     string `json:"target_user_name"`
	RequesterScheduleID string `json:"requester_schedule_id"`
	TargetScheduleID   string `json:"target_schedule_id"`
	Reason             string `json:"reason"`
	Status             int    `json:"status"`
	StatusText         string `json:"status_text"`
	ApproverID         string `json:"approver_id"`
	ApproverName       string `json:"approver_name"`
	ApprovedAt         string `json:"approved_at"`
	ApprovedRemark     string `json:"approved_remark"`
	CreatedAt          string `json:"created_at"`
	UpdatedAt          string `json:"updated_at"`
}

// DutyStatsResponse 值班统计响应
type DutyStatsResponse struct {
	TotalDuties    int64              `json:"total_duties"`
	CompletedDuties int64              `json:"completed_duties"`
	AbsentDuties   int64              `json:"absent_duties"`
	PendingDuties  int64              `json:"pending_duties"`
	CompletionRate  float64           `json:"completion_rate"` // 完成率
	UserStats      []UserDutyStatItem `json:"user_stats"`
}

// UserDutyStatItem 个人值班统计项
type UserDutyStatItem struct {
	UserID     string `json:"user_id"`
	UserName   string `json:"user_name"`
	DeptName   string `json:"dept_name"`
	TotalCount int64  `json:"total_count"`
	Completed  int64  `json:"completed"`
	Absent     int64  `json:"absent"`
	Rate       float64 `json:"rate"`
}

// statusTextMap 排班状态文本
var scheduleStatusText = map[int]string{
	0: "待值班", 1: "已到岗", 2: "已完成", 3: "缺勤", 4: "调班",
}

// swapStatusText 调班状态文本
var swapStatusTextMap = map[int]string{
	0: "待审批", 1: "已批准", 2: "已拒绝", 3: "已取消",
}

// GetScheduleStatusText 获取排班状态文本
func GetScheduleStatusText(status int) string {
	if text, ok := scheduleStatusText[status]; ok {
		return text
	}
	return "未知"
}

// GetSwapStatusText 获取调班状态文本
func GetSwapStatusText(status int) string {
	if text, ok := swapStatusTextMap[status]; ok {
		return text
	}
	return "未知"
}

// FormatDate 格式化日期
func FormatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

// FormatTime 格式化时间
func FormatTime(t time.Time) string {
	return t.Format(time.RFC3339)
}

// FormatTimePtr 格式化可空时间
func FormatTimePtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}
