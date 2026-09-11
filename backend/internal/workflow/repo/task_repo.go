package repo

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/model"
)

// TaskRepo manages flow_tasks and flow_histories.
type TaskRepo interface {
	// CreateTask creates a new flow task.
	CreateTask(ctx context.Context, tx *gorm.DB, task *model.FlowTask) error
	// GetTaskByID retrieves a task by its primary key.
	GetTaskByID(ctx context.Context, id uuid.UUID) (*model.FlowTask, error)
	// UpdateTask saves changes to an existing task.
	UpdateTask(ctx context.Context, tx *gorm.DB, task *model.FlowTask) error
	// ListTodoTasks returns pending tasks for a user.
	ListTodoTasks(ctx context.Context, assigneeID uuid.UUID, page, pageSize int) ([]model.FlowTask, int64, error)
	// ListDoneTasks returns completed tasks for a user.
	ListDoneTasks(ctx context.Context, assigneeID uuid.UUID, page, pageSize int) ([]model.FlowTask, int64, error)
	// ListTasksByInstance returns all tasks for an instance.
	ListTasksByInstance(ctx context.Context, instanceID uuid.UUID) ([]model.FlowTask, error)

	// CreateHistory records a flow history entry.
	CreateHistory(ctx context.Context, tx *gorm.DB, hist *model.FlowHistory) error
	// ListHistory returns the history for an instance, oldest first.
	ListHistory(ctx context.Context, instanceID uuid.UUID) ([]model.FlowHistory, error)
}

type taskRepo struct {
	db *gorm.DB
}

// NewTaskRepo creates a TaskRepo backed by the given GORM DB.
func NewTaskRepo(db *gorm.DB) TaskRepo {
	return &taskRepo{db: db}
}

func (r *taskRepo) getDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.db
}

func (r *taskRepo) CreateTask(ctx context.Context, tx *gorm.DB, task *model.FlowTask) error {
	return r.getDB(tx).WithContext(ctx).Create(task).Error
}

func (r *taskRepo) GetTaskByID(ctx context.Context, id uuid.UUID) (*model.FlowTask, error) {
	var task model.FlowTask
	err := r.taskQuery(ctx).Where("flow_tasks.id = ?", id).First(&task).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &task, err
}

func (r *taskRepo) UpdateTask(ctx context.Context, tx *gorm.DB, task *model.FlowTask) error {
	return r.getDB(tx).WithContext(ctx).Save(task).Error
}

func (r *taskRepo) ListTodoTasks(ctx context.Context, assigneeID uuid.UUID, page, pageSize int) ([]model.FlowTask, int64, error) {
	var tasks []model.FlowTask
	var total int64

	query := r.taskQuery(ctx).
		Where("flow_tasks.assignee_id = ? AND flow_tasks.status = 0", assigneeID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("flow_tasks.created_at DESC").Offset(offset).Limit(pageSize).Find(&tasks).Error
	return tasks, total, err
}

func (r *taskRepo) ListDoneTasks(ctx context.Context, assigneeID uuid.UUID, page, pageSize int) ([]model.FlowTask, int64, error) {
	var tasks []model.FlowTask
	var total int64

	query := r.taskQuery(ctx).
		Where("flow_tasks.assignee_id = ? AND flow_tasks.status > 0", assigneeID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("flow_tasks.completed_at DESC").Offset(offset).Limit(pageSize).Find(&tasks).Error
	return tasks, total, err
}

func (r *taskRepo) ListTasksByInstance(ctx context.Context, instanceID uuid.UUID) ([]model.FlowTask, error) {
	var tasks []model.FlowTask
	err := r.db.WithContext(ctx).
		Where("instance_id = ?", instanceID).
		Order("created_at ASC").
		Find(&tasks).Error
	return tasks, err
}

func (r *taskRepo) CreateHistory(ctx context.Context, tx *gorm.DB, hist *model.FlowHistory) error {
	return r.getDB(tx).WithContext(ctx).Create(hist).Error
}

func (r *taskRepo) ListHistory(ctx context.Context, instanceID uuid.UUID) ([]model.FlowHistory, error) {
	var histories []model.FlowHistory
	err := r.db.WithContext(ctx).Model(&model.FlowHistory{}).Select("flow_histories.*, actor.real_name AS operator_name").Joins("LEFT JOIN users actor ON actor.id=flow_histories.operator_id").
		Where("flow_histories.instance_id = ?", instanceID).Order("flow_histories.created_at ASC").Find(&histories).Error
	return histories, err
}

func (r *taskRepo) taskQuery(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Model(&model.FlowTask{}).Select("flow_tasks.*, actor.real_name AS assignee_name, definition.name AS definition_name, instance.status AS instance_status, instance.business_type, instance.business_key").Joins("LEFT JOIN users actor ON actor.id=flow_tasks.assignee_id").Joins("JOIN flow_instances instance ON instance.id=flow_tasks.instance_id").Joins("JOIN flow_definitions definition ON definition.id=instance.definition_id")
}
