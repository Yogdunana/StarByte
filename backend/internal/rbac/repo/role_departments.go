package repo

import (
	"context"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func LoadRoleDepartments(ctx context.Context, db *gorm.DB, user uuid.UUID) (map[string][]uuid.UUID, error) {
	var rows []struct {
		Code         string
		DepartmentID *uuid.UUID
	}
	err := db.WithContext(ctx).Raw(`SELECT r.code,d.id AS department_id FROM user_role_departments s JOIN user_roles ur ON ur.id=s.user_role_id JOIN roles r ON r.id=ur.role_id LEFT JOIN departments d ON d.id=s.department_id AND d.status=0
 WHERE ur.user_id=? AND r.status=0 AND (ur.expired_at IS NULL OR ur.expired_at>NOW())`, user).Scan(&rows).Error
	scopes := map[string][]uuid.UUID{}
	for _, row := range rows {
		if _, ok := scopes[row.Code]; !ok {
			scopes[row.Code] = []uuid.UUID{}
		}
		if row.DepartmentID != nil {
			scopes[row.Code] = append(scopes[row.Code], *row.DepartmentID)
		}
	}
	return scopes, err
}
