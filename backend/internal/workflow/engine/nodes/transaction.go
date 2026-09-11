package nodes

import (
	"gorm.io/gorm"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/engine"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/events"
)

type transactionalHandler interface {
	WithRuntime(*gorm.DB, *events.EventBus) engine.NodeHandler
}

func (r *NodeRegistry) ForTransaction(tx *gorm.DB, bus *events.EventBus) engine.NodeRegistry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	bound := NewNodeRegistry()
	for key, handler := range r.handlers {
		if transactional, ok := handler.(transactionalHandler); ok {
			handler = transactional.WithRuntime(tx, bus)
		}
		bound.handlers[key] = handler
	}
	return bound
}
func (n *ApprovalNode) WithRuntime(tx *gorm.DB, bus *events.EventBus) engine.NodeHandler {
	copy := *n
	copy.TaskRepo = repo.NewTaskRepo(tx)
	copy.Approvers = repo.NewApproverRepo(tx)
	copy.EventBus = bus
	if n.BusinessApprovers != nil {
		copy.BusinessApprovers = n.BusinessApprovers.ForTransaction(tx)
	}
	return &copy
}
func (n *NotificationTaskNode) WithRuntime(_ *gorm.DB, bus *events.EventBus) engine.NodeHandler {
	copy := *n
	copy.EventBus = bus
	return &copy
}
