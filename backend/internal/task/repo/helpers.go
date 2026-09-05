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

func allowedSort(sortBy, sortOrder string) (string, string) {
	cols := map[string]string{
		"created_at": "t.created_at",
		"updated_at": "t.updated_at",
		"due_date":   "t.due_date",
		"priority":   "t.priority",
		"status":     "t.status",
		"sort_order": "t.sort_order",
		"title":      "t.title",
	}
	col, ok := cols[sortBy]
	if !ok {
		col = "t.created_at"
	}
	order := "DESC"
	if sortOrder == "asc" || sortOrder == "ASC" {
		order = "ASC"
	}
	return col, order
}
