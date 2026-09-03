package repo

import (
	"context"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/audit/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuditRepo 审计日志数据访问接口
type AuditRepo interface {
	// Create 创建审计日志（异步写入由中间件调用）
	Create(ctx context.Context, entry *model.AuditLog) error

	// GetByID 根据 ID 查询单条审计日志
	GetByID(ctx context.Context, id uuid.UUID) (*model.AuditLog, error)

	// List 分页查询审计日志列表
	List(ctx context.Context, req *ListParams) ([]model.AuditLog, int64, error)

	// Export 查询符合条件的审计日志（不分页，用于导出）
	Export(ctx context.Context, req *ListParams) ([]model.AuditLog, error)

	// ArchiveBefore 将指定时间之前的日志标记为已归档
	ArchiveBefore(ctx context.Context, before time.Time) (int64, error)

	// DeleteArchived 删除已归档且超过指定天数的日志
	DeleteArchived(ctx context.Context, before time.Time) (int64, error)

	// CreateArchive 创建归档记录
	CreateArchive(ctx context.Context, archive *model.AuditLogArchive) error
}

// ListParams 列表查询参数
type ListParams struct {
	Page       int
	PageSize   int
	Username   string
	Operation  string
	Method     string
	Path       string
	IP         string
	RequestID  string
	StatusMin  *int
	StatusMax  *int
	StartTime  *time.Time
	EndTime    *time.Time
	IsArchived *bool
}

type auditRepo struct {
	db *gorm.DB
}

// NewAuditRepo 创建审计日志 Repo
func NewAuditRepo(db *gorm.DB) AuditRepo {
	return &auditRepo{db: db}
}

func (r *auditRepo) buildQuery(ctx context.Context, req *ListParams) *gorm.DB {
	query := r.db.WithContext(ctx).Model(&model.AuditLog{})

	if req.Username != "" {
		query = query.Where("username LIKE ?", "%"+req.Username+"%")
	}
	if req.Operation != "" {
		query = query.Where("operation LIKE ?", "%"+req.Operation+"%")
	}
	if req.Method != "" {
		query = query.Where("method = ?", req.Method)
	}
	if req.Path != "" {
		query = query.Where("path LIKE ?", "%"+req.Path+"%")
	}
	if req.IP != "" {
		query = query.Where("ip = ?", req.IP)
	}
	if req.RequestID != "" {
		query = query.Where("request_id = ?", req.RequestID)
	}
	if req.StatusMin != nil {
		query = query.Where("response_status >= ?", *req.StatusMin)
	}
	if req.StatusMax != nil {
		query = query.Where("response_status <= ?", *req.StatusMax)
	}
	if req.StartTime != nil {
		query = query.Where("created_at >= ?", *req.StartTime)
	}
	if req.EndTime != nil {
		query = query.Where("created_at <= ?", *req.EndTime)
	}
	if req.IsArchived != nil {
		query = query.Where("is_archived = ?", *req.IsArchived)
	}

	return query
}

func (r *auditRepo) Create(ctx context.Context, entry *model.AuditLog) error {
	return r.db.WithContext(ctx).Create(entry).Error
}

func (r *auditRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.AuditLog, error) {
	var log model.AuditLog
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&log).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &log, err
}

func (r *auditRepo) List(ctx context.Context, req *ListParams) ([]model.AuditLog, int64, error) {
	var logs []model.AuditLog
	var total int64

	query := r.buildQuery(ctx, req)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (req.Page - 1) * req.PageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(req.PageSize).Find(&logs).Error

	return logs, total, err
}

func (r *auditRepo) Export(ctx context.Context, req *ListParams) ([]model.AuditLog, error) {
	var logs []model.AuditLog

	// 导出限制：最多导出 10000 条
	query := r.buildQuery(ctx, req)
	err := query.Order("created_at DESC").Limit(10000).Find(&logs).Error

	return logs, err
}

func (r *auditRepo) ArchiveBefore(ctx context.Context, before time.Time) (int64, error) {
	result := r.db.WithContext(ctx).Model(&model.AuditLog{}).
		Where("created_at < ? AND is_archived = false", before).
		Update("is_archived", true)
	return result.RowsAffected, result.Error
}

func (r *auditRepo) DeleteArchived(ctx context.Context, before time.Time) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("created_at < ? AND is_archived = true", before).
		Delete(&model.AuditLog{})
	return result.RowsAffected, result.Error
}

func (r *auditRepo) CreateArchive(ctx context.Context, archive *model.AuditLogArchive) error {
	return r.db.WithContext(ctx).Create(archive).Error
}
