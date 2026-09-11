package repo

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Yogdunana/StarByte/backend/internal/meeting/model"
	rbac "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
)

type AccessRepo interface {
	CanAccess(context.Context, uuid.UUID, model.Viewer) (bool, error)
	AllowedIDs(context.Context, []uuid.UUID, model.Viewer) (map[uuid.UUID]bool, error)
}
type accessRepo struct{ db *gorm.DB }

func NewAccessRepo(db *gorm.DB) AccessRepo { return &accessRepo{db} }

// MeetingScope recognizes only server-produced clauses. Attendance grants read
// access to a cross-department meeting, never the right to manage its contents.
func MeetingScope(viewer model.Viewer) *rbac.DataScopeCondition {
	if viewer.ID == uuid.Nil {
		return &rbac.DataScopeCondition{Query: "1 = 0"}
	}
	scope := viewer.Scope
	if scope != nil && !scope.IsSelf && strings.TrimSpace(scope.Query) == "" {
		return &rbac.DataScopeCondition{}
	}
	query := "m.organizer_id = ?"
	args := []interface{}{viewer.ID}
	if !viewer.Manage {
		query += " OR EXISTS (SELECT 1 FROM meeting_attendees participant WHERE participant.meeting_id=m.id AND participant.user_id=?)"
		args = append(args, viewer.ID)
	}
	if scope != nil && !scope.IsSelf {
		switch strings.TrimSpace(scope.Query) {
		case "department_id = ?", "department_id IN ?":
			query += " OR u." + scope.Query
			args = append(args, scope.Args...)
		}
	}
	return &rbac.DataScopeCondition{Query: "(" + query + ")", Args: args}
}
func (r *accessRepo) CanAccess(ctx context.Context, id uuid.UUID, viewer model.Viewer) (bool, error) {
	scope := MeetingScope(viewer)
	query := r.db.WithContext(ctx).Table("meetings m").Joins("LEFT JOIN users u ON u.id=m.organizer_id").Where("m.id=?", id)
	if scope.Query != "" {
		query = query.Where(scope.Query, scope.Args...)
	}
	var count int64
	err := query.Count(&count).Error
	return count > 0, err
}

func (r *accessRepo) AllowedIDs(ctx context.Context, ids []uuid.UUID, viewer model.Viewer) (map[uuid.UUID]bool, error) {
	allowed := map[uuid.UUID]bool{}
	if len(ids) == 0 {
		return allowed, nil
	}
	scope := MeetingScope(viewer)
	query := r.db.WithContext(ctx).Table("meetings m").Joins("LEFT JOIN users u ON u.id=m.organizer_id").Where("m.id IN ?", ids)
	if scope.Query != "" {
		query = query.Where(scope.Query, scope.Args...)
	}
	var rows []uuid.UUID
	if err := query.Pluck("m.id", &rows).Error; err != nil {
		return nil, err
	}
	for _, id := range rows {
		allowed[id] = true
	}
	return allowed, nil
}
