// Package repo 数据访问层 GORM 实现 - 提醒仓储
// 日程管理模块 (Schedule Module) - 对应 GitHub Issue #78
//
// 本文件实现 repo.ScheduleReminderRepository 接口
package repo

import (
	"context"
	"time"

	"schedule-service/internal/model"

	"gorm.io/gorm"
)

// ============================================================================
// ScheduleReminderRepository GORM 实现
// ============================================================================

// gormScheduleReminderRepository 提醒仓储实现
type gormScheduleReminderRepository struct {
	db *gorm.DB
}

// NewGormScheduleReminderRepository 构造提醒仓储
func NewGormScheduleReminderRepository(db *gorm.DB) ScheduleReminderRepository {
	return &gormScheduleReminderRepository{db: db}
}

// db 取出事务 DB 或默认 DB
func (r *gormScheduleReminderRepository) db(ctx context.Context) *gorm.DB {
	return DBFromCtx(ctx, r.db)
}

// reminderOrderWhitelist 提醒允许排序字段白名单
var reminderOrderWhitelist = map[string]bool{
	"id":           true,
	"remind_time":  true,
	"remind_offset": true,
	"created_at":   true,
	"updated_at":   true,
	"status":       true,
}

// buildReminderQuery 构建提醒查询条件 Scope
func buildReminderQuery(opts ReminderQueryOptions) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if opts.EventID != 0 {
			db = db.Where("event_id = ?", opts.EventID)
		}
		if opts.UserID != 0 {
			db = db.Where("user_id = ?", opts.UserID)
		}
		if opts.Status != "" {
			db = db.Where("status = ?", opts.Status)
		}
		if !opts.TriggerBefore.IsZero() {
			db = db.Where("remind_time <= ?", opts.TriggerBefore)
		}
		return db
	}
}

// ----------------------------------------------------------------------------
// 基础 CRUD
// ----------------------------------------------------------------------------

// Create 创建提醒
func (r *gormScheduleReminderRepository) Create(ctx context.Context, reminder *model.ScheduleReminder) error {
	if err := r.db(ctx).Create(reminder).Error; err != nil {
		return translateGormErr(err)
	}
	return nil
}

// Update 更新提醒
func (r *gormScheduleReminderRepository) Update(ctx context.Context, reminder *model.ScheduleReminder) error {
	result := r.db(ctx).
		Model(&model.ScheduleReminder{}).
		Where("id = ?", reminder.ID).
		Updates(map[string]any{
			"remind_offset":   reminder.RemindOffset,
			"remind_method":    reminder.RemindMethod,
			"remind_time":     reminder.RemindTime,
			"status":          reminder.Status,
			"snooze_minutes":  reminder.SnoozeMinutes,
			"triggered_at":    reminder.TriggeredAt,
		})
	if result.Error != nil {
		return translateGormErr(result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete 删除提醒
func (r *gormScheduleReminderRepository) Delete(ctx context.Context, id uint) error {
	result := r.db(ctx).
		Where("id = ?", id).
		Delete(&model.ScheduleReminder{})
	if result.Error != nil {
		return translateGormErr(result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// GetByID 查询提醒详情
func (r *gormScheduleReminderRepository) GetByID(ctx context.Context, id uint) (*model.ScheduleReminder, error) {
	var reminder model.ScheduleReminder
	err := r.db(ctx).First(&reminder, id).Error
	if err != nil {
		return nil, translateGormErr(err)
	}
	return &reminder, nil
}

// ----------------------------------------------------------------------------
// 列表查询
// ----------------------------------------------------------------------------

// ListByEvent 查询某事件的所有提醒
func (r *gormScheduleReminderRepository) ListByEvent(ctx context.Context, eventID uint) ([]*model.ScheduleReminder, error) {
	var reminders []*model.ScheduleReminder
	err := r.db(ctx).
		Where("event_id = ?", eventID).
		Order("remind_time ASC").
		Find(&reminders).Error
	if err != nil {
		return nil, translateGormErr(err)
	}
	return reminders, nil
}

// ListByUser 查询某用户的所有提醒
func (r *gormScheduleReminderRepository) ListByUser(ctx context.Context, userID uint, params QueryParams) (*PageResult[*model.ScheduleReminder], error) {
	return r.List(ctx, ReminderQueryOptions{
		UserID: userID,
	}, params)
}

// List 按组合条件查询
func (r *gormScheduleReminderRepository) List(ctx context.Context, opts ReminderQueryOptions, params QueryParams) (*PageResult[*model.ScheduleReminder], error) {
	var (
		total     int64
		reminders []*model.ScheduleReminder
	)
	q := r.db(ctx).Model(&model.ScheduleReminder{}).Scopes(buildReminderQuery(opts))

	if err := q.Count(&total).Error; err != nil {
		return nil, translateGormErr(err)
	}
	q = applyOrder(q, params.OrderBy, params.Order, reminderOrderWhitelist)
	q = q.Scopes(paginate(params.Page, params.PageSize))
	if err := q.Find(&reminders).Error; err != nil {
		return nil, translateGormErr(err)
	}
	return &PageResult[*model.ScheduleReminder]{
		List:     reminders,
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
	}, nil
}

// ListPendingToFire 调度器扫描用：查询待触发且到达触发时间的提醒
//   条件: status=pending AND remind_time <= before
//   排序: remind_time ASC (早到先触发)
//   limit 限制单次扫描数量，避免一次拉取过多
func (r *gormScheduleReminderRepository) ListPendingToFire(ctx context.Context, before time.Time, limit int) ([]*model.ScheduleReminder, error) {
	if limit <= 0 {
		limit = 100 // 默认上限
	}
	if limit > 1000 {
		limit = 1000 // 上限保护
	}
	var reminders []*model.ScheduleReminder
	err := r.db(ctx).
		Where("status = ? AND remind_time <= ?", "pending", before).
		Order("remind_time ASC").
		Limit(limit).
		Find(&reminders).Error
	if err != nil {
		return nil, translateGormErr(err)
	}
	return reminders, nil
}

// ----------------------------------------------------------------------------
// 状态更新
// ----------------------------------------------------------------------------

// UpdateStatus 更新提醒状态
func (r *gormScheduleReminderRepository) UpdateStatus(ctx context.Context, id uint, status string) error {
	result := r.db(ctx).
		Model(&model.ScheduleReminder{}).
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

// MarkTriggered 标记为已触发，记录实际触发时间
//   使用乐观更新：仅当 status=pending 或 status=snoozed 时才更新
//   防止重复触发
func (r *gormScheduleReminderRepository) MarkTriggered(ctx context.Context, id uint, at time.Time) error {
	result := r.db(ctx).
		Model(&model.ScheduleReminder{}).
		Where("id = ? AND status IN ?", id, []string{"pending", "snoozed"}).
		Updates(map[string]any{
			"status":      "triggered",
			"triggered_at": at,
		})
	if result.Error != nil {
		return translateGormErr(result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// Snooze 推迟提醒：更新状态为 snoozed，并根据 snoozeMinutes 重新计算 remind_time
//   返回更新后的提醒（含新的 RemindTime），便于调度器重新入队
func (r *gormScheduleReminderRepository) Snooze(ctx context.Context, id uint, snoozeMinutes int) (*model.ScheduleReminder, error) {
	// 先查当前提醒
	var reminder model.ScheduleReminder
	if err := r.db(ctx).First(&reminder, id).Error; err != nil {
		return nil, translateGormErr(err)
	}

	// 校验当前状态：仅 pending 或 snoozed 可推迟
	if reminder.Status != "pending" && reminder.Status != "snoozed" {
		return nil, ErrNotFound // 状态不允许推迟视为"找不到"可操作记录
	}

	// 计算新的触发时间：now + snoozeMinutes
	newRemindTime := time.Now().Add(time.Duration(snoozeMinutes) * time.Minute)
	// 累计推迟时长
	totalSnooze := reminder.SnoozeMinutes + snoozeMinutes

	// 更新
	result := r.db(ctx).
		Model(&model.ScheduleReminder{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":         "snoozed",
			"snooze_minutes":  totalSnooze,
			"remind_time":     newRemindTime,
			"triggered_at":   nil, // 清空旧的触发时间
		})
	if result.Error != nil {
		return nil, translateGormErr(result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, ErrNotFound
	}

	// 同步内存对象
	reminder.Status = "snoozed"
	reminder.SnoozeMinutes = totalSnooze
	reminder.RemindTime = newRemindTime
	reminder.TriggeredAt = nil
	return &reminder, nil
}

// CancelByEvent 批量取消某事件下的所有提醒
//   事件取消/删除时联动调用
//   仅取消 pending 和 snoozed 状态，已触发的不动
func (r *gormScheduleReminderRepository) CancelByEvent(ctx context.Context, eventID uint) error {
	result := r.db(ctx).
		Model(&model.ScheduleReminder{}).
		Where("event_id = ? AND status IN ?", eventID, []string{"pending", "snoozed"}).
		Update("status", "canceled")
	if result.Error != nil {
		return translateGormErr(result.Error)
	}
	// 即便影响 0 行也视为成功（事件可能本就无 pending 提醒）
	return nil
}
