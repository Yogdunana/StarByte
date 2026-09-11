// Package repo 数据访问层 Repository 接口定义
// 日程管理模块 (Schedule Module) - 对应 GitHub Issue #78
// 错误码范围: 10500-10999 (由 service 层转换，repo 层返回原始/sentinel 错误)
//
// 四层架构: handler -> service -> repo -> model
// 本文件定义 repo 层接口契约，不包含实现
//
// 事务策略: 通过 context.Context 传递 *gorm.DB
//   - service 层调用 TransactionManager.Transaction 开启事务
//   - 在事务闭包内，repo 通过 ctx 取出事务 *gorm.DB
//   - 无事务时，repo 使用默认 *gorm.DB
package repo

import (
	"context"
	"errors"
	"time"

	"schedule-service/internal/model"

	"gorm.io/gorm"
)

// ============================================================================
// 通用 Sentinel 错误
// ============================================================================

// 错误说明:
// repo 层不直接返回业务错误码，而是返回 sentinel 错误
// 由 service 层捕获后映射为 10500-10999 范围的业务错误码

var (
	// ErrNotFound 记录不存在（命中软删除或主键不存在）
	ErrNotFound = errors.New("record not found")
	// ErrDuplicate 唯一键冲突（如重复创建事件）
	ErrDuplicate = errors.New("duplicate record")
	// ErrConcurrent 并发冲突（乐观锁版本号不匹配）
	ErrConcurrent = errors.New("concurrent update conflict")
	// ErrTxInContext ctx 中未找到事务 DB（编码错误）
	ErrTxInContext = errors.New("transaction db not found in context")
)

// ============================================================================
// 通用查询参数与分页结果
// ============================================================================

// QueryParams 通用分页与排序参数
type QueryParams struct {
	// Page 页码，从 1 开始
	Page int
	// PageSize 每页条数，0 表示不分页
	PageSize int
	// OrderBy 排序字段名，如 "id" / "start_time"
	OrderBy string
	// Order 排序方向 "asc" / "desc"
	Order string
}

// PageResult 分页结果泛型容器
type PageResult[T any] struct {
	// List 当前页数据
	List []T
	// Total 总条数
	Total int64
	// Page 当前页码
	Page int
	// PageSize 每页条数
	PageSize int
}

// ============================================================================
// 事务管理
// ============================================================================

// txKey 事务 DB 在 context 中的键类型
// 使用自定义类型避免与其他包键冲突
type txKey struct{}

// WithTx 将事务 *gorm.DB 注入 context
// 由 TransactionManager 实现内部调用，repo 通过 FromCtx 取出
func WithTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// FromTx 从 context 取出事务 *gorm.DB
// repo 实现内部调用：优先取 ctx 中的 tx，无则返回 fallback (默认 DB)
func FromTx(ctx context.Context, fallback *gorm.DB) (*gorm.DB, bool) {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok && tx != nil {
		return tx, true
	}
	return fallback, false
}

// TransactionManager 事务管理器接口
// service 层通过它开启事务，保证多表写入一致性
type TransactionManager interface {
	// Transaction 开启事务并执行 fn
	//   - fn 内的 repo 调用会自动使用同一事务
	//   - fn 返回 error 时自动回滚
	//   - fn 返回 nil 时自动提交
	//   - panic 时自动回滚
	Transaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// ============================================================================
// ScheduleEventRepository 日程事件仓储
// ============================================================================

// EventQueryOptions 日程事件过滤条件
// 用于组合查询：按所有者、时间范围、日历类型、状态等多维度过滤
type EventQueryOptions struct {
	// OwnerID 按所有者过滤，0 表示不限
	OwnerID uint
	// CalendarType 按日历类型过滤，空表示不限
	CalendarType string
	// Status 按事件状态过滤，空表示不限
	Status string
	// StartTimeFrom 开始时间下限（含）
	StartTimeFrom time.Time
	// StartTimeTo 开始时间上限（含）
	StartTimeTo time.Time
	// MeetingID 按关联会议 ID 过滤，nil 表示不限
	MeetingID *uint
	// ExcludeCancelled 是否排除已取消事件
	ExcludeCancelled bool
}

// ScheduleEventRepository 日程事件仓储接口
type ScheduleEventRepository interface {
	// ----- 基础 CRUD -----

	// Create 创建日程事件
	Create(ctx context.Context, event *model.ScheduleEvent) error

	// Update 更新日程事件（全字段更新，依赖乐观锁 Version 字段）
	// 返回 ErrConcurrent 表示版本号不匹配
	Update(ctx context.Context, event *model.ScheduleEvent) error

	// Delete 软删除日程事件
	Delete(ctx context.Context, id uint) error

	// GetByID 按主键查询事件
	// 返回 ErrNotFound 表示不存在
	GetByID(ctx context.Context, id uint) (*model.ScheduleEvent, error)

	// ----- 列表查询 -----

	// List 按组合条件分页查询
	List(ctx context.Context, opts EventQueryOptions, params QueryParams) (*PageResult[*model.ScheduleEvent], error)

	// ListByOwner 查询某用户拥有的日程（个人 + 自己创建的共享）
	ListByOwner(ctx context.Context, ownerID uint, params QueryParams) (*PageResult[*model.ScheduleEvent], error)

	// ListVisibleByUser 查询某用户可见的所有日程
	//   包含: 自己拥有的 + 被共享给自己的 (ShareScope=user 且在 ShareTargetIDs 中)
	ListVisibleByUser(ctx context.Context, userID uint, params QueryParams) (*PageResult[*model.ScheduleEvent], error)

	// ListByAttendee 查询某用户作为参与人加入的日程
	ListByAttendee(ctx context.Context, userID uint, params QueryParams) (*PageResult[*model.ScheduleEvent], error)

	// ListByTimeRange 按时间范围查询指定用户的日程
	//   用于日/周/月视图渲染，返回与 [from, to] 区间有交集的事件
	//   含重复事件的展开由 service 层根据 RRULE 计算，repo 仅返回母事件
	ListByTimeRange(ctx context.Context, userID uint, from, to time.Time) ([]*model.ScheduleEvent, error)

	// ----- 字段更新 -----

	// UpdateStatus 更新事件状态
	UpdateStatus(ctx context.Context, id uint, status string) error

	// UpdateMeetingID 关联/取消关联会议 (meetingID=nil 表示取消关联)
	UpdateMeetingID(ctx context.Context, eventID uint, meetingID *uint) error

	// ----- 反查 -----

	// GetByMeetingID 通过会议 ID 反查关联的日程事件
	// 返回 ErrNotFound 表示无关联
	GetByMeetingID(ctx context.Context, meetingID uint) (*model.ScheduleEvent, error)

	// ListExceptions 查询某重复母事件的"例外"事件列表
	// 即 RecurrenceID == recurrenceID 的所有记录
	ListExceptions(ctx context.Context, recurrenceID uint) ([]*model.ScheduleEvent, error)
}

// ============================================================================
// ScheduleReminderRepository 提醒仓储
// ============================================================================

// ReminderQueryOptions 提醒过滤条件
type ReminderQueryOptions struct {
	// EventID 按事件过滤，0 表示不限
	EventID uint
	// UserID 按接收人过滤，0 表示不限
	UserID uint
	// Status 按状态过滤，空表示不限
	Status string
	// TriggerBefore 查询触发时间早于此值的待触发提醒（调度器扫描用）
	// 0 表示不限
	TriggerBefore time.Time
}

// ScheduleReminderRepository 提醒仓储接口
type ScheduleReminderRepository interface {
	// ----- 基础 CRUD -----

	Create(ctx context.Context, reminder *model.ScheduleReminder) error
	Update(ctx context.Context, reminder *model.ScheduleReminder) error
	Delete(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*model.ScheduleReminder, error)

	// ----- 列表查询 -----

	// ListByEvent 查询某事件的所有提醒
	ListByEvent(ctx context.Context, eventID uint) ([]*model.ScheduleReminder, error)

	// ListByUser 查询某用户的所有提醒
	ListByUser(ctx context.Context, userID uint, params QueryParams) (*PageResult[*model.ScheduleReminder], error)

	// List 按组合条件查询
	List(ctx context.Context, opts ReminderQueryOptions, params QueryParams) (*PageResult[*model.ScheduleReminder], error)

	// ListPendingToFire 调度器扫描用：查询待触发且到达触发时间的提醒
	//   条件: status=pending AND remind_time <= before
	//   排序: remind_time ASC (早到先触发)
	ListPendingToFire(ctx context.Context, before time.Time, limit int) ([]*model.ScheduleReminder, error)

	// ----- 状态更新 -----

	// UpdateStatus 更新提醒状态
	UpdateStatus(ctx context.Context, id uint, status string) error

	// MarkTriggered 标记为已触发，记录实际触发时间
	MarkTriggered(ctx context.Context, id uint, at time.Time) error

	// Snooze 推迟提醒：更新状态为 snoozed，并根据 snoozeMinutes 重新计算 remind_time
	//   返回更新后的提醒（含新的 RemindTime），便于调度器重新入队
	Snooze(ctx context.Context, id uint, snoozeMinutes int) (*model.ScheduleReminder, error)

	// CancelByEvent 批量取消某事件下的所有提醒
	//   事件取消/删除时联动调用
	CancelByEvent(ctx context.Context, eventID uint) error
}

// ============================================================================
// ScheduleEventAttendeeRepository 参与人关联仓储
// ============================================================================

// AttendeeQueryOptions 参与人过滤条件
type AttendeeQueryOptions struct {
	// EventID 按事件过滤，0 表示不限
	EventID uint
	// UserID 按用户过滤，0 表示不限
	UserID uint
	// Role 按角色过滤，空表示不限
	Role string
	// ResponseStatus 按响应状态过滤，空表示不限
	ResponseStatus string
}

// ScheduleEventAttendeeRepository 参与人关联仓储接口
type ScheduleEventAttendeeRepository interface {
	// ----- 基础 CRUD -----

	Create(ctx context.Context, attendee *model.ScheduleEventAttendee) error

	// BatchCreate 批量添加参与人（事件创建时一次添加多人）
	//   在事务内执行，全部成功或全部回滚
	BatchCreate(ctx context.Context, attendees []*model.ScheduleEventAttendee) error

	Update(ctx context.Context, attendee *model.ScheduleEventAttendee) error
	Delete(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*model.ScheduleEventAttendee, error)

	// ----- 唯一性校验 -----

	// GetByEventAndUser 查询某事件中某参与人记录
	//   返回 ErrNotFound 表示该用户未参与此事件
	//   用于：判断用户是否已是参与人、获取其响应状态
	GetByEventAndUser(ctx context.Context, eventID, userID uint) (*model.ScheduleEventAttendee, error)

	// ----- 列表查询 -----

	// ListByEvent 查询某事件的全部参与人
	ListByEvent(ctx context.Context, eventID uint) ([]*model.ScheduleEventAttendee, error)

	// ListByUser 查询某用户参与的全部事件关联记录
	ListByUser(ctx context.Context, userID uint, params QueryParams) (*PageResult[*model.ScheduleEventAttendee], error)

	// List 按组合条件查询
	List(ctx context.Context, opts AttendeeQueryOptions, params QueryParams) (*PageResult[*model.ScheduleEventAttendee], error)

	// ----- 响应状态更新 -----

	// UpdateResponse 更新参与人响应状态并记录响应时间
	//   status: accepted / declined / tentative
	UpdateResponse(ctx context.Context, id uint, status string, respondedAt time.Time) error

	// ----- 批量操作 -----

	// RemoveByEvent 移除某事件下除组织者外的全部参与人
	//   事件重建参与人列表时调用
	RemoveByEvent(ctx context.Context, eventID uint) error

	// CountByEvent 统计某事件的参与人数（按角色/状态可选过滤）
	CountByEvent(ctx context.Context, eventID uint, role string) (int64, error)
}
