package repo

import (
	"context"
	"strings"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/contract/dto"
	"github.com/Yogdunana/StarByte/backend/internal/contract/model"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, row *model.Contract) error
	Update(ctx context.Context, row *model.Contract) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Contract, error)
	GetNamed(ctx context.Context, id uuid.UUID) (*model.ContractNamed, error)
	List(ctx context.Context, req *dto.ListContractRequest, scope *rbacModel.DataScopeCondition) ([]model.ContractNamed, int64, error)
	ListExpiring(ctx context.Context, from, until time.Time, scope *rbacModel.DataScopeCondition) ([]model.ContractNamed, error)
	MarkExpired(ctx context.Context, now time.Time) (int64, error)
	ListTemplates(ctx context.Context) ([]model.Template, error)
	GetTemplate(ctx context.Context, id uuid.UUID) (*model.Template, error)
	GetUser(ctx context.Context, id uuid.UUID) (*model.NamedUser, error)
}

type repository struct{ db *gorm.DB }

func New(db *gorm.DB) Repository { return &repository{db: db} }

func (r *repository) Create(ctx context.Context, row *model.Contract) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *repository) Update(ctx context.Context, row *model.Contract) error {
	return r.db.WithContext(ctx).Save(row).Error
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Contract{}, "id = ?", id).Error
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*model.Contract, error) {
	var row model.Contract
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&row).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &row, err
}

func (r *repository) named(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Table("contracts AS c").
		Select(`c.*, u.department_id,
			COALESCE(u.real_name, u.username, '') AS owner_name,
			COALESCE(t.name, '') AS template_name,
			COALESCE(f.original_name, f.name, '') AS file_name`).
		Joins("LEFT JOIN users u ON u.id = c.user_id").
		Joins("LEFT JOIN contract_templates t ON t.id = c.template_id").
		Joins("LEFT JOIN files f ON f.id = c.file_id")
}

func (r *repository) GetNamed(ctx context.Context, id uuid.UUID) (*model.ContractNamed, error) {
	var row model.ContractNamed
	err := r.named(ctx).Where("c.id = ?", id).First(&row).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &row, err
}

func (r *repository) List(ctx context.Context, req *dto.ListContractRequest, scope *rbacModel.DataScopeCondition) ([]model.ContractNamed, int64, error) {
	q := applyList(r.named(ctx), req, scope)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	page, size := normalizePage(req.Page, req.PageSize)
	var rows []model.ContractNamed
	err := q.Order("c.created_at DESC").Offset((page - 1) * size).Limit(size).Find(&rows).Error
	return rows, total, err
}

func applyList(q *gorm.DB, req *dto.ListContractRequest, scope *rbacModel.DataScopeCondition) *gorm.DB {
	if scope != nil && !scope.IsEmpty() {
		q = q.Where(scope.Query, scope.Args...)
	}
	if req.Status != nil {
		q = q.Where("c.status = ?", *req.Status)
	}
	if req.ContractType != nil {
		q = q.Where("c.contract_type = ?", *req.ContractType)
	}
	if kw := strings.TrimSpace(req.Keyword); kw != "" {
		like := "%" + kw + "%"
		q = q.Where("c.title ILIKE ? OR c.party_name ILIKE ?", like, like)
	}
	return q
}

func (r *repository) ListExpiring(ctx context.Context, from, until time.Time, scope *rbacModel.DataScopeCondition) ([]model.ContractNamed, error) {
	q := r.named(ctx)
	if scope != nil && !scope.IsEmpty() {
		q = q.Where(scope.Query, scope.Args...)
	}
	var rows []model.ContractNamed
	err := q.Where("c.status = ? AND c.expired_at IS NOT NULL AND c.expired_at <= ? AND c.expired_at >= ?",
		model.StatusActive, until, from).
		Order("c.expired_at ASC").Find(&rows).Error
	return rows, err
}

func (r *repository) MarkExpired(ctx context.Context, now time.Time) (int64, error) {
	res := r.db.WithContext(ctx).Model(&model.Contract{}).
		Where("status = ? AND expired_at IS NOT NULL AND expired_at < ?", model.StatusActive, now).
		Updates(map[string]interface{}{"status": model.StatusExpired, "updated_at": now})
	return res.RowsAffected, res.Error
}

func (r *repository) ListTemplates(ctx context.Context) ([]model.Template, error) {
	var rows []model.Template
	err := r.db.WithContext(ctx).Where("status = ?", model.TplActive).Order("code").Find(&rows).Error
	return rows, err
}

func (r *repository) GetTemplate(ctx context.Context, id uuid.UUID) (*model.Template, error) {
	var row model.Template
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&row).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &row, err
}

func (r *repository) GetUser(ctx context.Context, id uuid.UUID) (*model.NamedUser, error) {
	var row model.NamedUser
	err := r.db.WithContext(ctx).Table("users").
		Select("id, COALESCE(real_name,'') AS real_name, COALESCE(username,'') AS username").
		Where("id = ?", id).First(&row).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &row, err
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
