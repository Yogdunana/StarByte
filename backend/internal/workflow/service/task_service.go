package service

import (
	"context"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/engine"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/model"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// taskServiceImpl handles flow task business logic.
type taskServiceImpl struct {
	taskRepo   repo.TaskRepo
	instRepo   repo.InstanceRepo
	flowEngine *engine.FlowEngine
	db         *gorm.DB
}

// NewTaskService creates a TaskService.
func NewTaskService(taskRepo repo.TaskRepo, instRepo repo.InstanceRepo, flowEngine *engine.FlowEngine, db *gorm.DB) TaskService {
	return &taskServiceImpl{
		taskRepo:   taskRepo,
		instRepo:   instRepo,
		flowEngine: flowEngine,
		db:         db,
	}
}

// ListTodoTasks returns pending tasks for a user.
func (s *taskServiceImpl) ListTodoTasks(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.FlowTask, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	tasks, total, err := s.taskRepo.ListTodoTasks(ctx, userID, page, pageSize)
	if err != nil {
		return nil, 0, response.NewAppErrorf(response.CodeInternalError,
			"failed to list todo tasks: %v", err)
	}
	return tasks, total, nil
}

// ListDoneTasks returns completed tasks for a user.
func (s *taskServiceImpl) ListDoneTasks(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.FlowTask, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	tasks, total, err := s.taskRepo.ListDoneTasks(ctx, userID, page, pageSize)
	if err != nil {
		return nil, 0, response.NewAppErrorf(response.CodeInternalError,
			"failed to list done tasks: %v", err)
	}
	return tasks, total, nil
}

// GetTaskByID retrieves a task by ID.
func (s *taskServiceImpl) GetTaskByID(ctx context.Context, id uuid.UUID) (*model.FlowTask, error) {
	task, err := s.taskRepo.GetTaskByID(ctx, id)
	if err != nil {
		return nil, response.NewAppErrorf(response.CodeInternalError,
			"failed to query task: %v", err)
	}
	if task == nil {
		return nil, response.NewAppError(response.CodeWorkflowTaskNotFnd,
			"流程任务不存在")
	}
	return task, nil
}

// CompleteTask completes a flow task with the given action.
func (s *taskServiceImpl) CompleteTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID, action string, comment string, formData map[string]interface{}) error {
	taskAction := engine.TaskAction(action)
	if taskAction != engine.ActionApprove &&
		taskAction != engine.ActionReject &&
		taskAction != engine.ActionTransfer &&
		taskAction != engine.ActionWithdraw {
		return response.NewAppError(response.CodeBadRequest,
			"无效的操作类型")
	}

	return s.flowEngine.CompleteTask(ctx, taskID, userID, taskAction, comment, formData)
}

// TransferTask transfers a task to another user.
func (s *taskServiceImpl) TransferTask(ctx context.Context, taskID uuid.UUID, fromUserID uuid.UUID, toUserID uuid.UUID, comment string) error {
	task, err := s.taskRepo.GetTaskByID(ctx, taskID)
	if err != nil || task == nil {
		return response.NewAppError(response.CodeWorkflowTaskNotFnd,
			"流程任务不存在")
	}
	if task.Status != 0 {
		return response.NewAppError(response.CodeWorkflowTaskStatus,
			"流程任务状态不允许操作")
	}
	if task.AssigneeID == nil || *task.AssigneeID != fromUserID {
		return response.NewAppError(response.CodeWorkflowTaskNoAccess,
			"无权操作流程任务")
	}

	// Update task with new assignee.
	now := time.Now()
	task.AssigneeID = &toUserID
	task.Action = "transfer"
	task.Comment = comment
	task.UpdatedAt = now

	// Record history.
	hist := &model.FlowHistory{
		ID:         uuid.New(),
		InstanceID: task.InstanceID,
		TaskID:     &task.ID,
		NodeID:     task.NodeID,
		NodeName:   task.NodeName,
		NodeType:   "approval",
		OperatorID: &fromUserID,
		Action:     "transfer",
		Comment:    comment,
		CreatedAt:  now,
	}

	txErr := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := s.taskRepo.UpdateTask(ctx, tx, task); err != nil {
			return err
		}
		if err := s.taskRepo.CreateHistory(ctx, tx, hist); err != nil {
			return err
		}
		return nil
	})

	return txErr
}

// RollbackTask rolls back a task to a previous node.
func (s *taskServiceImpl) RollbackTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID, targetNodeID string, comment string) error {
	task, err := s.taskRepo.GetTaskByID(ctx, taskID)
	if err != nil || task == nil {
		return response.NewAppError(response.CodeWorkflowTaskNotFnd,
			"流程任务不存在")
	}
	if task.Status != 0 {
		return response.NewAppError(response.CodeWorkflowTaskStatus,
			"流程任务状态不允许操作")
	}
	if task.AssigneeID == nil || *task.AssigneeID != userID {
		return response.NewAppError(response.CodeWorkflowTaskNoAccess,
			"无权操作流程任务")
	}

	// Cancel the current task and create a new task at the target node.
	now := time.Now()
	task.Status = 4 // withdrawn/rolled back
	task.Action = "rollback"
	task.Comment = comment
	task.CompletedAt = &now
	task.UpdatedAt = now

	hist := &model.FlowHistory{
		ID:         uuid.New(),
		InstanceID: task.InstanceID,
		TaskID:     &task.ID,
		NodeID:     task.NodeID,
		NodeName:   task.NodeName,
		NodeType:   "approval",
		OperatorID: &userID,
		Action:     "rollback",
		Comment:    comment,
		FromNodeID: task.NodeID,
		ToNodeID:   targetNodeID,
		CreatedAt:  now,
	}

	txErr := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := s.taskRepo.UpdateTask(ctx, tx, task); err != nil {
			return err
		}
		if err := s.taskRepo.CreateHistory(ctx, tx, hist); err != nil {
			return err
		}
		return nil
	})

	return txErr
}
