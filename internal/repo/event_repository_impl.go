// Package repo 数据访问层 GORM 实现 - 日程事件仓储
// 日程管理模块 (Schedule Module) - 对应 GitHub Issue #78
//
// 本文件实现 repo.ScheduleEventRepository 接口
// 所有方法自动从 context 取出事务 DB，无事务则用默认 DB
package repo

import (
	"context"
	"strconv"
	"time"

	"schedule-service/internal/model"

	"gorm.io/gorm"
)

// ============================================================================
// ScheduleEventRepository GORM 实现
// ============================================================================

// gormScheduleEventRepository 日程事件仓储实现
type gormScheduleEventRepository struct {
	db *gorm.DB
}

// NewGormScheduleEventRepository 构造日程事件仓储
//   db: 默认 *gorm.DB（无事务时使用）
func NewGormScheduleEventRepository(db *gorm.DB) ScheduleEventRepository {
	return &gormScheduleEventRepository{db: db}
}

// db 取出事务 DB 或默认 DB
func (r *gormScheduleEventRepository) db(ctx context.Context) *gorm.DB {
	return DBFromCtx(ctx, r.db)
}

// ----------------------------------------------------------------------------
// 基础 CRUD
// ----------------------------------------------------------------------------

// Create 创建日程事件
func (r *gormScheduleEventRepository) Create(ctx context.Context, event *model.ScheduleEvent) error {
	if err := r.db(ctx).Create(event).Error; err != nil {
		return translateGormErr(err)
	}
	return nil
}

// Update 更新日程事件（全字段更新，依赖乐观锁 Version 字段）
//   实现：基于 ID + Version 条件更新，若影响行数 0 则返回 ErrConcurrent
func (r *gormScheduleEventRepository) Update(ctx context.Context, event *model.ScheduleEvent) error {
	result := r.db(ctx).
		Model(&model.ScheduleEvent{}).
		Where("id = ? AND version = ?", event.ID, event.Version).
		Updates(map[string]any{
			"title":            event.Title,
			"description":      event.Description,
			"start_time":       event.StartTime,
			"end_time":         event.EndTime,
			"all_day":          event.AllDay,
			"location":         event.Location,
			"calendar_type":    event.CalendarType,
			"share_scope":      event.ShareScope,
			"share_target_ids": event.ShareTargetIDs,
			"recurrence_rule":  event.RecurrenceRule,
			"status":           event.Status,
			"meeting_id":       event.MeetingID,
			"updater_id":       event.UpdaterID,
			"version":          gorm.Expr("version + ?", 1),
		})
	if result.Error != nil {
		return translateGormErr(result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrConcurrent
	}
	// 同步内存对象的 Version +1
	event.Version++
	return nil
}

// Delete 软删除日程事件
//   gorm.DeletedAt 字段会自动处理软删除
func (r *gormScheduleEventRepository) Delete(ctx context.Context, id uint) error {
	result := r.db(ctx).
		Where("id = ?", id).
		Delete(&model.ScheduleEvent{})
	if result.Error != nil {
		return translateGormErr(result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// GetByID 按主键查询事件
func (r *gormScheduleEventRepository) GetByID(ctx context.Context, id uint) (*model.ScheduleEvent, error) {
	var event model.ScheduleEvent
	err := r.db(ctx).First(&event, id).Error
	if err != nil {
		return nil, translateGormErr(err)
	}
	return &event, nil
}

// ----------------------------------------------------------------------------
// 列表查询
// ----------------------------------------------------------------------------

// eventOrderWhitelist 允许排序的字段白名单（防 SQL 注入）
var eventOrderWhitelist = map[string]bool{
	"id":         true,
	"start_time": true,
	"end_time":   true,
	"created_at": true,
	"updated_at": true,
	"status":     true,
}

// buildEventQuery 构建事件查询条件 Scope
func buildEventQuery(opts EventQueryOptions) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if opts.OwnerID != 0 {
			db = db.Where("owner_id = ?", opts.OwnerID)
		}
		if opts.CalendarType != "" {
			db = db.Where("calendar_type = ?", opts.CalendarType)
		}
		if opts.Status != "" {
			db = db.Where("status = ?", opts.Status)
		}
		if !opts.StartTimeFrom.IsZero() {
			db = db.Where("start_time >= ?", opts.StartTimeFrom)
		}
		if !opts.StartTimeTo.IsZero() {
			db = db.Where("start_time <= ?", opts.StartTimeTo)
		}
		if opts.MeetingID != nil {
			db = db.Where("meeting_id = ?", *opts.MeetingID)
		}
		if opts.ExcludeCancelled {
			db = db.Where("status <> ?", "cancelled")
		}
		return db
	}
}

// List 按组合条件分页查询
func (r *gormScheduleEventRepository) List(ctx context.Context, opts EventQueryOptions, params QueryParams) (*PageResult[*model.ScheduleEvent], error) {
	var (
		total  int64
		events []*model.ScheduleEvent
	)
	q := r.db(ctx).Model(&model.ScheduleEvent{}).Scopes(buildEventQuery(opts))

	// 统计总数
	if err := q.Count(&total).Error; err != nil {
		return nil, translateGormErr(err)
	}

	// 应用排序与分页
	q = applyOrder(q, params.OrderBy, params.Order, eventOrderWhitelist)
	q = q.Scopes(paginate(params.Page, params.PageSize))

	if err := q.Find(&events).Error; err != nil {
		return nil, translateGormErr(err)
	}
	return &PageResult[*model.ScheduleEvent]{
		List:     events,
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
	}, nil
}

// ListByOwner 查询某用户拥有的日程（个人 + 自己创建的共享）
func (r *gormScheduleEventRepository) ListByOwner(ctx context.Context, ownerID uint, params QueryParams) (*PageResult[*model.ScheduleEvent], error) {
	return r.List(ctx, EventQueryOptions{
		OwnerID: ownerID,
	}, params)
}

// ListVisibleByUser 查询某用户可见的所有日程
//   包含: 自己拥有的 + 被共享给自己的 (ShareScope=user 且在 ShareTargetIDs 中)
//   实现: 用 JSON_CONTAINS 判断 ShareTargetIDs 是否包含 userID
//   注: 不同 DB 的 JSON 函数语法不同，此处用通用 LIKE 兜底（精确匹配用 MySQL 的 JSON_CONTAINS）
func (r *gormScheduleEventRepository) ListVisibleByUser(ctx context.Context, userID uint, params QueryParams) (*PageResult[*model.ScheduleEvent], error) {
	var (
		total  int64
		events []*model.ScheduleEvent
	)
	// 查询条件: owner_id = userID OR (calendar_type='shared' AND share_scope='user' AND share_target_ids LIKE %userID%)
	//   share_target_ids 为 JSON 数组，如 [1001,1002]，简化用 LIKE 实现近似匹配
	//   严格实现应使用 DB 原生 JSON 函数，实现时按具体驱动替换
	uidStr := strconv.FormatUint(uint64(userID), 10)
	q := r.db(ctx).Model(&model.ScheduleEvent{}).Where(
		"owner_id = ? OR (calendar_type = 'shared' AND share_scope = 'user' AND share_target_ids LIKE ?)",
		userID, "%"+uidStr+"%",
	)

	if err := q.Count(&total).Error; err != nil {
		return nil, translateGormErr(err)
	}
	q = applyOrder(q, params.OrderBy, params.Order, eventOrderWhitelist)
	q = q.Scopes(paginate(params.Page, params.PageSize))
	if err := q.Find(&events).Error; err != nil {
		return nil, translateGormErr(err)
	}
	return &PageResult[*model.ScheduleEvent]{
		List:     events,
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
	}, nil
}

// ListByAttendee 查询某用户作为参与人加入的日程
//   通过 JOIN schedule_event_attendee 表查询
func (r *gormScheduleEventRepository) ListByAttendee(ctx context.Context, userID uint, params QueryParams) (*PageResult[*model.ScheduleEvent], error) {
	var (
		total  int64
		events []*model.ScheduleEvent
	)
	q := r.db(ctx).
		Model(&model.ScheduleEvent{}).
		Joins("JOIN schedule_event_attendee ON schedule_event_attendee.event_id = schedule_event.id").
		Where("schedule_event_attendee.user_id = ? AND schedule_event_attendee.deleted_at IS NULL", userID)

	if err := q.Count(&total).Error; err != nil {
		return nil, translateGormErr(err)
	}
	q = applyOrder(q, params.OrderBy, params.Order, eventOrderWhitelist)
	q = q.Scopes(paginate(params.Page, params.PageSize))
	if err := q.Find(&events).Error; err != nil {
		return nil, translateGormErr(err)
	}
	return &PageResult[*model.ScheduleEvent]{
		List:     events,
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
	}, nil
}

// ListByTimeRange 按时间范围查询指定用户的日程
//   返回与 [from, to] 区间有交集的事件: start_time <= to AND end_time >= from
//   含重复事件的展开由 service 层根据 RRULE 计算，repo 仅返回母事件
func (r *gormScheduleEventRepository) ListByTimeRange(ctx context.Context, userID uint, from, to time.Time) ([]*model.ScheduleEvent, error) {
	var events []*model.ScheduleEvent
	q := r.db(ctx).
		Model(&model.ScheduleEvent{}).
		Where(
			"(owner_id = ? OR "+
				"(calendar_type = 'shared' AND share_scope = 'user' AND share_target_ids LIKE ?) OR "+
				"id IN (SELECT event_id FROM schedule_event_attendee WHERE user_id = ? AND deleted_at IS NULL))",
			userID, "%"+strconv.FormatUint(uint64(userID), 10)+"%", userID,
		).
		Where("start_time <= ? AND end_time >= ?", to, from). // 区间交集
		Where("status <> ?", "cancelled").
		Order("start_time ASC")

	if err := q.Find(&events).Error; err != nil {
		return nil, translateGormErr(err)
	}
	return events, nil
}

// ----------------------------------------------------------------------------
// 字段更新
// ----------------------------------------------------------------------------

// UpdateStatus 更新事件状态
func (r *gormScheduleEventRepository) UpdateStatus(ctx context.Context, id uint, status string) error {
	result := r.db(ctx).
		Model(&model.ScheduleEvent{}).
		Where("id = ?", id).
		Update("status", status)
	if result.Error != nil {
		return translateGormErr(result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// UpdateMeetingID 关联/取消关联会议 (meetingID=nil 表示取消关联)
func (r *gormScheduleEventRepository) UpdateMeetingID(ctx context.Context, eventID uint, meetingID *uint) error {
	result := r.db(ctx).
		Model(&model.ScheduleEvent{}).
		Where("id = ?", eventID).
		Update("meeting_id", meetingID)
	if result.Error != nil {
		return translateGormErr(result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// ----------------------------------------------------------------------------
// 反查
// ----------------------------------------------------------------------------

// GetByMeetingID 通过会议 ID 反查关联的日程事件
//   注: meeting_id 为软引用，不做唯一性约束，取第一条
func (r *gormScheduleEventRepository) GetByMeetingID(ctx context.Context, meetingID uint) (*model.ScheduleEvent, error) {
	var event model.ScheduleEvent
	err := r.db(ctx).
		Where("meeting_id = ?", meetingID).
		First(&event).Error
	if err != nil {
		return nil, translateGormErr(err)
	}
	return &event, nil
}

// ListExceptions 查询某重复母事件的"例外"事件列表
func (r *gormScheduleEventRepository) ListExceptions(ctx context.Context, recurrenceID uint) ([]*model.ScheduleEvent, error) {
	var events []*model.ScheduleEvent
	err := r.db(ctx).
		Where("recurrence_id = ?", recurrenceID).
		Order("start_time ASC").
		Find(&events).Error
	if err != nil {
		return nil, translateGormErr(err)
	}
	return events, nil
}

