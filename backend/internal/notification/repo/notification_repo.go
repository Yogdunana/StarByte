package repo

import (
	"context"

	"github.com/Yogdunana/StarByte/backend/internal/notification/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// NotificationRepo 通知数据访问接口
type NotificationRepo interface {
	// Create 创建通知
	Create(ctx context.Context, n *model.Notification) error
	// BatchCreate 批量创建通知
	BatchCreate(ctx context.Context, notifications []*model.Notification) error
	// ListByUser 查询用户通知列表（分页）
	ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int, category string, unreadOnly bool) ([]*model.Notification, int64, error)
	// GetByID 根据 ID 查询通知
	GetByID(ctx context.Context, id uuid.UUID) (*model.Notification, error)
	// MarkAsRead 标记通知为已读
	MarkAsRead(ctx context.Context, userID uuid.UUID, ids []uuid.UUID) error
	// MarkAllAsRead 标记用户全部通知为已读
	MarkAllAsRead(ctx context.Context, userID uuid.UUID, category string) error
	// Delete 删除通知
	Delete(ctx context.Context, userID uuid.UUID, id uuid.UUID) error
	// GetUnreadCount 获取用户未读通知数
	GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error)
}

type notificationRepo struct {
	db *gorm.DB
}

// NewNotificationRepo 创建通知 repo
func NewNotificationRepo(db *gorm.DB) NotificationRepo {
	return &notificationRepo{db: db}
}

func (r *notificationRepo) Create(ctx context.Context, n *model.Notification) error {
	return r.db.WithContext(ctx).Create(n).Error
}

func (r *notificationRepo) BatchCreate(ctx context.Context, notifications []*model.Notification) error {
	if len(notifications) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).CreateInBatches(notifications, 100).Error
}

func (r *notificationRepo) ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int, category string, unreadOnly bool) ([]*model.Notification, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	query := r.db.WithContext(ctx).Model(&model.Notification{}).Where("user_id = ?", userID)
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if unreadOnly {
		query = query.Where("is_read = false")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []*model.Notification
	if err := query.Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *notificationRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Notification, error) {
	var n model.Notification
	if err := r.db.WithContext(ctx).First(&n, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *notificationRepo) MarkAsRead(ctx context.Context, userID uuid.UUID, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(&model.Notification{}).
		Where("user_id = ? AND id IN ?", userID, ids).
		Update("is_read", true).Error
}

func (r *notificationRepo) MarkAllAsRead(ctx context.Context, userID uuid.UUID, category string) error {
	query := r.db.WithContext(ctx).Model(&model.Notification{}).
		Where("user_id = ? AND is_read = false", userID)
	if category != "" {
		query = query.Where("category = ?", category)
	}
	return query.Update("is_read", true).Error
}

func (r *notificationRepo) Delete(ctx context.Context, userID uuid.UUID, id uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND id = ?", userID, id).
		Delete(&model.Notification{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *notificationRepo) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Notification{}).
		Where("user_id = ? AND is_read = false", userID).
		Count(&count).Error
	return count, err
}

// ========== Template Repo ==========

// NotificationTemplateRepo 通知模板数据访问接口
type NotificationTemplateRepo interface {
	Create(ctx context.Context, t *model.NotificationTemplate) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.NotificationTemplate, error)
	GetByCode(ctx context.Context, code string) (*model.NotificationTemplate, error)
	List(ctx context.Context, page, pageSize int, keyword string) ([]*model.NotificationTemplate, int64, error)
	Update(ctx context.Context, t *model.NotificationTemplate) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type templateRepo struct {
	db *gorm.DB
}

// NewTemplateRepo 创建模板 repo
func NewTemplateRepo(db *gorm.DB) NotificationTemplateRepo {
	return &templateRepo{db: db}
}

func (r *templateRepo) Create(ctx context.Context, t *model.NotificationTemplate) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *templateRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.NotificationTemplate, error) {
	var t model.NotificationTemplate
	if err := r.db.WithContext(ctx).First(&t, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *templateRepo) GetByCode(ctx context.Context, code string) (*model.NotificationTemplate, error) {
	var t model.NotificationTemplate
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *templateRepo) List(ctx context.Context, page, pageSize int, keyword string) ([]*model.NotificationTemplate, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	query := r.db.WithContext(ctx).Model(&model.NotificationTemplate{})
	if keyword != "" {
		query = query.Where("name ILIKE ? OR code ILIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []*model.NotificationTemplate
	if err := query.Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *templateRepo) Update(ctx context.Context, t *model.NotificationTemplate) error {
	return r.db.WithContext(ctx).Save(t).Error
}

func (r *templateRepo) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&model.NotificationTemplate{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
