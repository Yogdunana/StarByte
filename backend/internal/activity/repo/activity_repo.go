package repo

import (
	"context"
	"strings"

	"github.com/Yogdunana/StarByte/backend/internal/activity/dto"
	"github.com/Yogdunana/StarByte/backend/internal/activity/model"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ActivityRepo interface {
	Create(ctx context.Context, a *model.Activity) error
	Update(ctx context.Context, a *model.Activity) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Activity, error)
	GetByIDForUpdate(ctx context.Context, id uuid.UUID) (*model.Activity, error)
	GetByIDWithNames(ctx context.Context, id uuid.UUID) (*model.ActivityWithNames, error)
	List(ctx context.Context, req *dto.ListActivityRequest, scope *rbacModel.DataScopeCondition) ([]model.ActivityWithNames, int64, error)
	UpdateCheckinToken(ctx context.Context, id uuid.UUID, secret string, nonce int64) error
	GetUser(ctx context.Context, id uuid.UUID) (*model.NamedUser, error)
	GetUserDepartment(ctx context.Context, userID uuid.UUID) (*uuid.UUID, error)
}

type activityRepo struct{ db *gorm.DB }

func NewActivityRepo(db *gorm.DB) ActivityRepo {
	return &activityRepo{db: db}
}

func (r *activityRepo) Create(ctx context.Context, a *model.Activity) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *activityRepo) Update(ctx context.Context, a *model.Activity) error {
	return r.db.WithContext(ctx).Save(a).Error
}

func (r *activityRepo) UpdateCheckinToken(ctx context.Context, id uuid.UUID, secret string, nonce int64) error {
	return r.db.WithContext(ctx).Model(&model.Activity{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"checkin_secret": secret,
			"checkin_nonce":  nonce,
		}).Error
}

func (r *activityRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Activity{}, "id = ?", id).Error
}

func (r *activityRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Activity, error) {
	return r.getByID(ctx, id, false)
}

func (r *activityRepo) GetByIDForUpdate(ctx context.Context, id uuid.UUID) (*model.Activity, error) {
	return r.getByID(ctx, id, true)
}

func (r *activityRepo) getByID(ctx context.Context, id uuid.UUID, forUpdate bool) (*model.Activity, error) {
	var a model.Activity
	q := r.db.WithContext(ctx).Where("id = ?", id)
	if forUpdate {
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	err := q.First(&a).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *activityRepo) namedQuery(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Table("activities AS a").
		Select(`a.*, COALESCE(u.real_name, u.username, '') AS organizer_name,
			(SELECT COUNT(*) FROM activity_registrations r WHERE r.activity_id = a.id AND r.status = 1) AS registered_count,
			(SELECT COUNT(*) FROM activity_registrations r WHERE r.activity_id = a.id AND r.checkin_status = 1) AS checked_in_count`).
		Joins("LEFT JOIN users u ON u.id = a.organizer_id").
		Where("a.deleted_at IS NULL")
}

func (r *activityRepo) GetByIDWithNames(ctx context.Context, id uuid.UUID) (*model.ActivityWithNames, error) {
	var row model.ActivityWithNames
	err := r.namedQuery(ctx).Where("a.id = ?", id).First(&row).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *activityRepo) applyListFilters(q *gorm.DB, req *dto.ListActivityRequest) *gorm.DB {
	if req == nil {
		return q
	}
	if req.Status != nil {
		q = q.Where("a.status = ?", *req.Status)
	}
	if req.Category != "" {
		q = q.Where("a.category = ?", req.Category)
	}
	if kw := strings.TrimSpace(req.Keyword); kw != "" {
		like := "%" + kw + "%"
		q = q.Where("a.title ILIKE ? OR a.location ILIKE ? OR a.description ILIKE ?", like, like, like)
	}
	if req.StartDate != "" {
		q = q.Where("a.start_time >= ?", req.StartDate)
	}
	if req.EndDate != "" {
		q = q.Where("a.start_time <= ?", req.EndDate+" 23:59:59")
	}
	return q
}

func (r *activityRepo) List(ctx context.Context, req *dto.ListActivityRequest, scope *rbacModel.DataScopeCondition) ([]model.ActivityWithNames, int64, error) {
	// Count 必须避开 namedQuery 里的 COUNT(*) 子查询，否则 GORM 会把整行扫进 int64。
	countQ := r.applyListFilters(r.db.WithContext(ctx).Table("activities AS a").Where("a.deleted_at IS NULL"), req)
	if scope != nil && !scope.IsEmpty() {
		countQ = countQ.Where(scope.Query, scope.Args...)
	}
	var total int64
	if err := countQ.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	page, size := normalizePage(req.Page, req.PageSize)
	q := r.applyListFilters(r.namedQuery(ctx), req)
	if scope != nil && !scope.IsEmpty() {
		q = q.Where(scope.Query, scope.Args...)
	}
	var rows []model.ActivityWithNames
	err := q.Order("a.start_time DESC").Offset((page - 1) * size).Limit(size).Find(&rows).Error
	return rows, total, err
}

// GetUserDepartment 返回用户所属部门 ID。用户不存在或 department_id 为 NULL 时返回 (nil, nil)。
func (r *activityRepo) GetUserDepartment(ctx context.Context, userID uuid.UUID) (*uuid.UUID, error) {
	var row struct {
		DepartmentID *uuid.UUID `gorm:"column:department_id"`
	}
	err := r.db.WithContext(ctx).Table("users").Select("department_id").Where("id = ?", userID).First(&row).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return row.DepartmentID, nil
}

func (r *activityRepo) GetUser(ctx context.Context, id uuid.UUID) (*model.NamedUser, error) {
	var u model.NamedUser
	err := r.db.WithContext(ctx).Table("users AS u").
		Select("u.id, u.real_name, u.username").
		Where("u.id = ?", id).First(&u).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
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
