package engine

import (
	"context"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/events"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

// BindTransaction joins a business transaction. The caller must invoke deliver
// only after that transaction commits; a rollback must discard it.
func (e *FlowEngine) BindTransaction(tx *gorm.DB) (*FlowEngine, func(context.Context), error) {
	if tx == nil {
		return nil, nil, response.NewAppError(response.CodeInternalError, "工作流需要业务事务")
	}
	bus, buffer := events.NewBufferedBus()
	bound, err := e.bindRuntime(tx, bus)
	if err != nil {
		return nil, nil, err
	}
	bound.businessTransaction = true
	deliver := func(ctx context.Context) {
		for _, err := range buffer.Flush(ctx, e.eventBus) {
			if e.logger != nil {
				e.logger.Error("workflow event delivery failed", zap.Error(err))
			}
		}
	}
	return bound, deliver, nil
}
func (e *FlowEngine) bindRuntime(tx *gorm.DB, bus *events.EventBus) (*FlowEngine, error) {
	registry, ok := e.registry.(TransactionRegistry)
	if !ok {
		return nil, response.NewAppError(response.CodeWorkflowNodeType, "节点注册表未提供事务绑定")
	}
	bound := *e
	bound.db, bound.inTransaction, bound.eventBus = tx, true, bus
	bound.defRepo = repo.NewDefinitionRepo(tx)
	bound.instRepo = repo.NewInstanceRepo(tx)
	bound.taskRepo = repo.NewTaskRepo(tx)
	bound.varRepo = repo.NewVariableRepo(tx)
	bound.registry = registry.ForTransaction(tx, bus)
	return &bound, nil
}
