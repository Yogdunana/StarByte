package service

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/engine"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/model"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
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

	viewer := model.ViewerFromContext(ctx)
	if viewer.ID == uuid.Nil {
		return nil, response.NewAppError(response.CodeForbidden, "缺少任务访问身份")
	}
	if task.AssigneeID == nil || *task.AssigneeID != viewer.ID {
		inst, err := s.instRepo.GetByID(ctx, task.InstanceID)
		if err != nil {
			return nil, err
		}
		if inst == nil {
			return nil, response.NewAppError(response.CodeWorkflowInstNotFound, "流程实例不存在")
		}
		if err := requireInstanceAccess(ctx, s.db, inst, true); err != nil {
			return nil, err
		}
	}
	return task, nil
}

// CompleteTask completes a flow task with the given action.
func (s *taskServiceImpl) CompleteTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID, action string, comment string, formData map[string]interface{}) error {
	taskAction := engine.TaskAction(action)
	if taskAction != engine.ActionApprove &&
		taskAction != engine.ActionReject &&
		taskAction != engine.ActionWithdraw {
		return response.NewAppError(response.CodeBadRequest,
			"无效的操作类型")
	}

	return s.flowEngine.CompleteTask(ctx, taskID, userID, taskAction, comment, formData)
}

// TransferTask transfers a task while retaining the original assignment history.
func (s *taskServiceImpl) TransferTask(ctx context.Context, taskID, fromUserID, toUserID uuid.UUID, comment string) error {
	return s.flowEngine.TransferTask(ctx, taskID, fromUserID, toUserID, comment)
}

// RollbackTask re-enters a previously approved node and creates fresh pending tasks.
func (s *taskServiceImpl) RollbackTask(ctx context.Context, taskID, userID uuid.UUID, targetNodeID, comment string) error {
	return s.flowEngine.RollbackTask(ctx, taskID, userID, targetNodeID, comment)
}

func (s *taskServiceImpl) TransferCandidates(ctx context.Context, id uuid.UUID, keyword string) ([]model.ApproverOption, error) {
	task, err := s.GetTaskByID(ctx, id)
	if err != nil {
		return nil, err
	}
	viewer := model.ViewerFromContext(ctx)
	if task.Status != 0 || task.AssigneeID == nil || *task.AssigneeID != viewer.ID {
		return nil, response.NewAppError(response.CodeForbidden, "仅当前处理人可选择接收人")
	}
	return repo.NewApproverRepo(s.db).Search(ctx, keyword, viewer.ID)
}
