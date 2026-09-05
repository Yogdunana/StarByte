package repo

import (
	"context"
	"strings"

	"github.com/Yogdunana/StarByte/backend/internal/internship/dto"
	"github.com/Yogdunana/StarByte/backend/internal/internship/model"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type InternshipRepo interface {
	Create(ctx context.Context, row *model.Internship) error
	Update(ctx context.Context, row *model.Internship) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Internship, error)
	GetByIDWithNames(ctx context.Context, id uuid.UUID) (*model.InternshipWithNames, error)
	List(ctx context.Context, req *dto.ListInternshipRequest, scope *rbacModel.DataScopeCondition) ([]model.InternshipWithNames, int64, error)
	ListByUser(ctx context.Context, userID uuid.UUID, status *int16) ([]model.InternshipWithNames, error)
	GetUser(ctx context.Context, id uuid.UUID) (*model.NamedUser, error)
	GetConfig(ctx context.Context, key string) (*model.SystemConfig, error)
	UpsertConfig(ctx context.Context, cfg *model.SystemConfig) error
	ListForStats(ctx context.Context, startDate, endDate, departmentID string, scope *rbacModel.DataScopeCondition) ([]model.InternshipWithNames, error)
}

type internshipRepo struct{ db *gorm.DB }

func NewInternshipRepo(db *gorm.DB) InternshipRepo {
	return &internshipRepo{db: db}
}

func (r *internshipRepo) Create(ctx context.Context, row *model.Internship) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *internshipRepo) Update(ctx context.Context, row *model.Internship) error {
	return r.db.WithContext(ctx).Save(row).Error
}

func (r *internshipRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Internship{}, "id = ?", id).Error
}

func (r *internshipRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Internship, error) {
	var row model.Internship
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&row).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *internshipRepo) namedQuery(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Table("internships AS i").
		Select(`i.*,
			COALESCE(u.real_name, u.username, '') AS user_name,
			COALESCE(u.avatar, '') AS user_avatar,
			COALESCE(d.name, '') AS department_name,
			COALESCE(m.real_name, m.username, '') AS mentor_name,
			COALESCE(s.real_name, s.username, '') AS supervisor_name`).
		Joins("LEFT JOIN users u ON u.id = i.user_id").
		Joins("LEFT JOIN departments d ON d.id = i.department_id").
		Joins("LEFT JOIN users m ON m.id = i.mentor_id").
		Joins("LEFT JOIN users s ON s.id = i.supervisor_id")
}

func (r *internshipRepo) GetByIDWithNames(ctx context.Context, id uuid.UUID) (*model.InternshipWithNames, error) {
	var row model.InternshipWithNames
	err := r.namedQuery(ctx).Where("i.id = ?", id).First(&row).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *internshipRepo) List(ctx context.Context, req *dto.ListInternshipRequest, scope *rbacModel.DataScopeCondition) ([]model.InternshipWithNames, int64, error) {
	q := applyListFilters(r.namedQuery(ctx), req, scope)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	page, size := normalizePage(req.Page, req.PageSize)
	var rows []model.InternshipWithNames
	err := q.Order("i.start_date DESC, i.created_at DESC").Offset((page - 1) * size).Limit(size).Find(&rows).Error
	return rows, total, err
}

func applyListFilters(q *gorm.DB, req *dto.ListInternshipRequest, scope *rbacModel.DataScopeCondition) *gorm.DB {
	q = applyScope(q, scope)
	if req.Status != nil {
		q = q.Where("i.status = ?", *req.Status)
	}
	if req.Type != nil {
		q = q.Where("i.type = ?", *req.Type)
	}
	if req.DepartmentID != "" {
		q = q.Where("i.department_id = ?", req.DepartmentID)
	}
	if req.UserID != "" {
		q = q.Where("i.user_id = ?", req.UserID)
	}
	if kw := strings.TrimSpace(req.Keyword); kw != "" {
		like := "%" + kw + "%"
		q = q.Where("i.title ILIKE ? OR i.organization ILIKE ? OR COALESCE(u.real_name, u.username, '') ILIKE ?", like, like, like)
	}
	return q
}

func (r *internshipRepo) ListByUser(ctx context.Context, userID uuid.UUID, status *int16) ([]model.InternshipWithNames, error) {
	q := r.namedQuery(ctx).Where("i.user_id = ?", userID)
	if status != nil {
		q = q.Where("i.status = ?", *status)
	}
	var rows []model.InternshipWithNames
	err := q.Order("i.start_date DESC, i.created_at DESC").Find(&rows).Error
	return rows, err
}

func (r *internshipRepo) GetUser(ctx context.Context, id uuid.UUID) (*model.NamedUser, error) {
	var u model.NamedUser
	err := r.db.WithContext(ctx).Table("users").
		Select("id, real_name, username, avatar, department_id").
		Where("id = ?", id).First(&u).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *internshipRepo) GetConfig(ctx context.Context, key string) (*model.SystemConfig, error) {
	var c model.SystemConfig
	err := r.db.WithContext(ctx).Where("config_key = ?", key).First(&c).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *internshipRepo) UpsertConfig(ctx context.Context, cfg *model.SystemConfig) error {
	if cfg.ID == uuid.Nil {
		cfg.ID = uuid.New()
		return r.db.WithContext(ctx).Create(cfg).Error
	}
	return r.db.WithContext(ctx).Save(cfg).Error
}

func (r *internshipRepo) ListForStats(ctx context.Context, startDate, endDate, departmentID string, scope *rbacModel.DataScopeCondition) ([]model.InternshipWithNames, error) {
	q := applyScope(r.namedQuery(ctx), scope)
	if startDate != "" {
		q = q.Where("i.start_date >= ? OR (i.end_date IS NULL OR i.end_date >= ?)", startDate, startDate)
	}
	if endDate != "" {
		q = q.Where("i.start_date <= ?", endDate)
	}
	if departmentID != "" {
		q = q.Where("i.department_id = ?", departmentID)
	}
	var rows []model.InternshipWithNames
	err := q.Order("i.start_date DESC").Find(&rows).Error
	return rows, err
}
