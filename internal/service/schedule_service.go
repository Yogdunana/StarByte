// Package service 业务逻辑层接口定义
// 日程管理模块 (Schedule Module) - 对应 GitHub Issue #78
//
// 四层架构: handler -> service -> repo -> model
// 本文件定义 service 层接口契约 + 业务错误码 + 请求/响应 DTO
//
// 错误码分配 (按子域分段):
//   10500-10599 通用错误 (参数校验/权限/事务/未找到)
//   10600-10699 日程事件 (ScheduleEvent)
//   10700-10799 提醒 (ScheduleReminder)
//   10800-10899 参与人 (ScheduleEventAttendee)
//   10900-10999 重复规则/其他
//
// 设计原则:
//   1. service 层依赖 repo 接口（不依赖具体实现）
//   2. service 层捕获 repo sentinel 错误并映射为业务错误码
//   3. service 层不感知 HTTP，输入/输出均为 DTO
//   4. 事务边界由 service 管理（通过 TransactionManager）
package service

import (
	"context"
	"errors"
	"time"

	"schedule-service/internal/model"
	"schedule-service/internal/repo"
)

// ============================================================================
// 业务错误码定义 (10500-10999)
// ============================================================================

// ErrCode 业务错误码类型
type ErrCode int

const (
	// ----- 10500-10599 通用错误 -----

	// ErrCodeInvalidParam 参数校验失败
	ErrCodeInvalidParam ErrCode = 10501
	// ErrCodeUnauthorized 未授权（未登录或 token 无效）
	ErrCodeUnauthorized ErrCode = 10502
	// ErrCodeForbidden 无权限操作该资源
	ErrCodeForbidden ErrCode = 10503
	// ErrCodeNotFound 资源不存在
	ErrCodeNotFound ErrCode = 10504
	// ErrCodeTxFailed 事务执行失败
	ErrCodeTxFailed ErrCode = 10505
	// ErrCodeInternal 内部错误（非预期）
	ErrCodeInternal ErrCode = 10599

	// ----- 10600-10699 日程事件 -----

	// ErrCodeEventNotFound 事件不存在
	ErrCodeEventNotFound ErrCode = 10601
	// ErrCodeEventTimeInvalid 时间区间非法（结束时间早于开始时间）
	ErrCodeEventTimeInvalid ErrCode = 10602
	// ErrCodeEventConcurrent 并发冲突（他人已修改，乐观锁失败）
	ErrCodeEventConcurrent ErrCode = 10603
	// ErrCodeEventDuplicate 同一所有者下重复事件（标题+时间）
	ErrCodeEventDuplicate ErrCode = 10604
	// ErrCodeEventRRULEInvalid RRULE 格式非法
	ErrCodeEventRRULEInvalid ErrCode = 10605
	// ErrCodeEventMeetingLinked 该事件已关联会议，无法重复关联
	ErrCodeEventMeetingLinked ErrCode = 10606
	// ErrCodeEventStatusTransition 非法状态转换（如已完成→草稿）
	ErrCodeEventStatusTransition ErrCode = 10607

	// ----- 10700-10799 提醒 -----

	// ErrCodeReminderNotFound 提醒不存在
	ErrCodeReminderNotFound ErrCode = 10701
	// ErrCodeReminderOffsetInvalid 提前提醒分钟数非法（负数或过大）
	ErrCodeReminderOffsetInvalid ErrCode = 10702
	// ErrCodeReminderAlreadyTriggered 提醒已触发，无法再次触发或推迟
	ErrCodeReminderAlreadyTriggered ErrCode = 10703
	// ErrCodeReminderSnoozeExceed 推迟次数/时长超限
	ErrCodeReminderSnoozeExceed ErrCode = 10704
	// ErrCodeReminderMethodUnsupported 不支持的提醒方式
	ErrCodeReminderMethodUnsupported ErrCode = 10705

	// ----- 10800-10899 参与人 -----

	// ErrCodeAttendeeNotFound 参与人记录不存在
	ErrCodeAttendeeNotFound ErrCode = 10801
	// ErrCodeAttendeeDuplicate 该用户已是参与人
	ErrCodeAttendeeDuplicate ErrCode = 10802
	// ErrCodeAttendeeCountExceed 参与人数量超限
	ErrCodeAttendeeCountExceed ErrCode = 10803
	// ErrCodeAttendeeResponseInvalid 非法的响应状态
	ErrCodeAttendeeResponseInvalid ErrCode = 10804
	// ErrCodeAttendeeOrganizerRequired 组织者不可移除
	ErrCodeAttendeeOrganizerRequired ErrCode = 10805

	// ----- 10900-10999 重复规则/其他 -----

	// ErrCodeRRULEParseFailed RRULE 解析失败
	ErrCodeRRULEParseFailed ErrCode = 10901
	// ErrCodeRRULEExpandFailed RRULE 展开失败（生成实例错误）
	ErrCodeRRULEExpandFailed ErrCode = 10902
	// ErrCodeRecurrenceRangeExceed 展开时间区间过大（拒绝展开）
	ErrCodeRecurrenceRangeExceed ErrCode = 10903
)

// BizError 业务错误
// 携带错误码与可读消息，由 handler 层转换为 HTTP 响应
type BizError struct {
	// Code 业务错误码
	Code ErrCode
	// Message 面向用户的错误描述（脱敏）
	Message string
	// Err 原始底层错误（仅日志记录，不返回客户端）
	Err error
}

// Error 实现 error 接口
func (e *BizError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

// Unwrap 支持 errors.Is/As
func (e *BizError) Unwrap() error {
	return e.Err
}

// NewBizError 构造业务错误
func NewBizError(code ErrCode, msg string, err error) *BizError {
	return &BizError{Code: code, Message: msg, Err: err}
}

// 预定义通用业务错误
var (
	// ErrInvalidParam 参数校验失败
	ErrInvalidParam = NewBizError(ErrCodeInvalidParam, "参数校验失败", nil)
	// ErrUnauthorized 未授权
	ErrUnauthorized = NewBizError(ErrCodeUnauthorized, "未授权", nil)
	// ErrForbidden 无权限
	ErrForbidden = NewBizError(ErrCodeForbidden, "无权限操作该资源", nil)
)

// ============================================================================
// 通用 DTO
// ============================================================================

// PageRequest 通用分页请求
type PageRequest struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	OrderBy  string `json:"order_by"`
	Order    string `json:"order"`
}

// PageResponse 通用分页响应
type PageResponse[T any] struct {
	List     []T   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

// ============================================================================
// 日程事件 DTO
// ============================================================================

// CreateEventRequest 创建日程事件请求
type CreateEventRequest struct {
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	AllDay        bool      `json:"all_day"`
	Location      string    `json:"location"`
	CalendarType  string    `json:"calendar_type"`
	ShareScope    string    `json:"share_scope"`
	ShareTargetIDs []uint   `json:"share_target_ids"`
	RecurrenceRule string   `json:"recurrence_rule"`
	Status        string    `json:"status"`
	MeetingID     *uint     `json:"meeting_id"`
	// Attendees 参与人初始列表，事件创建时一起写入
	Attendees []CreateAttendeeRequest `json:"attendees"`
	// Reminders 提醒初始列表，事件创建时一起写入
	Reminders []CreateReminderRequest `json:"reminders"`
}

// UpdateEventRequest 更新日程事件请求
// 仅包含允许更新的字段，OwnerID/CreatorID 不可改
type UpdateEventRequest struct {
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time"`
	AllDay         bool      `json:"all_day"`
	Location       string    `json:"location"`
	CalendarType   string    `json:"calendar_type"`
	ShareScope     string    `json:"share_scope"`
	ShareTargetIDs []uint    `json:"share_target_ids"`
	RecurrenceRule string    `json:"recurrence_rule"`
	Status         string    `json:"status"`
	MeetingID      *uint     `json:"meeting_id"`
	// Version 乐观锁版本号，必填，必须与当前数据库一致
	Version int `json:"version"`
}

// EventResponse 日程事件响应
type EventResponse struct {
	ID             uint      `json:"id"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time"`
	AllDay         bool      `json:"all_day"`
	Location       string    `json:"location"`
	CalendarType   string    `json:"calendar_type"`
	OwnerID        uint      `json:"owner_id"`
	ShareScope     string    `json:"share_scope"`
	ShareTargetIDs []uint    `json:"share_target_ids"`
	RecurrenceRule string    `json:"recurrence_rule"`
	RecurrenceID   *uint     `json:"recurrence_id"`
	Status         string    `json:"status"`
	MeetingID      *uint     `json:"meeting_id"`
	CreatorID      uint      `json:"creator_id"`
	Version        int       `json:"version"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	// Attendees 参与人列表（查询时按需填充）
	Attendees []*AttendeeResponse `json:"attendees,omitempty"`
	// Reminders 当前用户的提醒列表（按需填充）
	Reminders []*ReminderResponse `json:"reminders,omitempty"`
}

// EventListRequest 日程事件列表查询请求
type EventListRequest struct {
	PageRequest
	OwnerID         uint      `json:"owner_id"`
	CalendarType    string    `json:"calendar_type"`
	Status          string    `json:"status"`
	StartTimeFrom   time.Time `json:"start_time_from"`
	StartTimeTo     time.Time `json:"start_time_to"`
	MeetingID       *uint     `json:"meeting_id"`
	ExcludeCancelled bool      `json:"exclude_cancelled"`
}

// EventListResponse 日程事件列表响应
type EventListResponse = PageResponse[*EventResponse]

// ============================================================================
// 提醒 DTO
// ============================================================================

// CreateReminderRequest 创建提醒请求
type CreateReminderRequest struct {
	EventID      uint   `json:"event_id"`
	UserID       uint   `json:"user_id"`
	RemindOffset int    `json:"remind_offset"`
	RemindMethod string `json:"remind_method"`
}

// UpdateReminderRequest 更新提醒请求
type UpdateReminderRequest struct {
	RemindOffset int    `json:"remind_offset"`
	RemindMethod string `json:"remind_method"`
}

// ReminderResponse 提醒响应
type ReminderResponse struct {
	ID            uint       `json:"id"`
	EventID       uint       `json:"event_id"`
	UserID        uint       `json:"user_id"`
	RemindTime    time.Time  `json:"remind_time"`
	RemindOffset  int        `json:"remind_offset"`
	RemindMethod  string     `json:"remind_method"`
	Status        string     `json:"status"`
	SnoozeMinutes int        `json:"snooze_minutes"`
	TriggeredAt   *time.Time `json:"triggered_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// ReminderListRequest 提醒列表查询请求
type ReminderListRequest struct {
	PageRequest
	EventID  uint      `json:"event_id"`
	UserID   uint      `json:"user_id"`
	Status   string    `json:"status"`
	TriggerBefore time.Time `json:"trigger_before"`
}

// ReminderListResponse 提醒列表响应
type ReminderListResponse = PageResponse[*ReminderResponse]

// SnoozeRequest 推迟提醒请求
type SnoozeRequest struct {
	// SnoozeMinutes 推迟的分钟数
	SnoozeMinutes int `json:"snooze_minutes"`
}

// ============================================================================
// 参与人 DTO
// ============================================================================

// CreateAttendeeRequest 添加参与人请求
type CreateAttendeeRequest struct {
	UserID uint   `json:"user_id"`
	Role   string `json:"role"`
}

// AttendeeResponse 参与人响应
type AttendeeResponse struct {
	ID             uint       `json:"id"`
	EventID        uint       `json:"event_id"`
	UserID         uint       `json:"user_id"`
	ResponseStatus string     `json:"response_status"`
	Role           string     `json:"role"`
	RespondedAt    *time.Time `json:"responded_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// AttendeeListRequest 参与人列表查询请求
type AttendeeListRequest struct {
	PageRequest
	EventID        uint   `json:"event_id"`
	UserID         uint   `json:"user_id"`
	Role           string `json:"role"`
	ResponseStatus string `json:"response_status"`
}

// AttendeeListResponse 参与人列表响应
type AttendeeListResponse = PageResponse[*AttendeeResponse]

// RespondInviteRequest 参与人响应邀请请求
type RespondInviteRequest struct {
	// Status 响应状态: accepted / declined / tentative
	Status string `json:"status"`
}

// ============================================================================
// 调度器与扩展器抽象
// ============================================================================

// ReminderDispatcher 提醒发送器接口
// service 层依赖此抽象，实现可替换：
//   - 开发期: SyncDispatcher (日志打印)
//   - 生产期: QueueDispatcher (投递到消息队列)
//   - 高可用: TimerDispatcher (基于时间轮/定时器)
type ReminderDispatcher interface {
	// Dispatch 发送提醒
	//   - reminder: 待发送的提醒实体
	//   - event: 关联的日程事件（提供标题/时间等上下文）
	//   - 返回 error 表示发送失败，调度器可选择重试
	Dispatch(ctx context.Context, reminder *model.ScheduleReminder, event *model.ScheduleEvent) error
}

// RecurrenceExpander RRULE 重复规则展开器接口
// 职责：根据 RRULE 字符串和母事件，生成时间区间内的实例列表
// 实现：可使用 github.com/teambition/rrule-go 或自研
type RecurrenceExpander interface {
	// Expand 在 [from, to] 时间区间内展开重复事件实例
	//   - rule: RRULE 字符串 (RFC 5545)
	//   - baseStartTime: 母事件开始时间，作为展开基准
	//   - from/to: 展开区间（含两端）
	// 返回实例的开始时间列表，调用方据此生成虚拟 EventResponse
	Expand(ctx context.Context, rule string, baseStartTime time.Time, from, to time.Time) ([]time.Time, error)

	// Validate 校验 RRULE 字符串合法性
	//   创建/更新事件时调用，非法则返回 ErrCodeEventRRULEInvalid
	Validate(ctx context.Context, rule string) error
}

// ============================================================================
// ScheduleEventService 日程事件业务接口
// ============================================================================

// ScheduleEventService 日程事件业务接口
// 编排 repo + dispatcher + expander，处理事务与错误映射
type ScheduleEventService interface {
	// ----- 基础 CRUD -----

	// Create 创建日程事件
	//   - 事务内: 写 event → 写 attendees → 写 reminders
	//   - 校验: 时间区间合法、RRULE 合法、参与人去重
	//   - actorID: 操作者用户ID，用于权限校验与 CreatorID 赋值
	Create(ctx context.Context, actorID uint, req *CreateEventRequest) (*EventResponse, error)

	// Update 更新日程事件
	//   - 校验: 操作者是否为 owner、乐观锁版本号
	//   - 状态转换合法性（如已完成不可改回草稿）
	Update(ctx context.Context, actorID, eventID uint, req *UpdateEventRequest) (*EventResponse, error)

	// Delete 删除日程事件
	//   - 事务内: 软删 event → 软删 attendees → 取消 reminders
	Delete(ctx context.Context, actorID, eventID uint) error

	// GetByID 查询单个事件详情
	//   - 权限校验: 操作者须为 owner 或被共享者或参与人
	//   - 填充 attendees 与当前用户的 reminders
	GetByID(ctx context.Context, actorID, eventID uint) (*EventResponse, error)

	// ----- 列表查询 -----

	// List 组合条件分页查询
	List(ctx context.Context, actorID uint, req *EventListRequest) (*EventListResponse, error)

	// ListByTimeRange 日/周/月视图查询
	//   - 返回: 普通事件 + 重复事件展开后的虚拟实例
	//   - 展开由 RecurrenceExpander 完成
	ListByTimeRange(ctx context.Context, actorID uint, from, to time.Time) ([]*EventResponse, error)

	// ListVisibleByUser 查询用户可见的所有日程（自己拥有 + 被共享）
	ListVisibleByUser(ctx context.Context, actorID uint, req *PageRequest) (*EventListResponse, error)

	// ListByAttendee 查询用户作为参与人的日程
	ListByAttendee(ctx context.Context, actorID uint, req *PageRequest) (*EventListResponse, error)

	// ----- 状态与关联 -----

	// UpdateStatus 更新事件状态
	//   - 校验: 状态转换合法性、操作者权限
	UpdateStatus(ctx context.Context, actorID, eventID uint, status string) error

	// LinkMeeting 关联会议（meetingID 非 nil）
	//   - 校验: 事件未关联其他会议
	LinkMeeting(ctx context.Context, actorID, eventID, meetingID uint) error

	// UnlinkMeeting 取消会议关联
	UnlinkMeeting(ctx context.Context, actorID, eventID uint) error
}

// ============================================================================
// ScheduleReminderService 提醒业务接口
// ============================================================================

// ScheduleReminderService 提醒业务接口
type ScheduleReminderService interface {
	// ----- 基础 CRUD -----

	// Create 为事件添加提醒
	//   - 校验: 事件存在、用户为参与人、offset 合法、method 支持
	//   - 计算 RemindTime = event.StartTime - offset
	Create(ctx context.Context, actorID uint, req *CreateReminderRequest) (*ReminderResponse, error)

	// Update 更新提醒配置
	Update(ctx context.Context, actorID, reminderID uint, req *UpdateReminderRequest) (*ReminderResponse, error)

	// Delete 删除提醒
	Delete(ctx context.Context, actorID, reminderID uint) error

	// GetByID 查询提醒详情
	GetByID(ctx context.Context, actorID, reminderID uint) (*ReminderResponse, error)

	// ----- 列表查询 -----

	// ListByEvent 查询某事件的全部提醒
	ListByEvent(ctx context.Context, actorID, eventID uint) ([]*ReminderResponse, error)

	// ListByUser 查询某用户的全部提醒
	ListByUser(ctx context.Context, actorID uint, req *ReminderListRequest) (*ReminderListResponse, error)

	// ----- 触发与推迟 -----

	// Snooze 推迟提醒
	//   - 校验: 当前状态为 pending 或 snoozed
	//   - 计算: 新 RemindTime = now + snoozeMinutes
	//   - 限制: 累计推迟次数与最大推迟时长
	Snooze(ctx context.Context, actorID, reminderID uint, req *SnoozeRequest) (*ReminderResponse, error)

	// Cancel 取消提醒
	Cancel(ctx context.Context, actorID, reminderID uint) error

	// ----- 调度器入口 -----

	// FirePending 触发已到时间的待处理提醒
	//   - 由调度器（cron/timer）周期性调用
	//   - 内部: 查询 pending → 逐条 Dispatch → 标记 triggered
	//   - 失败重试与死信处理由实现决定
	FirePending(ctx context.Context, before time.Time, limit int) (int, error)
}

// ============================================================================
// ScheduleAttendeeService 参与人业务接口
// ============================================================================

// ScheduleAttendeeService 参与人业务接口
type ScheduleAttendeeService interface {
	// ----- 基础操作 -----

	// Add 添加参与人到事件
	//   - 校验: 事件存在、操作者为 owner、用户未已是参与人、数量未超限
	//   - 组织者角色不可重复添加（每事件唯一）
	Add(ctx context.Context, actorID, eventID uint, req *CreateAttendeeRequest) (*AttendeeResponse, error)

	// BatchAdd 批量添加参与人
	//   - 事务内批量写入，去重 + 数量校验
	BatchAdd(ctx context.Context, actorID, eventID uint, reqs []CreateAttendeeRequest) ([]*AttendeeResponse, error)

	// Update 更新参与人角色
	//   - 仅 owner 可调用；组织者角色变更需特殊处理
	Update(ctx context.Context, actorID, attendeeID uint, role string) (*AttendeeResponse, error)

	// Remove 移除参与人
	//   - 校验: 操作者为 owner 或本人移除自己
	//   - 组织者不可被移除
	Remove(ctx context.Context, actorID, attendeeID uint) error

	// GetByID 查询参与人记录详情
	GetByID(ctx context.Context, actorID, attendeeID uint) (*AttendeeResponse, error)

	// ----- 列表查询 -----

	// ListByEvent 查询某事件的参与人列表
	ListByEvent(ctx context.Context, actorID, eventID uint, req *PageRequest) (*AttendeeListResponse, error)

	// ListByUser 查询用户参与的日程关联列表
	ListByUser(ctx context.Context, actorID uint, req *AttendeeListRequest) (*AttendeeListResponse, error)

	// ----- 响应邀请 -----

	// Respond 响应事件邀请
	//   - actorID 必须为参与人本人
	//   - status: accepted / declined / tentative
	//   - 自动记录响应时间
	Respond(ctx context.Context, actorID, attendeeID uint, req *RespondInviteRequest) (*AttendeeResponse, error)

	// ----- 统计 -----

	// CountByEvent 统计事件参与人数（可按角色过滤）
	CountByEvent(ctx context.Context, actorID, eventID uint, role string) (int64, error)
}
