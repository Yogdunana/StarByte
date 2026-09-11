// Package model 数据访问层 Model 定义
// 日程管理模块 (Schedule Module) - 对应 GitHub Issue #78
// 错误码范围: 10500-10999
//
// 四层架构: handler -> service -> repo -> model
// 本文件为 model 层，仅定义数据结构与表映射，不包含业务逻辑
package model

import (
	"time"

	"gorm.io/gorm"
)

// ============================================================================
// 枚举常量定义
// ============================================================================

// --- 日历类型 (CalendarType) ---

const (
	// CalendarTypePersonal 个人日历：仅创建者可见
	CalendarTypePersonal = "personal"
	// CalendarTypeShared 共享日历：可被指定用户/部门访问
	CalendarTypeShared = "shared"
)

// --- 事件状态 (EventStatus) ---

const (
	// EventStatusDraft 草稿：未确认的日程
	EventStatusDraft = "draft"
	// EventStatusConfirmed 已确认：日程已确认
	EventStatusConfirmed = "confirmed"
	// EventStatusCancelled 已取消
	EventStatusCancelled = "cancelled"
	// EventStatusCompleted 已完成：日程已结束
	EventStatusCompleted = "completed"
)

// --- 共享范围 (ShareScope) ---

const (
	// ShareScopePrivate 私有：仅所有者可见
	ShareScopePrivate = "private"
	// ShareScopeUser 指定用户：share_target_ids 中存储用户 ID 列表
	ShareScopeUser = "user"
	// ShareScopePublic 公开：所有人可见
	ShareScopePublic = "public"
)

// --- 提醒方式 (RemindMethod) ---

const (
	// RemindMethodApp 应用内通知
	RemindMethodApp = "app"
	// RemindMethodEmail 邮件提醒
	RemindMethodEmail = "email"
	// RemindMethodSms 短信提醒
	RemindMethodSms = "sms"
	// RemindMethodWebhook Webhook 推送
	RemindMethodWebhook = "webhook"
)

// --- 提醒状态 (ReminderStatus) ---

const (
	// ReminderStatusPending 待触发：提醒尚未到达触发时间
	ReminderStatusPending = "pending"
	// ReminderStatusTriggered 已触发：提醒已发送
	ReminderStatusTriggered = "triggered"
	// ReminderStatusSnoozed 已推迟：用户主动延后提醒
	ReminderStatusSnoozed = "snoozed"
	// ReminderStatusCanceled 已取消
	ReminderStatusCanceled = "canceled"
)

// --- 参与人响应状态 (AttendeeResponseStatus) ---

const (
	// AttendeeResponsePending 待回复
	AttendeeResponsePending = "pending"
	// AttendeeResponseAccepted 已接受
	AttendeeResponseAccepted = "accepted"
	// AttendeeResponseDeclined 已拒绝
	AttendeeResponseDeclined = "declined"
	// AttendeeResponseTentative 待定（可能参加）
	AttendeeResponseTentative = "tentative"
)

// --- 参与人角色 (AttendeeRole) ---

const (
	// AttendeeRoleOrganizer 组织者：事件所有者
	AttendeeRoleOrganizer = "organizer"
	// AttendeeRoleRequired 必须参加
	AttendeeRoleRequired = "required"
	// AttendeeRoleOptional 可选参加
	AttendeeRoleOptional = "optional"
)

// ============================================================================
// 通用基础模型
// ============================================================================

// BaseModel 通用基础字段，嵌入到各业务表中
// 包含主键、创建/更新/删除时间（软删除）
type BaseModel struct {
	// ID 主键自增ID
	ID uint `gorm:"primaryKey;autoIncrement;comment:主键ID" json:"id"`
	// CreatedAt 创建时间
	CreatedAt time.Time `gorm:"index;comment:创建时间" json:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt time.Time `gorm:"index;comment:更新时间" json:"updated_at"`
	// DeletedAt 软删除时间，非空表示已删除
	DeletedAt gorm.DeletedAt `gorm:"index;comment:删除时间" json:"deleted_at"`
}

// ============================================================================
// schedule_event 日程事件表
// ============================================================================

// ScheduleEvent 日程事件表
// 存储个人日历与共享日历的事件主体信息
// 通过 CalendarType 字段区分个人/共享日历，避免引入额外的 calendar 表
type ScheduleEvent struct {
	BaseModel

	// Title 事件标题，必填，长度 1-200
	Title string `gorm:"type:varchar(200);not null;comment:事件标题" json:"title"`

	// Description 事件描述，可选，支持富文本
	Description string `gorm:"type:text;comment:事件描述" json:"description"`

	// StartTime 开始时间，必填
	// 对于全天事件，使用日期 00:00:00
	StartTime time.Time `gorm:"not null;index;comment:开始时间" json:"start_time"`

	// EndTime 结束时间，必填，必须晚于 StartTime
	// 对于全天事件，使用日期 23:59:59
	EndTime time.Time `gorm:"not null;index;comment:结束时间" json:"end_time"`

	// AllDay 是否为全天事件
	AllDay bool `gorm:"default:false;comment:是否全天事件" json:"all_day"`

	// Location 事件地点，可选
	// 支持纯文本地址或会议室名称
	Location string `gorm:"type:varchar(255);comment:事件地点" json:"location"`

	// CalendarType 日历类型，区分个人日历与共享日历
	// 取值: personal(个人) / shared(共享)
	// 参见常量 CalendarTypePersonal / CalendarTypeShared
	CalendarType string `gorm:"type:varchar(20);not null;default:personal;index;comment:日历类型 personal-个人 shared-共享" json:"calendar_type"`

	// OwnerID 事件所有者用户ID
	// 个人日历: 即创建者本人
	// 共享日历: 为日历归属的用户ID，可与他人共享
	OwnerID uint `gorm:"not null;index;comment:所有者用户ID" json:"owner_id"`

	// ShareScope 共享范围，仅当 CalendarType=shared 时有效
	// 取值: private(私有) / user(指定用户) / public(公开)
	// 参见常量 ShareScopePrivate / ShareScopeUser / ShareScopePublic
	ShareScope string `gorm:"type:varchar(20);default:private;comment:共享范围 private-私有 user-指定用户 public-公开" json:"share_scope"`

	// ShareTargetIDs 共享目标用户ID列表，JSON 数组存储
	// 仅当 ShareScope=user 时有效，记录被共享的用户ID集合
	// 例: [1001, 1002, 1003]
	ShareTargetIDs string `gorm:"type:json;comment:共享目标用户ID列表(JSON数组)" json:"share_target_ids"`

	// RecurrenceRule RRULE 重复规则字符串
// 遵循 RFC 5545 规范，例: "FREQ=WEEKLY;INTERVAL=1;BYDAY=MO,WE,FR;UNTIL=20261231T235959Z"
	// 为空表示不重复的单次事件
	RecurrenceRule string `gorm:"type:varchar(255);comment:RRULE重复规则(RFC5545)" json:"recurrence_rule"`

	// RecurrenceID 重复事件母事件ID
	// 当本事件为某个重复规则的"例外"事件时，指向母事件ID
	// 为空表示本事件为母事件或单次事件
	RecurrenceID *uint `gorm:"index;comment:重复事件母事件ID(例外事件指向母事件)" json:"recurrence_id"`

	// Status 事件状态
	// 取值: draft(草稿) / confirmed(已确认) / cancelled(已取消) / completed(已完成)
	// 参见常量 EventStatusDraft / EventStatusConfirmed 等
	Status string `gorm:"type:varchar(20);not null;default:draft;index;comment:事件状态 draft-草稿 confirmed-已确认 cancelled-已取消 completed-已完成" json:"status"`

	// MeetingID 关联会议ID（软引用）
	// 指向未来 meeting 模块的主键，此处不加外键约束以解耦两模块
	// 为空表示本事件未关联会议
	MeetingID *uint `gorm:"index;comment:关联会议ID(软引用,无外键约束)" json:"meeting_id"`

	// CreatorID 创建者用户ID
	// 与 OwnerID 区别：共享日历中可能由他人代为创建
	CreatorID uint `gorm:"not null;comment:创建者用户ID" json:"creator_id"`

	// UpdaterID 最后更新者用户ID
	UpdaterID uint `gorm:"comment:最后更新者用户ID" json:"updater_id"`

	// Version 乐观锁版本号，用于并发冲突检测
	// 防止多人同时编辑同一日程导致数据覆盖
	Version int `gorm:"default:1;comment:乐观锁版本号" json:"version"`
}

// TableName 指定日程事件表名
func (ScheduleEvent) TableName() string {
	return "schedule_event"
}

// ============================================================================
// schedule_reminder 提醒表
// ============================================================================

// ScheduleReminder 提醒表
// 为日程事件配置提醒规则，一个事件可关联多个提醒
type ScheduleReminder struct {
	BaseModel

	// EventID 关联的日程事件ID
	EventID uint `gorm:"not null;index;comment:关联日程事件ID" json:"event_id"`

	// UserID 接收提醒的用户ID
	// 同一事件不同参与人可有不同的提醒配置
	UserID uint `gorm:"not null;index;comment:接收提醒的用户ID" json:"user_id"`

	// RemindTime 提醒触发时间
	// 由系统根据事件开始时间减去偏移量计算得出
	// 调度器据此时间触发提醒
	RemindTime time.Time `gorm:"not null;index;comment:提醒触发时间" json:"remind_time"`

	// RemindOffset 提前提醒的分钟数
	// 例: 15 表示事件开始前15分钟提醒
	// 0 表示准时提醒（事件开始时触发）
	RemindOffset int `gorm:"not null;comment:提前提醒分钟数(0=准时)" json:"remind_offset"`

	// RemindMethod 提醒方式
	// 取值: app(应用内) / email(邮件) / sms(短信) / webhook(推送)
	// 参见常量 RemindMethodApp / RemindMethodEmail 等
	RemindMethod string `gorm:"type:varchar(20);not null;default:app;comment:提醒方式 app-应用内 email-邮件 sms-短信 webhook-推送" json:"remind_method"`

	// Status 提醒状态
	// 取值: pending(待触发) / triggered(已触发) / snoozed(已推迟) / canceled(已取消)
	// 参见常量 ReminderStatusPending / ReminderStatusTriggered 等
	Status string `gorm:"type:varchar(20);not null;default:pending;index;comment:提醒状态 pending-待触发 triggered-已触发 snoozed-已推迟 canceled-已取消" json:"status"`

	// SnoozeMinutes 推迟时长（分钟）
	// 当用户选择"稍后提醒"时，记录推迟的分钟数，便于重新计算 RemindTime
	SnoozeMinutes int `gorm:"default:0;comment:推迟时长(分钟)" json:"snooze_minutes"`

	// TriggeredAt 实际触发时间，用于审计与去重
	TriggeredAt *time.Time `gorm:"comment:实际触发时间" json:"triggered_at"`
}

// TableName 指定提醒表名
func (ScheduleReminder) TableName() string {
	return "schedule_reminder"
}

// ============================================================================
// schedule_event_attendee 参与人关联表
// ============================================================================

// ScheduleEventAttendee 参与人关联表
// 实现 schedule_event 与用户的多对多关联
// 支持按用户反查日程、管理参与人响应状态与角色
type ScheduleEventAttendee struct {
	BaseModel

	// EventID 关联的日程事件ID
	EventID uint `gorm:"not null;index;comment:关联日程事件ID" json:"event_id"`

	// UserID 参与人用户ID
	UserID uint `gorm:"not null;index;comment:参与人用户ID" json:"user_id"`

	// ResponseStatus 响应状态
	// 取值: pending(待回复) / accepted(已接受) / declined(已拒绝) / tentative(待定)
	// 参见常量 AttendeeResponsePending / AttendeeResponseAccepted 等
	ResponseStatus string `gorm:"type:varchar(20);not null;default:pending;comment:响应状态 pending-待回复 accepted-已接受 declined-已拒绝 tentative-待定" json:"response_status"`

	// Role 参与人角色
	// 取值: organizer(组织者) / required(必须参加) / optional(可选参加)
	// 参见常量 AttendeeRoleOrganizer / AttendeeRoleRequired 等
	Role string `gorm:"type:varchar(20);not null;default:required;comment:参与人角色 organizer-组织者 required-必须参加 optional-可选参加" json:"role"`

	// RespondedAt 用户响应时间，记录接受/拒绝的时间
	RespondedAt *time.Time `gorm:"comment:用户响应时间" json:"responded_at"`
}

// TableName 指定参与人关联表名
func (ScheduleEventAttendee) TableName() string {
	return "schedule_event_attendee"
}
