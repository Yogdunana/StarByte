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
	Create(ctx context.Context, entry *model.AuditLog) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.AuditLog, error)
	List(ctx context.Context, req *ListParams) ([]model.AuditLog, int64, error)
	Count(ctx context.Context, req *ListParams) (int64, error)
	Iterate(ctx context.Context, req *ListParams, batchSize int, fn func([]model.AuditLog) error) error
	DeleteBefore(ctx context.Context, before time.Time) (int64, error)
	CreateArchive(ctx context.Context, archive *model.AuditLogArchive) error
}

// ListParams 列表/导出/归档查询参数
type ListParams struct {
	Page      int
	PageSize  int
	Username  string
	UserID    *uuid.UUID
	Action    string
	Module    string
	Keyword   string
	Method    string
	IP        string
	StartTime *time.Time
	EndTime   *time.Time
}

type auditRepo struct {
	db *gorm.DB
}

func NewAuditRepo(db *gorm.DB) AuditRepo {
	return &auditRepo{db: db}
}

func (r *auditRepo) buildQuery(ctx context.Context, req *ListParams) *gorm.DB {
	query := r.db.WithContext(ctx).Model(&model.AuditLog{})
	if req == nil {
		return query
	}
	if req.Username != "" {
		query = query.Where("username LIKE ?", "%"+req.Username+"%")
	}
	if req.UserID != nil {
		query = query.Where("user_id = ?", *req.UserID)
	}
	if req.Action != "" {
		query = query.Where("action = ?", req.Action)
	}
	if req.Module != "" {
		query = query.Where("module = ?", req.Module)
	}
	if req.Method != "" {
		query = query.Where("method = ?", req.Method)
	}
	if req.IP != "" {
		query = query.Where("ip = ?", req.IP)
	}
	if req.Keyword != "" {
		like := "%" + req.Keyword + "%"
		query = query.Where("path LIKE ? OR request_params LIKE ?", like, like)
	}
	if req.StartTime != nil {
		query = query.Where("created_at >= ?", *req.StartTime)
	}
	if req.EndTime != nil {
		query = query.Where("created_at <= ?", *req.EndTime)
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
	page, pageSize := req.Page, req.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = model.DefaultListPageSize
	}
	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error
	return logs, total, err
}

func (r *auditRepo) Count(ctx context.Context, req *ListParams) (int64, error) {
	var total int64
	err := r.buildQuery(ctx, req).Count(&total).Error
	return total, err
}

func (r *auditRepo) Iterate(ctx context.Context, req *ListParams, batchSize int, fn func([]model.AuditLog) error) error {
	if batchSize <= 0 {
		batchSize = model.DefaultIterateBatch
	}
	offset := 0
	for {
		var logs []model.AuditLog
		err := r.buildQuery(ctx, req).
			Order("created_at DESC").
			Offset(offset).
			Limit(batchSize).
			Find(&logs).Error
		if err != nil {
			return err
		}
		if len(logs) == 0 {
			return nil
		}
		if err := fn(logs); err != nil {
			return err
		}
		if len(logs) < batchSize {
			return nil
		}
		offset += batchSize
	}
}

func (r *auditRepo) DeleteBefore(ctx context.Context, before time.Time) (int64, error) {
	result := r.db.WithContext(ctx).Where("created_at < ?", before).Delete(&model.AuditLog{})
	return result.RowsAffected, result.Error
}

func (r *auditRepo) CreateArchive(ctx context.Context, archive *model.AuditLogArchive) error {
	return r.db.WithContext(ctx).Create(archive).Error
}
