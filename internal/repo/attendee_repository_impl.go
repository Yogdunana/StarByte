// Package repo 数据访问层 GORM 实现 - 参与人关联仓储
// 日程管理模块 (Schedule Module) - 对应 GitHub Issue #78
//
// 本文件实现 repo.ScheduleEventAttendeeRepository 接口
package repo

import (
	"context"
	"time"

	"schedule-service/internal/model"

	"gorm.io/gorm"
)

// ============================================================================
// ScheduleEventAttendeeRepository GORM 实现
// ============================================================================

// gormScheduleEventAttendeeRepository 参与人仓储实现
type gormScheduleEventAttendeeRepository struct {
	db *gorm.DB
}

// NewGormScheduleEventAttendeeRepository 构造参与人仓储
func NewGormScheduleEventAttendeeRepository(db *gorm.DB) ScheduleEventAttendeeRepository {
	return &gormScheduleEventAttendeeRepository{db: db}
}

// db 取出事务 DB 或默认 DB
func (r *gormScheduleEventAttendeeRepository) db(ctx context.Context) *gorm.DB {
	return DBFromCtx(ctx, r.db)
}

// attendeeOrderWhitelist 参与人允许排序字段白名单
var attendeeOrderWhitelist = map[string]bool{
	"id":              true,
	"event_id":        true,
	"user_id":         true,
	"role":            true,
	"response_status": true,
	"created_at":     true,
	"updated_at":      true,
}

// buildAttendeeQuery 构建参与人查询条件 Scope
func buildAttendeeQuery(opts AttendeeQueryOptions) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if opts.EventID != 0 {
			db = db.Where("event_id = ?", opts.EventID)
		}
		if opts.UserID != 0 {
			db = db.Where("user_id = ?", opts.UserID)
		}
		if opts.Role != "" {
			db = db.Where("role = ?", opts.Role)
		}
		if opts.ResponseStatus != "" {
			db = db.Where("response_status = ?", opts.ResponseStatus)
		}
		return db
	}
}

// ----------------------------------------------------------------------------
// 基础 CRUD
// ----------------------------------------------------------------------------

// Create 添加参与人
func (r *gormScheduleEventAttendeeRepository) Create(ctx context.Context, attendee *model.ScheduleEventAttendee) error {
	if err := r.db(ctx).Create(attendee).Error; err != nil {
		return translateGormErr(err)
	}
	return nil
}

// BatchCreate 批量添加参与人
//   在事务内执行，全部成功或全部回滚
//   使用 CreateInBatches 单条 SQL 批量插入
func (r *gormScheduleEventAttendeeRepository) BatchCreate(ctx context.Context, attendees []*model.ScheduleEventAttendee) error {
	if len(attendees) == 0 {
		return nil
	}
	// batchSize=100 分批插入，单条 SQL 不超过参数上限
	if err := r.db(ctx).
		CreateInBatches(attendees, 100).Error; err != nil {
		return translateGormErr(err)
	}
	return nil
}

// Update 更新参与人记录
func (r *gormScheduleEventAttendeeRepository) Update(ctx context.Context, attendee *model.ScheduleEventAttendee) error {
	result := r.db(ctx).
		Model(&model.ScheduleEventAttendee{}).
		Where("id = ?", attendee.ID).
		Updates(map[string]any{
			"role":            attendee.Role,
			"response_status": attendee.ResponseStatus,
			"responded_at":    attendee.RespondedAt,
		})
	if result.Error != nil {
		return translateGormErr(result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete 软删除参与人记录
func (r *gormScheduleEventAttendeeRepository) Delete(ctx context.Context, id uint) error {
	result := r.db(ctx).
		Where("id = ?", id).
		Delete(&model.ScheduleEventAttendee{})
	if result.Error != nil {
		return translateGormErr(result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// GetByID 查询参与人记录详情
func (r *gormScheduleEventAttendeeRepository) GetByID(ctx context.Context, id uint) (*model.ScheduleEventAttendee, error) {
	var attendee model.ScheduleEventAttendee
	err := r.db(ctx).First(&attendee, id).Error
	if err != nil {
		return nil, translateGormErr(err)
	}
	return &attendee, nil
}

// ----------------------------------------------------------------------------
// 唯一性校验
// ----------------------------------------------------------------------------

// GetByEventAndUser 查询某事件中某参与人记录
//   用于判断用户是否已是参与人、获取其响应状态
func (r *gormScheduleEventAttendeeRepository) GetByEventAndUser(ctx context.Context, eventID, userID uint) (*model.ScheduleEventAttendee, error) {
	var attendee model.ScheduleEventAttendee
	err := r.db(ctx).
		Where("event_id = ? AND user_id = ?", eventID, userID).
		First(&attendee).Error
	if err != nil {
		return nil, translateGormErr(err)
	}
	return &attendee, nil
}

// ----------------------------------------------------------------------------
// 列表查询
// ----------------------------------------------------------------------------

// ListByEvent 查询某事件的全部参与人
func (r *gormScheduleEventAttendeeRepository) ListByEvent(ctx context.Context, eventID uint) ([]*model.ScheduleEventAttendee, error) {
	var attendees []*model.ScheduleEventAttendee
	err := r.db(ctx).
		Where("event_id = ?", eventID).
		Order("role ASC, created_at ASC"). // 组织者优先，同角色按加入时间
		Find(&attendees).Error
	if err != nil {
		return nil, translateGormErr(err)
	}
	return attendees, nil
}

// ListByUser 查询某用户参与的全部事件关联记录
func (r *gormScheduleEventAttendeeRepository) ListByUser(ctx context.Context, userID uint, params QueryParams) (*PageResult[*model.ScheduleEventAttendee], error) {
	return r.List(ctx, AttendeeQueryOptions{
		UserID: userID,
	}, params)
}

// List 按组合条件查询
func (r *gormScheduleEventAttendeeRepository) List(ctx context.Context, opts AttendeeQueryOptions, params QueryParams) (*PageResult[*model.ScheduleEventAttendee], error) {
	var (
		total     int64
		attendees []*model.ScheduleEventAttendee
	)
	q := r.db(ctx).Model(&model.ScheduleEventAttendee{}).Scopes(buildAttendeeQuery(opts))

	if err := q.Count(&total).Error; err != nil {
		return nil, translateGormErr(err)
	}
	q = applyOrder(q, params.OrderBy, params.Order, attendeeOrderWhitelist)
	q = q.Scopes(paginate(params.Page, params.PageSize))
	if err := q.Find(&attendees).Error; err != nil {
		return nil, translateGormErr(err)
	}
	return &PageResult[*model.ScheduleEventAttendee]{
		List:     attendees,
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
	}, nil
}

// ----------------------------------------------------------------------------
// 响应状态更新
// ----------------------------------------------------------------------------

// UpdateResponse 更新参与人响应状态并记录响应时间
//   status: accepted / declined / tentative
func (r *gormScheduleEventAttendeeRepository) UpdateResponse(ctx context.Context, id uint, status string, respondedAt time.Time) error {
	result := r.db(ctx).
		Model(&model.ScheduleEventAttendee{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"response_status": status,
			"responded_at":     respondedAt,
		})
	if result.Error != nil {
		return translateGormErr(result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// ----------------------------------------------------------------------------
// 批量操作
// ----------------------------------------------------------------------------

// RemoveByEvent 移除某事件下除组织者外的全部参与人
//   事件重建参与人列表时调用
//   组织者不可被移除（业务规则在 service 层校验，repo 仅做条件过滤）
func (r *gormScheduleEventAttendeeRepository) RemoveByEvent(ctx context.Context, eventID uint) error {
	result := r.db(ctx).
		Where("event_id = ? AND role <> ?", eventID, "organizer").
		Delete(&model.ScheduleEventAttendee{})
	if result.Error != nil {
		return translateGormErr(result.Error)
	}
	// 影响 0 行视为成功（事件可能本就无其他参与人）
	return nil
}

// CountByEvent 统计某事件的参与人数（按角色可选过滤）
//   role 为空表示统计全部角色
func (r *gormScheduleEventAttendeeRepository) CountByEvent(ctx context.Context, eventID uint, role string) (int64, error) {
	var count int64
	q := r.db(ctx).Model(&model.ScheduleEventAttendee{}).Where("event_id = ?", eventID)
	if role != "" {
		q = q.Where("role = ?", role)
	}
	if err := q.Count(&count).Error; err != nil {
		return 0, translateGormErr(err)
	}
	return count, nil
}
