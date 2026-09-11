package repo

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/Yogdunana/StarByte/backend/internal/task/model"
)

type AssignmentRepo interface {
	Pick(context.Context, model.AssignmentPolicy, []uuid.UUID) (*model.NamedUser, error)
	Roles(context.Context, string) ([]model.AssignmentRole, error)
}
type assignmentRepo struct{ db *gorm.DB }

func NewAssignmentRepo(db *gorm.DB) AssignmentRepo { return &assignmentRepo{db} }
func (r *assignmentRepo) Roles(ctx context.Context, keyword string) ([]model.AssignmentRole, error) {
	rows := []model.AssignmentRole{}
	q := r.db.WithContext(ctx).Table("roles").Select("id,name").Where("status=0")
	if keyword != "" {
		q = q.Where("name ILIKE ? OR code ILIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	err := q.Order("sort_order,id").Limit(20).Scan(&rows).Error
	return rows, err
}
func (r *assignmentRepo) Pick(ctx context.Context, p model.AssignmentPolicy, excluded []uuid.UUID) (*model.NamedUser, error) {
	key := p.Mode + ":" + p.DepartmentID.String()
	if p.RoleID != nil {
		key += ":" + p.RoleID.String()
	}
	now := time.Now()
	cursor := model.AssignmentCursor{ID: uuid.New(), RuleKey: key, CreatedAt: now, UpdatedAt: now}
	if err := r.db.WithContext(ctx).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "rule_key"}}, DoNothing: true}).Create(&cursor).Error; err != nil {
		return nil, err
	}
	// This lock lasts until the surrounding task-creation transaction commits.
	// Concurrent publications cannot select using the same stale queue position.
	cursor = model.AssignmentCursor{}
	if err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("rule_key=?", key).First(&cursor).Error; err != nil {
		return nil, err
	}
	q := r.db.WithContext(ctx).Table("users u").Select("u.id,u.username,u.real_name,u.department_id").Where("u.status=0 AND u.deleted_at IS NULL AND u.department_id=?", p.DepartmentID)
	if len(excluded) > 0 {
		q = q.Where("u.id NOT IN ?", excluded)
	}
	if p.RoleID != nil {
		q = q.Where("EXISTS(SELECT 1 FROM user_roles ur JOIN roles r ON r.id=ur.role_id WHERE ur.user_id=u.id AND ur.role_id=? AND r.status=0 AND (ur.expired_at IS NULL OR ur.expired_at>NOW()))", *p.RoleID)
	}
	if p.Mode == "round_robin" {
		if cursor.LastUserID != nil {
			q = q.Clauses(clause.OrderBy{Expression: clause.Expr{SQL: "CASE WHEN u.id > ? THEN 0 ELSE 1 END, u.id", Vars: []interface{}{*cursor.LastUserID}}})
		} else {
			q = q.Order("u.id")
		}
	} else {
		q = q.Order("(SELECT COUNT(*) FROM tasks t WHERE t.assignee_id=u.id AND t.deleted_at IS NULL AND t.status IN(0,1,4)) ASC, u.id ASC")
	}
	var user model.NamedUser
	if err := q.Take(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	cursor.LastUserID = &user.ID
	cursor.UpdatedAt = time.Now()
	if err := r.db.WithContext(ctx).Save(&cursor).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
