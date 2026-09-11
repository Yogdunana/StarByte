package engine

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/model"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/events"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

// TransactionRegistry rebinds handlers that persist tasks to the same unit of work.
type TransactionRegistry interface {
	ForTransaction(*gorm.DB, *events.EventBus) NodeRegistry
}

func (e *FlowEngine) transaction(ctx context.Context, id uuid.UUID, operation func(*FlowEngine) error) error {
	if id != uuid.Nil && !e.businessTransaction {
		instance, err := e.instRepo.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if instance != nil && IsProtectedBusiness(instance.BusinessType) {
			return response.NewAppError(response.CodeForbidden, "业务流程须在对应业务面板操作，不能通过通用流程接口修改")
		}
	}
	if e.db == nil || e.inTransaction {
		return operation(e)
	}
	bus, buffer := events.NewBufferedBus()
	err := e.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if id != uuid.Nil {
			if err := repo.NewRuntimeRepo(tx).LockInstance(ctx, id); err != nil {
				return err
			}
		}
		bound, err := e.bindRuntime(tx, bus)
		if err != nil {
			return err
		}
		return operation(bound)
	})
	if err == nil {
		for _, err := range buffer.Flush(ctx, e.eventBus) {
			if e.logger != nil {
				e.logger.Error("workflow event delivery failed", zap.Error(err))
			}
		}
	}
	return err
}
func (e *FlowEngine) Start(ctx context.Context, key, businessKey, businessType string, initiator uuid.UUID, variables map[string]interface{}) (*model.FlowInstance, error) {
	var result *model.FlowInstance
	err := e.transaction(ctx, uuid.Nil, func(bound *FlowEngine) error {
		var err error
		result, err = bound.start(ctx, key, businessKey, businessType, initiator, variables)
		return err
	})
	return result, err
}
func (e *FlowEngine) CompleteTask(ctx context.Context, taskID, user uuid.UUID, action TaskAction, comment string, form map[string]interface{}) error {
	task, err := e.taskRepo.GetTaskByID(ctx, taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return response.NewAppError(response.CodeWorkflowTaskNotFnd, "流程任务不存在")
	}
	return e.transaction(ctx, task.InstanceID, func(bound *FlowEngine) error {
		// Re-read task and authority after acquiring the instance lock.
		return bound.completeTask(ctx, taskID, user, action, comment, form)
	})
}

func (e *FlowEngine) Terminate(ctx context.Context, id, user uuid.UUID, reason string) error {
	return e.transaction(ctx, id, func(bound *FlowEngine) error { return bound.terminate(ctx, id, user, reason) })
}
func (e *FlowEngine) Suspend(ctx context.Context, id, user uuid.UUID, reason string) error {
	return e.transaction(ctx, id, func(bound *FlowEngine) error { return bound.suspend(ctx, id, user, reason) })
}
func (e *FlowEngine) Resume(ctx context.Context, id, user uuid.UUID) error {
	return e.transaction(ctx, id, func(bound *FlowEngine) error { return bound.resume(ctx, id, user) })
}
