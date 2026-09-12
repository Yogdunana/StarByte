package repo

import (
	"github.com/Yogdunana/StarByte/backend/internal/knowledge/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ReadableScope is the list/search/tree visibility window for one viewer.
// Staff sees everything; others only see rows canRead would allow.
type ReadableScope struct {
	Staff  bool
	UserID uuid.UUID
	Perms  []string
}

func (s ReadableScope) hasStar() bool {
	for _, p := range s.Perms {
		if p == "*" {
			return true
		}
	}
	return false
}

func (s ReadableScope) permCodes() []string {
	out := make([]string, 0, len(s.Perms))
	for _, p := range s.Perms {
		if p != "" && p != "*" {
			out = append(out, p)
		}
	}
	return out
}

// Visible matches applyReadable / canRead for in-memory repos and tests.
func Visible(d *model.Doc, scope ReadableScope) bool {
	if d == nil {
		return false
	}
	if scope.Staff || (scope.UserID != uuid.Nil && d.AuthorID == scope.UserID) {
		return true
	}
	if d.Status != model.StatusPublished {
		return false
	}
	switch d.Visibility {
	case model.VisibilityPublic:
		return true
	case model.VisibilityAuthenticated:
		return scope.UserID != uuid.Nil
	case model.VisibilityPermission:
		if scope.UserID == uuid.Nil {
			return false
		}
		if d.PermissionCode == "" || scope.hasStar() {
			return true
		}
		for _, p := range scope.Perms {
			if p == d.PermissionCode {
				return true
			}
		}
		return false
	default:
		return false
	}
}

func applyReadable(q *gorm.DB, scope ReadableScope) *gorm.DB {
	if scope.Staff {
		return q
	}
	if scope.UserID == uuid.Nil {
		return q.Where("d.status = ? AND d.visibility = ?", model.StatusPublished, model.VisibilityPublic)
	}
	publishedOpen := "(d.status = ? AND d.visibility IN (?, ?))"
	args := []any{scope.UserID, model.StatusPublished, model.VisibilityPublic, model.VisibilityAuthenticated}
	if scope.hasStar() {
		return q.Where(
			"d.author_id = ? OR "+publishedOpen+" OR (d.status = ? AND d.visibility = ?)",
			append(args, model.StatusPublished, model.VisibilityPermission)...,
		)
	}
	if codes := scope.permCodes(); len(codes) > 0 {
		return q.Where(
			"d.author_id = ? OR "+publishedOpen+" OR (d.status = ? AND d.visibility = ? AND (d.permission_code = '' OR d.permission_code IN ?))",
			append(args, model.StatusPublished, model.VisibilityPermission, codes)...,
		)
	}
	return q.Where(
		"d.author_id = ? OR "+publishedOpen+" OR (d.status = ? AND d.visibility = ? AND d.permission_code = '')",
		append(args, model.StatusPublished, model.VisibilityPermission)...,
	)
}
