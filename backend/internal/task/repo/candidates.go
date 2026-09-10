package repo

import (
	"context"
	"strings"

	"github.com/Yogdunana/StarByte/backend/internal/task/model"
)

func (r *taskRepo) SearchUsers(ctx context.Context, keyword string) ([]model.NamedUser, error) {
	var rows []model.NamedUser
	q := r.db.WithContext(ctx).Table("users").Select("id,real_name,username").Where("status=0 AND deleted_at IS NULL")
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("real_name ILIKE ? OR username ILIKE ?", like, like)
	}
	err := q.Order("real_name,username,id").Limit(20).Find(&rows).Error
	return rows, err
}
