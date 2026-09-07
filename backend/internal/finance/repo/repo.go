package repo

import (
	"context"
	"strings"

	"github.com/Yogdunana/StarByte/backend/internal/finance/dto"
	"github.com/Yogdunana/StarByte/backend/internal/finance/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, row *model.Record) error
	Update(ctx context.Context, row *model.Record) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Record, error)
	GetNamed(ctx context.Context, id uuid.UUID) (*model.RecordNamed, error)
	List(ctx context.Context, req *dto.ListRecordRequest) ([]model.RecordNamed, int64, error)
	ListCategories(ctx context.Context) ([]model.Category, error)
	GetCategory(ctx context.Context, id uuid.UUID) (*model.Category, error)
	Summary(ctx context.Context, from, to, categoryID string) ([]model.SummaryRow, []model.CategorySumRow, error)
}

type repository struct{ db *gorm.DB }

func New(db *gorm.DB) Repository { return &repository{db: db} }

func (r *repository) Create(ctx context.Context, row *model.Record) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *repository) Update(ctx context.Context, row *model.Record) error {
	return r.db.WithContext(ctx).Save(row).Error
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Record{}, "id = ?", id).Error
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*model.Record, error) {
	var row model.Record
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&row).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &row, err
}

func (r *repository) named(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Table("finance_records AS r").
		Select(`r.*,
			COALESCE(c.name, '') AS category_name,
			COALESCE(d.name, '') AS department_name,
			COALESCE(u.real_name, u.username, '') AS creator_name`).
		Joins("LEFT JOIN finance_categories c ON c.id = r.category_id").
		Joins("LEFT JOIN departments d ON d.id = r.department_id").
		Joins("LEFT JOIN users u ON u.id = r.created_by")
}

func (r *repository) GetNamed(ctx context.Context, id uuid.UUID) (*model.RecordNamed, error) {
	var row model.RecordNamed
	err := r.named(ctx).Where("r.id = ?", id).First(&row).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &row, err
}

func (r *repository) List(ctx context.Context, req *dto.ListRecordRequest) ([]model.RecordNamed, int64, error) {
	q := applyList(r.named(ctx), req)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	page, size := normalizePage(req.Page, req.PageSize)
	var rows []model.RecordNamed
	err := q.Order("r.occurred_at DESC, r.created_at DESC").Offset((page - 1) * size).Limit(size).Find(&rows).Error
	return rows, total, err
}

func applyList(q *gorm.DB, req *dto.ListRecordRequest) *gorm.DB {
	if req.Direction != nil {
		q = q.Where("r.direction = ?", *req.Direction)
	}
	if req.CategoryID != "" {
		q = q.Where("r.category_id = ?", req.CategoryID)
	}
	if req.DepartmentID != "" {
		q = q.Where("r.department_id = ?", req.DepartmentID)
	}
	if req.From != "" {
		q = q.Where("r.occurred_at >= ?", req.From)
	}
	if req.To != "" {
		q = q.Where("r.occurred_at <= ?", req.To)
	}
	if kw := strings.TrimSpace(req.Keyword); kw != "" {
		like := "%" + kw + "%"
		q = q.Where("r.title ILIKE ? OR COALESCE(c.name, '') ILIKE ?", like, like)
	}
	return q
}

func (r *repository) ListCategories(ctx context.Context) ([]model.Category, error) {
	var rows []model.Category
	err := r.db.WithContext(ctx).Where("status = ?", model.CategoryActive).Order("code").Find(&rows).Error
	return rows, err
}

func (r *repository) GetCategory(ctx context.Context, id uuid.UUID) (*model.Category, error) {
	var row model.Category
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&row).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &row, err
}

func (r *repository) Summary(ctx context.Context, from, to, categoryID string) ([]model.SummaryRow, []model.CategorySumRow, error) {
	q := r.db.WithContext(ctx).Table("finance_records AS r")
	if from != "" {
		q = q.Where("r.occurred_at >= ?", from)
	}
	if to != "" {
		q = q.Where("r.occurred_at <= ?", to)
	}
	if categoryID != "" {
		q = q.Where("r.category_id = ?", categoryID)
	}
	var totals []model.SummaryRow
	if err := q.Select("r.direction, COALESCE(SUM(r.amount),0) AS total, COUNT(*) AS item_count").
		Group("r.direction").Find(&totals).Error; err != nil {
		return nil, nil, err
	}
	var cats []model.CategorySumRow
	cq := r.db.WithContext(ctx).Table("finance_records AS r").
		Select(`r.category_id, COALESCE(c.name,'') AS category_name, r.direction,
			COALESCE(SUM(r.amount),0) AS total, COUNT(*) AS item_count`).
		Joins("LEFT JOIN finance_categories c ON c.id = r.category_id")
	if from != "" {
		cq = cq.Where("r.occurred_at >= ?", from)
	}
	if to != "" {
		cq = cq.Where("r.occurred_at <= ?", to)
	}
	if categoryID != "" {
		cq = cq.Where("r.category_id = ?", categoryID)
	}
	err := cq.Group("r.category_id, c.name, r.direction").Order("total DESC").Find(&cats).Error
	return totals, cats, err
}

func normalizePage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}
