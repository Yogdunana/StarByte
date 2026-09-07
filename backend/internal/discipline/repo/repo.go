package repo

import (
	"context"
	"strings"

	"github.com/Yogdunana/StarByte/backend/internal/discipline/dto"
	"github.com/Yogdunana/StarByte/backend/internal/discipline/model"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, row *model.Record) error
	Update(ctx context.Context, row *model.Record) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Record, error)
	GetNamed(ctx context.Context, id uuid.UUID) (*model.RecordNamed, error)
	List(ctx context.Context, req *dto.ListRecordRequest, scope *rbacModel.DataScopeCondition) ([]model.RecordNamed, int64, error)
	GetUser(ctx context.Context, id uuid.UUID) (*model.NamedUser, error)
	CreateAppeal(ctx context.Context, row *model.Appeal) error
	ListAppeals(ctx context.Context, recordID uuid.UUID) ([]model.Appeal, error)
	HasOpenAppeal(ctx context.Context, recordID uuid.UUID) (bool, error)
}

type repository struct{ db *gorm.DB }

func New(db *gorm.DB) Repository { return &repository{db: db} }

func (r *repository) Create(ctx context.Context, row *model.Record) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *repository) Update(ctx context.Context, row *model.Record) error {
	return r.db.WithContext(ctx).Save(row).Error
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
	return r.db.WithContext(ctx).Table("discipline_records AS r").
		Select(`r.*, u.department_id,
			COALESCE(u.real_name, u.username, '') AS user_name,
			COALESCE(i.real_name, i.username, '') AS issuer_name,
			COALESCE(d.name, '') AS department_name`).
		Joins("LEFT JOIN users u ON u.id = r.user_id").
		Joins("LEFT JOIN users i ON i.id = r.issued_by").
		Joins("LEFT JOIN departments d ON d.id = u.department_id")
}

func (r *repository) GetNamed(ctx context.Context, id uuid.UUID) (*model.RecordNamed, error) {
	var row model.RecordNamed
	err := r.named(ctx).Where("r.id = ?", id).First(&row).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &row, err
}

func (r *repository) List(ctx context.Context, req *dto.ListRecordRequest, scope *rbacModel.DataScopeCondition) ([]model.RecordNamed, int64, error) {
	q := applyList(r.named(ctx), req, scope)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	page, size := normalizePage(req.Page, req.PageSize)
	var rows []model.RecordNamed
	err := q.Order("r.created_at DESC").Offset((page - 1) * size).Limit(size).Find(&rows).Error
	return rows, total, err
}

func applyList(q *gorm.DB, req *dto.ListRecordRequest, scope *rbacModel.DataScopeCondition) *gorm.DB {
	if scope != nil && !scope.IsEmpty() {
		q = q.Where(scope.Query, scope.Args...)
	}
	if req.Status != nil {
		q = q.Where("r.status = ?", *req.Status)
	}
	if req.Level != nil {
		q = q.Where("r.level = ?", *req.Level)
	}
	if req.UserID != "" {
		q = q.Where("r.user_id = ?", req.UserID)
	}
	if kw := strings.TrimSpace(req.Keyword); kw != "" {
		like := "%" + kw + "%"
		q = q.Where("r.title ILIKE ? OR COALESCE(u.real_name, u.username, '') ILIKE ?", like, like)
	}
	return q
}

func (r *repository) GetUser(ctx context.Context, id uuid.UUID) (*model.NamedUser, error) {
	var row model.NamedUser
	err := r.db.WithContext(ctx).Table("users").
		Select("id, COALESCE(real_name,'') AS real_name, COALESCE(username,'') AS username, department_id").
		Where("id = ?", id).First(&row).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &row, err
}

func (r *repository) CreateAppeal(ctx context.Context, row *model.Appeal) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *repository) ListAppeals(ctx context.Context, recordID uuid.UUID) ([]model.Appeal, error) {
	var rows []model.Appeal
	err := r.db.WithContext(ctx).Where("record_id = ?", recordID).Order("created_at DESC").Find(&rows).Error
	return rows, err
}

func (r *repository) HasOpenAppeal(ctx context.Context, recordID uuid.UUID) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.Appeal{}).
		Where("record_id = ? AND status = ?", recordID, model.AppealPending).Count(&n).Error
	return n > 0, err
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
