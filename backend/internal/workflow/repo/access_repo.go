package repo

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/model"
)

type AccessRepo interface {
	CanRead(context.Context, uuid.UUID, model.Viewer, bool) (bool, error)
	CanManageDefinition(context.Context, uuid.UUID, model.Viewer) (bool, error)
}
type accessRepo struct{ db *gorm.DB }

func NewAccessRepo(db *gorm.DB) AccessRepo { return &accessRepo{db: db} }

// visibleInstances limits manager access by the initiator's organization and
// always permits the initiator (and, for reads, actual assigned participants).
func visibleInstances(query *gorm.DB, viewer model.Viewer, participants bool) *gorm.DB {
	if viewer.ID == uuid.Nil {
		return query.Where("1 = 0")
	}
	scope := viewer.Scope
	if scope != nil && !scope.IsSelf && strings.TrimSpace(scope.Query) == "" {
		return query
	}
	clause := "flow_instances.initiator_id = ?"
	args := []interface{}{viewer.ID}
	if participants {
		clause += " OR EXISTS (SELECT 1 FROM flow_tasks participant WHERE participant.instance_id = flow_instances.id AND participant.assignee_id = ?)"
		args = append(args, viewer.ID)
	}
	if scope != nil && !scope.IsSelf {
		switch strings.TrimSpace(scope.Query) {
		case "department_id = ?", "department_id IN ?":
			clause += " OR EXISTS (SELECT 1 FROM users owner WHERE owner.id = flow_instances.initiator_id AND owner.deleted_at IS NULL AND owner." + scope.Query + ")"
			args = append(args, scope.Args...)
		}
	}
	return query.Where("("+clause+")", args...)
}
func (r *accessRepo) CanRead(ctx context.Context, id uuid.UUID, viewer model.Viewer, participants bool) (bool, error) {
	var count int64
	err := visibleInstances(r.db.WithContext(ctx).Model(&model.FlowInstance{}), viewer, participants).Where("flow_instances.id = ?", id).Count(&count).Error
	return count > 0, err
}

func (r *accessRepo) CanManageDefinition(ctx context.Context, id uuid.UUID, viewer model.Viewer) (bool, error) {
	if viewer.ID == uuid.Nil {
		return false, nil
	}
	scope := viewer.Scope
	query := r.db.WithContext(ctx).Model(&model.FlowDefinition{}).Where("id = ?", id)
	if scope == nil || scope.IsSelf || strings.TrimSpace(scope.Query) != "" {
		clause, args := "created_by = ?", []interface{}{viewer.ID}
		if scope != nil && !scope.IsSelf {
			switch strings.TrimSpace(scope.Query) {
			case "department_id = ?", "department_id IN ?":
				clause += " OR EXISTS (SELECT 1 FROM users owner WHERE owner.id = flow_definitions.created_by AND owner.deleted_at IS NULL AND owner." + scope.Query + ")"
				args = append(args, scope.Args...)
			}
		}
		query = query.Where("("+clause+")", args...)
	}
	var count int64
	err := query.Count(&count).Error
	return count > 0, err
}
