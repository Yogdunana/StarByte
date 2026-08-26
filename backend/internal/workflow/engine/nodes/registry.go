package nodes

import (
	"context"
	"fmt"
	"sync"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/engine"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

// NodeRegistry holds registered node handlers by type.
type NodeRegistry struct {
	mu       sync.RWMutex
	handlers map[string]engine.NodeHandler
}

// NewNodeRegistry creates a new NodeRegistry.
func NewNodeRegistry() *NodeRegistry {
	return &NodeRegistry{
		handlers: make(map[string]engine.NodeHandler),
	}
}

// Register registers a node handler.
// Panics if a handler for the same type is already registered.
func (r *NodeRegistry) Register(handler engine.NodeHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()

	t := handler.Type()
	if _, exists := r.handlers[t]; exists {
		panic(fmt.Sprintf("node handler for type '%s' already registered", t))
	}
	r.handlers[t] = handler
}

// Get returns the handler for the given node type, or an error if not found.
func (r *NodeRegistry) Get(nodeType string) (engine.NodeHandler, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	h, ok := r.handlers[nodeType]
	if !ok {
		return nil, response.NewAppErrorf(response.CodeWorkflowNodeType,
			"node type '%s' is not supported", nodeType)
	}
	return h, nil
}

// MustGet returns the handler for the given node type, panicking if not found.
func (r *NodeRegistry) MustGet(nodeType string) engine.NodeHandler {
	h, err := r.Get(nodeType)
	if err != nil {
		panic(err)
	}
	return h
}

// --- StartNode ---

// StartNode handles the "start" node type.
// It has no special logic — the flow simply passes through.
type StartNode struct{}

func (StartNode) Type() string { return "start" }

func (StartNode) Execute(ctx context.Context, inst *model.FlowInstance, node *engine.FlowNode, graph *engine.FlowGraph, vars map[string]interface{}) ([]string, error) {
	edges := graph.GetNextNodes(node.ID, "")
	result := make([]string, 0, len(edges))
	for _, e := range edges {
		result = append(result, e.Target)
	}
	return result, nil
}

func (StartNode) OnEnter(ctx context.Context, inst *model.FlowInstance, node *engine.FlowNode, vars map[string]interface{}) error {
	return nil
}

func (StartNode) OnLeave(ctx context.Context, inst *model.FlowInstance, node *engine.FlowNode, vars map[string]interface{}) error {
	return nil
}

func (StartNode) Validate(node *engine.FlowNode) error {
	// Start node must have at least one outgoing edge — validated at graph level.
	return nil
}

// --- EndNode ---

// EndNode handles the "end" node type.
// When the flow reaches an end node, the instance is marked as completed.
type EndNode struct{}

func (EndNode) Type() string { return "end" }

func (EndNode) Execute(ctx context.Context, inst *model.FlowInstance, node *engine.FlowNode, graph *engine.FlowGraph, vars map[string]interface{}) ([]string, error) {
	// End node has no outgoing edges — signal completion.
	return []string{}, nil
}

func (EndNode) OnEnter(ctx context.Context, inst *model.FlowInstance, node *engine.FlowNode, vars map[string]interface{}) error {
	return nil
}

func (EndNode) OnLeave(ctx context.Context, inst *model.FlowInstance, node *engine.FlowNode, vars map[string]interface{}) error {
	return nil
}

func (EndNode) Validate(node *engine.FlowNode) error {
	// End node should have no outgoing edges — validated at graph level.
	return nil
}
