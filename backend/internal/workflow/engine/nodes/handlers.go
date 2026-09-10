package nodes

import (
	"github.com/Yogdunana/StarByte/backend/internal/workflow/engine"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/events"
)

// --- DefaultRegistry ---

// NewDefaultRegistry creates a NodeRegistry with all built-in node handlers registered.
// Callbacks for service tasks and the event bus are wired by the caller.
func NewDefaultRegistry(exprEngine *engine.ExpressionEngine, eventBus *events.EventBus, taskRepo repo.TaskRepo) *NodeRegistry {
	registry := NewNodeRegistry()
	registry.Register(StartNode{})
	registry.Register(EndNode{})
	registry.Register(&ApprovalNode{
		TaskRepo: taskRepo,
		EventBus: eventBus,
	})
	registry.Register(&ExclusiveGatewayNode{
		ExprEngine: exprEngine,
	})
	registry.Register(ParallelGatewayNode{})
	registry.Register(&ServiceTaskNode{
		Callbacks: make(map[string]ServiceCallback),
	})
	registry.Register(&NotificationTaskNode{
		EventBus: eventBus,
	})
	return registry
}
