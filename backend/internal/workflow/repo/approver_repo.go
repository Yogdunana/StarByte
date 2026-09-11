package repo

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/model"
)

type ApproverRepo interface {
	Search(context.Context, string, uuid.UUID) ([]model.ApproverOption, error)
	ActiveUsers(context.Context, []uuid.UUID) ([]uuid.UUID, error)
	ByRole(context.Context, uuid.UUID) ([]uuid.UUID, error)
	DepartmentLeaders(context.Context, uuid.UUID) ([]uuid.UUID, error)
}
type approverRepo struct{ db *gorm.DB }

func NewApproverRepo(db *gorm.DB) ApproverRepo { return &approverRepo{db} }
func (r *approverRepo) ActiveUsers(ctx context.Context, ids []uuid.UUID) ([]uuid.UUID, error) {
	var result []uuid.UUID
	err := r.db.WithContext(ctx).Table("users").Where("id IN ? AND status=0 AND deleted_at IS NULL", ids).Order("id").Pluck("id", &result).Error
	return result, err
}
func (r *approverRepo) roleUsers(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Table("users u").Joins("JOIN user_roles ur ON ur.user_id=u.id").Joins("JOIN roles r ON r.id=ur.role_id").Where("u.status=0 AND u.deleted_at IS NULL AND r.status=0 AND (ur.expired_at IS NULL OR ur.expired_at>NOW())")
}
func (r *approverRepo) ByRole(ctx context.Context, role uuid.UUID) ([]uuid.UUID, error) {
	var result []uuid.UUID
	err := r.roleUsers(ctx).Where("r.id=?", role).Distinct().Order("u.id").Pluck("u.id", &result).Error
	return result, err
}
func (r *approverRepo) DepartmentLeaders(ctx context.Context, initiator uuid.UUID) ([]uuid.UUID, error) {
	var result []uuid.UUID
	err := r.roleUsers(ctx).Where("r.code='minister' AND u.department_id=(SELECT department_id FROM users WHERE id=? AND deleted_at IS NULL)", initiator).Distinct().Order("u.id").Pluck("u.id", &result).Error
	return result, err
}

func (r *approverRepo) Search(ctx context.Context, keyword string, exclude uuid.UUID) ([]model.ApproverOption, error) {
	result := []model.ApproverOption{}
	query := r.db.WithContext(ctx).Table("users u").Select("u.id, COALESCE(NULLIF(u.real_name,''),u.username) AS name, COALESCE(d.name,'') AS department_name").Joins("LEFT JOIN departments d ON d.id=u.department_id").Where("u.status=0 AND u.deleted_at IS NULL AND u.id<>?", exclude)
	if keyword != "" {
		query = query.Where("u.real_name ILIKE ? OR u.username ILIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	err := query.Order("u.real_name,u.id").Limit(20).Scan(&result).Error
	return result, err
}
