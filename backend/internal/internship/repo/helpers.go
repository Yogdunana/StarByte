package repo

import (
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"gorm.io/gorm"
)

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

func applyScope(q *gorm.DB, scope *rbacModel.DataScopeCondition) *gorm.DB {
	if scope == nil || scope.IsEmpty() {
		return q
	}
	return q.Where(scope.Query, scope.Args...)
}
