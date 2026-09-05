package repo

import (
	"context"
	"strings"
	"time"

	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/internal/task/dto"
	"github.com/Yogdunana/StarByte/backend/internal/task/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (r *taskRepo) ListMine(ctx context.Context, userID uuid.UUID, kind string, req *dto.MyTaskRequest) ([]model.TaskWithNames, int64, error) {
	q := r.namedQuery(ctx)
	now := time.Now()
	switch kind {
	case "todo":
		q = q.Where("t.assignee_id = ? AND t.status IN ?", userID, []int16{model.StatusPending, model.StatusDoing, model.StatusHeld})
	case "done":
		q = q.Where("t.assignee_id = ? AND t.status = ?", userID, model.StatusDone)
	case "created":
		q = q.Where("t.creator_id = ?", userID)
		if req.Status != nil {
			q = q.Where("t.status = ?", *req.Status)
		}
	case "overdue":
		q = q.Where("t.assignee_id = ? AND t.due_date IS NOT NULL AND t.due_date < ? AND t.status IN ?",
			userID, now, []int16{model.StatusPending, model.StatusDoing, model.StatusHeld})
	default:
		q = q.Where("t.assignee_id = ?", userID)
	}
	if req.Priority != nil {
		q = q.Where("t.priority = ?", *req.Priority)
	}
	if kw := strings.TrimSpace(req.Keyword); kw != "" {
		like := "%" + kw + "%"
		q = q.Where("t.title ILIKE ? OR t.description ILIKE ?", like, like)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	page, size := normalizePage(req.Page, req.PageSize)
	var rows []model.TaskWithNames
	err := q.Order("t.due_date ASC NULLS LAST, t.created_at DESC").
		Offset((page - 1) * size).Limit(size).Find(&rows).Error
	return rows, total, err
}

func (r *taskRepo) ListChildren(ctx context.Context, parentID uuid.UUID) ([]model.Task, error) {
	var rows []model.Task
	err := r.db.WithContext(ctx).Where("parent_id = ?", parentID).Order("sort_order ASC, created_at ASC").Find(&rows).Error
	return rows, err
}

func (r *taskRepo) ListDueSoon(ctx context.Context, now, until time.Time) ([]model.Task, error) {
	var rows []model.Task
	err := r.db.WithContext(ctx).
		Where("due_date IS NOT NULL AND due_date > ? AND due_date <= ? AND due_reminded_at IS NULL AND status IN ?",
			now, until, []int16{model.StatusPending, model.StatusDoing, model.StatusHeld}).
		Find(&rows).Error
	return rows, err
}

func (r *taskRepo) ListOverdue(ctx context.Context, now time.Time) ([]model.Task, error) {
	var rows []model.Task
	err := r.db.WithContext(ctx).
		Where("due_date IS NOT NULL AND due_date < ? AND overdue_reminded_at IS NULL AND status IN ?",
			now, []int16{model.StatusPending, model.StatusDoing, model.StatusHeld}).
		Find(&rows).Error
	return rows, err
}

func (r *taskRepo) Stats(ctx context.Context, req *dto.StatsRequest, now time.Time, scope *rbacModel.DataScopeCondition) (*dto.StatsResponse, error) {
	q := applyScope(r.db.WithContext(ctx).Model(&model.Task{}), scope)
	if req.DepartmentID != "" {
		q = q.Where("department_id = ?", req.DepartmentID)
	}
	if req.StartDate != "" {
		q = q.Where("created_at >= ?", req.StartDate)
	}
	if req.EndDate != "" {
		q = q.Where("created_at <= ?", req.EndDate+" 23:59:59")
	}
	out := &dto.StatsResponse{
		ByStatus:   map[string]int64{"pending": 0, "doing": 0, "done": 0, "cancelled": 0, "held": 0},
		ByPriority: map[string]int64{"low": 0, "medium": 0, "high": 0, "urgent": 0},
	}
	if err := q.Count(&out.Total).Error; err != nil {
		return nil, err
	}
	type pair struct {
		K int16
		C int64
	}
	var statuses []pair
	if err := q.Select("status AS k, COUNT(*) AS c").Group("status").Scan(&statuses).Error; err != nil {
		return nil, err
	}
	statusKeys := map[int16]string{0: "pending", 1: "doing", 2: "done", 3: "cancelled", 4: "held"}
	for _, p := range statuses {
		if name, ok := statusKeys[p.K]; ok {
			out.ByStatus[name] = p.C
		}
	}
	var prios []pair
	if err := q.Select("priority AS k, COUNT(*) AS c").Group("priority").Scan(&prios).Error; err != nil {
		return nil, err
	}
	prioKeys := map[int16]string{0: "low", 1: "medium", 2: "high", 3: "urgent"}
	for _, p := range prios {
		if name, ok := prioKeys[p.K]; ok {
			out.ByPriority[name] = p.C
		}
	}
	overdueQ := q.Where("due_date IS NOT NULL AND due_date < ? AND status IN ?",
		now, []int16{model.StatusPending, model.StatusDoing, model.StatusHeld})
	if err := overdueQ.Count(&out.Overdue).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (r *taskRepo) GetUser(ctx context.Context, id uuid.UUID) (*model.NamedUser, error) {
	var u model.NamedUser
	err := r.db.WithContext(ctx).Table("users").
		Select("id, real_name, username, COALESCE(avatar_url, '') AS avatar").
		Where("id = ?", id).First(&u).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *taskRepo) FindUsersByUsername(ctx context.Context, names []string) ([]model.NamedUser, error) {
	if len(names) == 0 {
		return nil, nil
	}
	var rows []model.NamedUser
	err := r.db.WithContext(ctx).Table("users").
		Select("id, real_name, username, COALESCE(avatar_url, '') AS avatar").
		Where("username IN ?", names).Find(&rows).Error
	return rows, err
}
