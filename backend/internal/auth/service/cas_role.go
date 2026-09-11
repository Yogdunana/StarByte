package service

import (
	"context"
	"strings"

	rbacmodel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	rbacrepo "github.com/Yogdunana/StarByte/backend/internal/rbac/repo"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// RoleAssigner binds a default role to an auto-provisioned user.
type RoleAssigner interface {
	AssignDefault(ctx context.Context, userID uuid.UUID) error
}

type roleAssigner struct {
	db   *gorm.DB
	repo rbacrepo.RoleRepo
	code string
}

// NewRoleAssigner assigns roleCode (e.g. member) on first CAS login. Nil-safe if db/repo missing.
func NewRoleAssigner(db *gorm.DB, roles rbacrepo.RoleRepo, roleCode string) RoleAssigner {
	code := strings.TrimSpace(roleCode)
	if code == "" {
		code = "member"
	}
	return &roleAssigner{db: db, repo: roles, code: code}
}

func (a *roleAssigner) AssignDefault(ctx context.Context, userID uuid.UUID) error {
	if a == nil || a.db == nil || a.repo == nil || userID == uuid.Nil {
		return nil
	}
	role, err := a.repo.GetByCode(ctx, a.code)
	if err != nil || role == nil {
		return err
	}
	ur := rbacmodel.UserRole{
		ID:     uuid.New(),
		UserID: userID,
		RoleID: role.ID,
	}
	return a.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&ur).Error
}
