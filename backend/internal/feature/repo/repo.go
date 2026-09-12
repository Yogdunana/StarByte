package repo

import (
	"context"
	"strings"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/feature/dto"
	"github.com/Yogdunana/StarByte/backend/internal/feature/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	List(ctx context.Context, q dto.ListQuery) ([]model.Flag, int64, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Flag, error)
	GetByKey(ctx context.Context, key string) (*model.Flag, error)
	ListAll(ctx context.Context) ([]model.Flag, error)
	Create(ctx context.Context, flag *model.Flag) error
	Update(ctx context.Context, flag *model.Flag) error
	Delete(ctx context.Context, id uuid.UUID) error
	CreateAudit(ctx context.Context, row *model.Audit) error
	ListAudits(ctx context.Context, q dto.AuditQuery) ([]model.Audit, int64, error)
	LatestMutableAudit(ctx context.Context, flagID uuid.UUID) (*model.Audit, error)
	CreateExposure(ctx context.Context, row *model.Exposure) error
	SummarizeExposures(ctx context.Context, flagKey string, since time.Time) ([]model.ExposureBucket, error)
}

type repo struct{ db *gorm.DB }

func New(db *gorm.DB) Repository { return &repo{db: db} }

func (r *repo) List(ctx context.Context, q dto.ListQuery) ([]model.Flag, int64, error) {
	page, size := normalizePage(q.Page, q.PageSize)
	query := r.db.WithContext(ctx).Model(&model.Flag{})
	if kw := strings.TrimSpace(q.Keyword); kw != "" {
		like := "%" + kw + "%"
		query = query.Where("flag_key ILIKE ? OR name ILIKE ?", like, like)
	}
	if g := strings.TrimSpace(q.GroupName); g != "" {
		query = query.Where("group_name = ?", g)
	}
	if q.Enabled != nil {
		query = query.Where("enabled = ?", *q.Enabled)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.Flag
	err := query.Order("priority DESC, updated_at DESC").Offset((page - 1) * size).Limit(size).Find(&rows).Error
	return rows, total, err
}

func (r *repo) GetByID(ctx context.Context, id uuid.UUID) (*model.Flag, error) {
	var row model.Flag
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&row).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &row, err
}

func (r *repo) GetByKey(ctx context.Context, key string) (*model.Flag, error) {
	var row model.Flag
	err := r.db.WithContext(ctx).Where("flag_key = ?", key).First(&row).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &row, err
}

func (r *repo) ListAll(ctx context.Context) ([]model.Flag, error) {
	var rows []model.Flag
	err := r.db.WithContext(ctx).Order("priority DESC, flag_key").Find(&rows).Error
	return rows, err
}

func (r *repo) Create(ctx context.Context, flag *model.Flag) error {
	return r.db.WithContext(ctx).Create(flag).Error
}

func (r *repo) Update(ctx context.Context, flag *model.Flag) error {
	return r.db.WithContext(ctx).Save(flag).Error
}

func (r *repo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Flag{}, "id = ?", id).Error
}

func (r *repo) CreateAudit(ctx context.Context, row *model.Audit) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *repo) ListAudits(ctx context.Context, q dto.AuditQuery) ([]model.Audit, int64, error) {
	page, size := normalizePage(q.Page, q.PageSize)
	query := r.db.WithContext(ctx).Model(&model.Audit{})
	if k := strings.TrimSpace(q.FlagKey); k != "" {
		query = query.Where("flag_key = ?", k)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.Audit
	err := query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&rows).Error
	return rows, total, err
}

func (r *repo) LatestMutableAudit(ctx context.Context, flagID uuid.UUID) (*model.Audit, error) {
	var row model.Audit
	err := r.db.WithContext(ctx).
		Where("flag_id = ? AND before_json IS NOT NULL AND action IN ?", flagID, []string{
			model.ActionUpdate, model.ActionToggle, model.ActionScheduleOn, model.ActionScheduleOff, model.ActionRollback,
		}).
		Order("created_at DESC").
		First(&row).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &row, err
}

func (r *repo) CreateExposure(ctx context.Context, row *model.Exposure) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *repo) SummarizeExposures(ctx context.Context, flagKey string, since time.Time) ([]model.ExposureBucket, error) {
	var rows []model.ExposureBucket
	err := r.db.WithContext(ctx).Model(&model.Exposure{}).
		Select("variant, enabled, COUNT(DISTINCT user_id) AS count").
		Where("flag_key = ? AND created_at >= ? AND user_id IS NOT NULL", flagKey, since).
		Group("variant, enabled").
		Order("count DESC").
		Scan(&rows).Error
	return rows, err
}

func normalizePage(page, size int) (int, int) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	return page, size
}
