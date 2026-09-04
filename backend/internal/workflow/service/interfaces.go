package service

import (
	"context"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/dto"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/model"
	"github.com/google/uuid"
)

// DefinitionService 流程定义服务接口
type DefinitionService interface {
	Create(ctx context.Context, req *dto.CreateDefinitionRequest, userID uuid.UUID) (*model.FlowDefinition, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.FlowDefinition, error)
	Update(ctx context.Context, id uuid.UUID, req *dto.UpdateDefinitionRequest, userID uuid.UUID) (*model.FlowDefinition, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, page, pageSize int, keyword, category string, status *int) ([]model.FlowDefinition, int64, error)
	Publish(ctx context.Context, id uuid.UUID, req *dto.PublishDefinitionRequest, userID uuid.UUID) (*model.FlowDefinitionVersion, error)
	SaveDraft(ctx context.Context, id uuid.UUID, req *dto.SaveDraftRequest, userID uuid.UUID) (*model.FlowDefinition, error)
	ListVersions(ctx context.Context, definitionID uuid.UUID) ([]model.FlowDefinitionVersion, error)
	GetVersionByID(ctx context.Context, id uuid.UUID) (*model.FlowDefinitionVersion, error)
}

// InstanceService 流程实例服务接口
type InstanceService interface {
	Start(ctx context.Context, defID uuid.UUID, businessKey, businessType string, initiatorID uuid.UUID, variables map[string]interface{}) (*model.FlowInstance, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.FlowInstance, error)
	List(ctx context.Context, page, pageSize int, status *int, definitionID, initiatorID *uuid.UUID) ([]model.FlowInstance, int64, error)
	Terminate(ctx context.Context, instanceID uuid.UUID, operatorID uuid.UUID, reason string) error
	Suspend(ctx context.Context, instanceID uuid.UUID, operatorID uuid.UUID, reason string) error
	Resume(ctx context.Context, instanceID uuid.UUID, operatorID uuid.UUID) error
	ListHistory(ctx context.Context, instanceID uuid.UUID) ([]model.FlowHistory, error)
}

// TaskService 流程任务服务接口
type TaskService interface {
	ListTodoTasks(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.FlowTask, int64, error)
	ListDoneTasks(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.FlowTask, int64, error)
	GetTaskByID(ctx context.Context, id uuid.UUID) (*model.FlowTask, error)
	CompleteTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID, action string, comment string, formData map[string]interface{}) error
	TransferTask(ctx context.Context, taskID uuid.UUID, fromUserID uuid.UUID, toUserID uuid.UUID, comment string) error
	RollbackTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID, targetNodeID string, comment string) error
}
