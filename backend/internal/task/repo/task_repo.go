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

type TaskRepo interface {
	Create(ctx context.Context, t *model.Task) error
	Update(ctx context.Context, t *model.Task) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Task, error)
	GetByIDWithNames(ctx context.Context, id uuid.UUID) (*model.TaskWithNames, error)
	List(ctx context.Context, req *dto.ListTaskRequest, scope *rbacModel.DataScopeCondition) ([]model.TaskWithNames, int64, error)
	ListMine(ctx context.Context, userID uuid.UUID, kind string, req *dto.MyTaskRequest) ([]model.TaskWithNames, int64, error)
	ListChildren(ctx context.Context, parentID uuid.UUID) ([]model.Task, error)
	ListDueSoon(ctx context.Context, now, until time.Time) ([]model.Task, error)
	ListOverdue(ctx context.Context, now time.Time) ([]model.Task, error)
	Stats(ctx context.Context, req *dto.StatsRequest, now time.Time) (*dto.StatsResponse, error)
	GetUser(ctx context.Context, id uuid.UUID) (*model.NamedUser, error)
	FindUsersByUsername(ctx context.Context, names []string) ([]model.NamedUser, error)
}

type taskRepo struct{ db *gorm.DB }

func NewTaskRepo(db *gorm.DB) TaskRepo {
	return &taskRepo{db: db}
}

func (r *taskRepo) Create(ctx context.Context, t *model.Task) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *taskRepo) Update(ctx context.Context, t *model.Task) error {
	return r.db.WithContext(ctx).Save(t).Error
}

func (r *taskRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Task{}, "id = ?", id).Error
}

func (r *taskRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Task, error) {
	var t model.Task
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&t).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *taskRepo) namedQuery(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Table("tasks AS t").
		Select(`t.*,
			COALESCE(c.real_name, c.username, '') AS creator_name,
			COALESCE(a.real_name, a.username, '') AS assignee_name,
			COALESCE(a.avatar_url, '') AS assignee_avatar,
			COALESCE(d.name, '') AS department_name,
			COALESCE(p.title, '') AS parent_title,
			(SELECT COUNT(*) FROM task_comments tc WHERE tc.task_id = t.id) AS comment_count,
			(SELECT COUNT(*) FROM task_attachments ta WHERE ta.task_id = t.id) AS attachment_count`).
		Joins("LEFT JOIN users c ON c.id = t.creator_id").
		Joins("LEFT JOIN users a ON a.id = t.assignee_id").
		Joins("LEFT JOIN departments d ON d.id = t.department_id").
		Joins("LEFT JOIN tasks p ON p.id = t.parent_id")
}

func (r *taskRepo) GetByIDWithNames(ctx context.Context, id uuid.UUID) (*model.TaskWithNames, error) {
	var row model.TaskWithNames
	err := r.namedQuery(ctx).Where("t.id = ?", id).First(&row).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *taskRepo) List(ctx context.Context, req *dto.ListTaskRequest, scope *rbacModel.DataScopeCondition) ([]model.TaskWithNames, int64, error) {
	q := applyScope(r.namedQuery(ctx), scope)
	q = applyListFilters(q, req)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	page, size := normalizePage(req.Page, req.PageSize)
	col, order := allowedSort(req.SortBy, req.SortOrder)
	var rows []model.TaskWithNames
	err := q.Order(col + " " + order).Offset((page - 1) * size).Limit(size).Find(&rows).Error
	return rows, total, err
}

func applyListFilters(q *gorm.DB, req *dto.ListTaskRequest) *gorm.DB {
	if req.Status != nil {
		q = q.Where("t.status = ?", *req.Status)
	}
	if req.Priority != nil {
		q = q.Where("t.priority = ?", *req.Priority)
	}
	if req.AssigneeID != "" {
		q = q.Where("t.assignee_id = ?", req.AssigneeID)
	}
	if req.DepartmentID != "" {
		q = q.Where("t.department_id = ?", req.DepartmentID)
	}
	if kw := strings.TrimSpace(req.Keyword); kw != "" {
		like := "%" + kw + "%"
		q = q.Where("t.title ILIKE ? OR t.description ILIKE ?", like, like)
	}
	if tag := strings.TrimSpace(req.Tags); tag != "" {
		q = q.Where("t.tags ILIKE ?", "%"+tag+"%")
	}
	if req.ParentID != "" {
		q = q.Where("t.parent_id = ?", req.ParentID)
	}
	return q
}
