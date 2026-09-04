package repo

import (
	"context"

	"github.com/Yogdunana/StarByte/backend/internal/member/dto"
	"github.com/Yogdunana/StarByte/backend/internal/member/model"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ApplicationRepo 入会申请数据访问。
type ApplicationRepo interface {
	Create(ctx context.Context, app *model.MemberApplication) error
	Update(ctx context.Context, app *model.MemberApplication) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.MemberApplication, error)
	GetByIDWithNames(ctx context.Context, id uuid.UUID) (*model.ApplicationWithNames, error)
	List(ctx context.Context, req *dto.ListApplicationRequest, scope *rbacModel.DataScopeCondition) ([]model.ApplicationWithNames, int64, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]model.ApplicationWithNames, error)
	HasOpenApplication(ctx context.Context, userID uuid.UUID) (bool, error)
	CreateHistory(ctx context.Context, h *model.ApplicationHistory) error
	ListHistory(ctx context.Context, applicationID uuid.UUID) ([]model.ApplicationHistory, error)
	ListDepartments(ctx context.Context) ([]model.NamedItem, error)
	Stats(ctx context.Context, start, end, groupBy string) ([]model.StatBucket, error)
}

type applicationRepo struct {
	db *gorm.DB
}

// NewApplicationRepo 创建申请仓储。
func NewApplicationRepo(db *gorm.DB) ApplicationRepo {
	return &applicationRepo{db: db}
}

func (r *applicationRepo) Create(ctx context.Context, app *model.MemberApplication) error {
	return r.db.WithContext(ctx).Create(app).Error
}

func (r *applicationRepo) Update(ctx context.Context, app *model.MemberApplication) error {
	return r.db.WithContext(ctx).Save(app).Error
}

func (r *applicationRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.MemberApplication, error) {
	var app model.MemberApplication
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&app).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &app, nil
}

func (r *applicationRepo) namedQuery(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Table("member_applications AS a").
		Select("a.*, u.username, COALESCE(d.name, '') AS department_name, COALESCE(r.real_name, '') AS reviewer_name").
		Joins("LEFT JOIN users u ON u.id = a.user_id").
		Joins("LEFT JOIN departments d ON d.id = a.department_id").
		Joins("LEFT JOIN users r ON r.id = a.reviewer_id")
}

func applyAppScope(q *gorm.DB, scope *rbacModel.DataScopeCondition) *gorm.DB {
	if scope == nil || scope.IsEmpty() {
		return q
	}
	return q.Where(scope.Query, scope.Args...)
}

func (r *applicationRepo) GetByIDWithNames(ctx context.Context, id uuid.UUID) (*model.ApplicationWithNames, error) {
	var row model.ApplicationWithNames
	err := r.namedQuery(ctx).Where("a.id = ?", id).First(&row).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *applicationRepo) List(ctx context.Context, req *dto.ListApplicationRequest, scope *rbacModel.DataScopeCondition) ([]model.ApplicationWithNames, int64, error) {
	q := r.namedQuery(ctx)
	q = applyAppScope(q, scope)
	if req.Status != nil {
		q = q.Where("a.status = ?", *req.Status)
	}
	if req.ApplicantType != nil {
		q = q.Where("a.type = ?", *req.ApplicantType)
	}
	if req.DepartmentID != "" {
		q = q.Where("a.department_id = ?", req.DepartmentID)
	}
	if kw := req.Keyword; kw != "" {
		like := "%" + kw + "%"
		q = q.Where("a.real_name ILIKE ? OR a.student_no ILIKE ? OR u.username ILIKE ?", like, like, like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	page, pageSize := normalizePage(req.Page, req.PageSize)
	var rows []model.ApplicationWithNames
	err := q.Order("a.created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error
	return rows, total, err
}

func (r *applicationRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]model.ApplicationWithNames, error) {
	var rows []model.ApplicationWithNames
	err := r.namedQuery(ctx).Where("a.user_id = ?", userID).Order("a.created_at DESC").Find(&rows).Error
	return rows, err
}

func (r *applicationRepo) HasOpenApplication(ctx context.Context, userID uuid.UUID) (bool, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.MemberApplication{}).
		Where("user_id = ? AND status IN ?", userID, []int16{
			model.AppPending, model.AppReviewing, model.AppInterviewing, model.AppSupplement,
		}).Count(&n).Error
	return n > 0, err
}

func (r *applicationRepo) CreateHistory(ctx context.Context, h *model.ApplicationHistory) error {
	return r.db.WithContext(ctx).Create(h).Error
}

func (r *applicationRepo) ListHistory(ctx context.Context, applicationID uuid.UUID) ([]model.ApplicationHistory, error) {
	var rows []model.ApplicationHistory
	err := r.db.WithContext(ctx).Where("application_id = ?", applicationID).
		Order("created_at ASC").Find(&rows).Error
	return rows, err
}

func (r *applicationRepo) ListDepartments(ctx context.Context) ([]model.NamedItem, error) {
	var rows []model.NamedItem
	err := r.db.WithContext(ctx).Table("departments").
		Select("id, name").Where("status = ?", 0).Order("sort_order ASC, name ASC").
		Scan(&rows).Error
	return rows, err
}

func (r *applicationRepo) Stats(ctx context.Context, start, end, groupBy string) ([]model.StatBucket, error) {
	q := r.db.WithContext(ctx).Table("member_applications AS a")
	if start != "" {
		q = q.Where("a.created_at >= ?", start)
	}
	if end != "" {
		q = q.Where("a.created_at < ?", end+" 23:59:59")
	}
	var (
		selectSQL string
		groupSQL  string
	)
	switch groupBy {
	case "department":
		selectSQL = "COALESCE(a.department_id::text, 'none') AS key, COALESCE(d.name, '未分配') AS label, COUNT(*) AS count"
		groupSQL = "a.department_id, d.name"
		q = q.Joins("LEFT JOIN departments d ON d.id = a.department_id")
	case "type":
		selectSQL = "a.type::text AS key, CASE a.type WHEN 1 THEN '会员' WHEN 2 THEN '干事' ELSE '其他' END AS label, COUNT(*) AS count"
		groupSQL = "a.type"
	default:
		selectSQL = "TO_CHAR(a.created_at, 'YYYY-MM-DD') AS key, TO_CHAR(a.created_at, 'YYYY-MM-DD') AS label, COUNT(*) AS count"
		groupSQL = "TO_CHAR(a.created_at, 'YYYY-MM-DD')"
	}
	var rows []model.StatBucket
	err := q.Select(selectSQL).Group(groupSQL).Order("count DESC").Scan(&rows).Error
	return rows, err
}

func normalizePage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return page, pageSize
}
