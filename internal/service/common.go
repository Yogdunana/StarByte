// Package service 业务逻辑层实现 - 公共辅助
// 日程管理模块 (Schedule Module) - 对应 GitHub Issue #78
//
// 本文件包含:
//   1. repo sentinel 错误到 BizError 的映射
//   2. Model → Response DTO 的转换函数
//   3. 通用校验辅助
package service

import (
	"encoding/json"
	"errors"
	"time"

	"schedule-service/internal/model"
	"schedule-service/internal/repo"
)

// ============================================================================
// repo 错误 → BizError 映射
// ============================================================================

// wrapRepoErr 将 repo 层 sentinel 错误包装为对应子域的 BizError
//   notFoundCode: 资源不存在时使用的错误码（如 ErrCodeEventNotFound）
//   domain: 子域名称，用于日志区分
//   调用示例: wrapRepoErr(err, ErrCodeEventNotFound, "event")
func wrapRepoErr(err error, notFoundCode ErrCode, domain string) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, repo.ErrNotFound):
		return NewBizError(notFoundCode, domain+" not found", err)
	case errors.Is(err, repo.ErrDuplicate):
		return NewBizError(ErrCodeInvalidParam, domain+" already exists", err)
	case errors.Is(err, repo.ErrConcurrent):
		return NewBizError(ErrCodeEventConcurrent, domain+" has been modified by others", err)
	case errors.Is(err, repo.ErrTxInContext):
		return NewBizError(ErrCodeTxFailed, "transaction context error", err)
	default:
		return NewBizError(ErrCodeInternal, "internal error", err)
	}
}

// wrapEventErr 事件子域错误映射快捷函数
func wrapEventErr(err error) error {
	return wrapRepoErr(err, ErrCodeEventNotFound, "event")
}

// wrapReminderErr 提醒子域错误映射快捷函数
func wrapReminderErr(err error) error {
	return wrapRepoErr(err, ErrCodeReminderNotFound, "reminder")
}

// wrapAttendeeErr 参与人子域错误映射快捷函数
func wrapAttendeeErr(err error) error {
	return wrapRepoErr(err, ErrCodeAttendeeNotFound, "attendee")
}

// ============================================================================
// Model → Response DTO 转换
// ============================================================================

// eventToResponse ScheduleEvent → EventResponse
//   attendees/reminders 默认不填充，由调用方按需赋值
func eventToResponse(e *model.ScheduleEvent) *EventResponse {
	if e == nil {
		return nil
	}
	resp := &EventResponse{
		ID:             e.ID,
		Title:          e.Title,
		Description:    e.Description,
		StartTime:      e.StartTime,
		EndTime:        e.EndTime,
		AllDay:         e.AllDay,
		Location:       e.Location,
		CalendarType:   e.CalendarType,
		OwnerID:        e.OwnerID,
		ShareScope:     e.ShareScope,
		RecurrenceRule: e.RecurrenceRule,
		RecurrenceID:   e.RecurrenceID,
		Status:         e.Status,
		MeetingID:      e.MeetingID,
		CreatorID:      e.CreatorID,
		Version:        e.Version,
		CreatedAt:      e.CreatedAt,
		UpdatedAt:      e.UpdatedAt,
	}
	// 解析 ShareTargetIDs JSON 字符串为 []uint
	resp.ShareTargetIDs = parseShareTargetIDs(e.ShareTargetIDs)
	return resp
}

// eventsToListResponse 批量转换并包装为分页响应
func eventsToListResponse(list []*model.ScheduleEvent, total int64, page, pageSize int) *EventListResponse {
	items := make([]*EventResponse, 0, len(list))
	for _, e := range list {
		items = append(items, eventToResponse(e))
	}
	return &EventListResponse{
		List:     items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
}

// reminderToResponse ScheduleReminder → ReminderResponse
func reminderToResponse(r *model.ScheduleReminder) *ReminderResponse {
	if r == nil {
		return nil
	}
	return &ReminderResponse{
		ID:            r.ID,
		EventID:       r.EventID,
		UserID:        r.UserID,
		RemindTime:    r.RemindTime,
		RemindOffset:  r.RemindOffset,
		RemindMethod:  r.RemindMethod,
		Status:        r.Status,
		SnoozeMinutes: r.SnoozeMinutes,
		TriggeredAt:   r.TriggeredAt,
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
	}
}

// remindersToListResponse 批量转换并包装为分页响应
func remindersToListResponse(list []*model.ScheduleReminder, total int64, page, pageSize int) *ReminderListResponse {
	items := make([]*ReminderResponse, 0, len(list))
	for _, r := range list {
		items = append(items, reminderToResponse(r))
	}
	return &ReminderListResponse{
		List:     items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
}

// attendeeToResponse ScheduleEventAttendee → AttendeeResponse
func attendeeToResponse(a *model.ScheduleEventAttendee) *AttendeeResponse {
	if a == nil {
		return nil
	}
	return &AttendeeResponse{
		ID:             a.ID,
		EventID:        a.EventID,
		UserID:         a.UserID,
		ResponseStatus: a.ResponseStatus,
		Role:           a.Role,
		RespondedAt:    a.RespondedAt,
		CreatedAt:      a.CreatedAt,
		UpdatedAt:      a.UpdatedAt,
	}
}

// attendeesToListResponse 批量转换并包装为分页响应
func attendeesToListResponse(list []*model.ScheduleEventAttendee, total int64, page, pageSize int) *AttendeeListResponse {
	items := make([]*AttendeeResponse, 0, len(list))
	for _, a := range list {
		items = append(items, attendeeToResponse(a))
	}
	return &AttendeeListResponse{
		List:     items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
}

// ============================================================================
// 辅助函数
// ============================================================================

// parseShareTargetIDs 解析共享目标用户 ID JSON 数组字符串为 []uint
//   输入示例: "[1001,1002,1003]"
//   非法或空字符串返回空切片
func parseShareTargetIDs(s string) []uint {
	if s == "" || s == "null" {
		return []uint{}
	}
	var ids []uint
	if err := json.Unmarshal([]byte(s), &ids); err != nil {
		// 解析失败返回空切片，避免阻塞主流程
		return []uint{}
	}
	return ids
}

// formatShareTargetIDs 将 []uint 格式化为 JSON 字符串
//   空列表返回空字符串（DB 中允许）
func formatShareTargetIDs(ids []uint) string {
	if len(ids) == 0 {
		return ""
	}
	b, err := json.Marshal(ids)
	if err != nil {
		return ""
	}
	return string(b)
}

// toRepoQueryParams 将 service.PageRequest 转为 repo.QueryParams
func toRepoQueryParams(p PageRequest) repo.QueryParams {
	return repo.QueryParams{
		Page:     p.Page,
		PageSize: p.PageSize,
		OrderBy:  p.OrderBy,
		Order:    p.Order,
	}
}

// ----------------------------------------------------------------------------
// 校验辅助
// ----------------------------------------------------------------------------

// validateEventTime 校验事件时间区间合法性
//   规则: EndTime 必须晚于 StartTime；全天事件允许同日不同时
func validateEventTime(start, end time.Time, allDay bool) error {
	if start.IsZero() || end.IsZero() {
		return NewBizError(ErrCodeEventTimeInvalid, "start_time and end_time are required", nil)
	}
	if !end.After(start) {
		return NewBizError(ErrCodeEventTimeInvalid, "end_time must be after start_time", nil)
	}
	return nil
}

// validateCalendarType 校验日历类型
func validateCalendarType(t string) bool {
	switch t {
	case model.CalendarTypePersonal, model.CalendarTypeShared:
		return true
	}
	return false
}

// validateShareScope 校验共享范围
func validateShareScope(s string) bool {
	switch s {
	case model.ShareScopePrivate, model.ShareScopeUser, model.ShareScopePublic:
		return true
	}
	return false
}

// validateEventStatus 校验事件状态
func validateEventStatus(s string) bool {
	switch s {
	case model.EventStatusDraft, model.EventStatusConfirmed,
		model.EventStatusCancelled, model.EventStatusCompleted:
		return true
	}
	return false
}

// validateRemindMethod 校验提醒方式
func validateRemindMethod(m string) bool {
	switch m {
	case model.RemindMethodApp, model.RemindMethodEmail,
		model.RemindMethodSms, model.RemindMethodWebhook:
		return true
	}
	return false
}

// validateReminderStatus 校验提醒状态
func validateReminderStatus(s string) bool {
	switch s {
	case model.ReminderStatusPending, model.ReminderStatusTriggered,
		model.ReminderStatusSnoozed, model.ReminderStatusCanceled:
		return true
	}
	return false
}

// validateAttendeeRole 校验参与人角色
func validateAttendeeRole(r string) bool {
	switch r {
	case model.AttendeeRoleOrganizer, model.AttendeeRoleRequired, model.AttendeeRoleOptional:
		return true
	}
	return false
}

// validateResponseStatus 校验参与人响应状态
func validateResponseStatus(s string) bool {
	switch s {
	case model.AttendeeResponsePending, model.AttendeeResponseAccepted,
		model.AttendeeResponseDeclined, model.AttendeeResponseTentative:
		return true
	}
	return false
}

// eventStatusTransitionAllowed 校验事件状态转换合法性
//   draft     → confirmed / cancelled
//   confirmed → completed / cancelled
//   completed → (终态，不可转换)
//   cancelled → (终态，不可转换)
func eventStatusTransitionAllowed(from, to string) bool {
	if from == to {
		return true
	}
	switch from {
	case model.EventStatusDraft:
		return to == model.EventStatusConfirmed || to == model.EventStatusCancelled
	case model.EventStatusConfirmed:
		return to == model.EventStatusCompleted || to == model.EventStatusCancelled
	}
	return false
}

// maxAttendeesPerEvent 单事件参与人上限
const maxAttendeesPerEvent = 1000

// maxSnoozeMinutes 单次推迟最大分钟数（24小时）
const maxSnoozeMinutes = 1440

// maxSnoozeCount 累计推迟次数上限
const maxSnoozeCount = 10

// maxReminderOffset 提前提醒最大分钟数（30天）
const maxReminderOffset = 43200
