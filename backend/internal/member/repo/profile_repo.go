package repo

import (
	"context"
	"strings"

	"github.com/Yogdunana/StarByte/backend/internal/member/dto"
	"github.com/Yogdunana/StarByte/backend/internal/member/model"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProfileRepo 人员档案数据访问。
type ProfileRepo interface {
	Create(ctx context.Context, p *model.MemberProfile) error
	Update(ctx context.Context, p *model.MemberProfile) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.MemberProfile, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) (*model.MemberProfile, error)
	GetByIDWithNames(ctx context.Context, id uuid.UUID) (*model.ProfileWithNames, error)
	GetByStudentNo(ctx context.Context, studentNo string, excludeID *uuid.UUID) (*model.MemberProfile, error)
	List(ctx context.Context, req *dto.ListProfileRequest, scope *rbacModel.DataScopeCondition) ([]model.ProfileWithNames, int64, error)
	CreateHistories(ctx context.Context, rows []model.ProfileHistory) error
	ListHistory(ctx context.Context, profileID uuid.UUID) ([]model.ProfileHistory, error)
	Stats(ctx context.Context, groupBy string) ([]model.StatBucket, error)
}

type profileRepo struct {
	db *gorm.DB
}

// NewProfileRepo 创建档案仓储。
func NewProfileRepo(db *gorm.DB) ProfileRepo {
	return &profileRepo{db: db}
}

func (r *profileRepo) Create(ctx context.Context, p *model.MemberProfile) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *profileRepo) Update(ctx context.Context, p *model.MemberProfile) error {
	return r.db.WithContext(ctx).Save(p).Error
}

func (r *profileRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.MemberProfile, error) {
	var p model.MemberProfile
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&p).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *profileRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*model.MemberProfile, error) {
	var p model.MemberProfile
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&p).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *profileRepo) namedQuery(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Table("member_profiles AS p").
		Select("p.*, u.username, COALESCE(d.name, '') AS department_name, COALESCE(pos.name, '') AS position_name").
		Joins("LEFT JOIN users u ON u.id = p.user_id").
		Joins("LEFT JOIN departments d ON d.id = p.department_id").
		Joins("LEFT JOIN positions pos ON pos.id = p.position_id")
}

func (r *profileRepo) GetByIDWithNames(ctx context.Context, id uuid.UUID) (*model.ProfileWithNames, error) {
	var row model.ProfileWithNames
	err := r.namedQuery(ctx).Where("p.id = ?", id).First(&row).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *profileRepo) GetByStudentNo(ctx context.Context, studentNo string, excludeID *uuid.UUID) (*model.MemberProfile, error) {
	if studentNo == "" {
		return nil, nil
	}
	q := r.db.WithContext(ctx).Where("student_no = ?", studentNo)
	if excludeID != nil {
		q = q.Where("id <> ?", *excludeID)
	}
	var p model.MemberProfile
	err := q.First(&p).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *profileRepo) List(ctx context.Context, req *dto.ListProfileRequest, scope *rbacModel.DataScopeCondition) ([]model.ProfileWithNames, int64, error) {
	q := r.namedQuery(ctx)
	q = applyAppScope(q, scope)
	if req.DepartmentID != "" {
		q = q.Where("p.department_id = ?", req.DepartmentID)
	}
	if req.MemberType != nil {
		q = q.Where("p.member_type = ?", *req.MemberType)
	}
	if req.Status != nil {
		q = q.Where("p.status = ?", *req.Status)
	}
	if ids := parseIDs(req.IDs); len(ids) > 0 {
		q = q.Where("p.id IN ?", ids)
	}
	if kw := req.Keyword; kw != "" {
		like := "%" + kw + "%"
		q = q.Where(
			"p.real_name ILIKE ? OR p.student_no ILIKE ? OR p.skills::text ILIKE ? OR d.name ILIKE ?",
			like, like, like, like,
		)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	page, pageSize := normalizePage(req.Page, req.PageSize)
	var rows []model.ProfileWithNames
	err := q.Order("p.created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error
	return rows, total, err
}

func (r *profileRepo) CreateHistories(ctx context.Context, rows []model.ProfileHistory) error {
	if len(rows) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&rows).Error
}

func (r *profileRepo) ListHistory(ctx context.Context, profileID uuid.UUID) ([]model.ProfileHistory, error) {
	var rows []model.ProfileHistory
	err := r.db.WithContext(ctx).Where("profile_id = ?", profileID).
		Order("created_at DESC").Find(&rows).Error
	return rows, err
}

func (r *profileRepo) Stats(ctx context.Context, groupBy string) ([]model.StatBucket, error) {
	q := r.db.WithContext(ctx).Table("member_profiles AS p")
	var selectSQL, groupSQL string
	switch groupBy {
	case "grade":
		selectSQL = "COALESCE(NULLIF(p.grade, ''), '未填') AS key, COALESCE(NULLIF(p.grade, ''), '未填') AS label, COUNT(*) AS count"
		groupSQL = "p.grade"
	case "type":
		selectSQL = "p.member_type::text AS key, CASE p.member_type WHEN 1 THEN '会员' WHEN 2 THEN '干事' WHEN 3 THEN '部长' WHEN 4 THEN '社长' ELSE '其他' END AS label, COUNT(*) AS count"
		groupSQL = "p.member_type"
	case "status":
		selectSQL = "p.status::text AS key, CASE p.status WHEN 0 THEN '正常' WHEN 1 THEN '禁用' WHEN 2 THEN '已退出' ELSE '其他' END AS label, COUNT(*) AS count"
		groupSQL = "p.status"
	default:
		selectSQL = "COALESCE(p.department_id::text, 'none') AS key, COALESCE(d.name, '未分配') AS label, COUNT(*) AS count"
		groupSQL = "p.department_id, d.name"
		q = q.Joins("LEFT JOIN departments d ON d.id = p.department_id")
	}
	var rows []model.StatBucket
	err := q.Select(selectSQL).Group(groupSQL).Order("count DESC").Scan(&rows).Error
	return rows, err
}

func parseIDs(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
