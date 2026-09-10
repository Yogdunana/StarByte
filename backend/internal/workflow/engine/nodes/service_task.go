package nodes

import (
	"context"
	"sync"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/engine"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

// --- ServiceTaskNode ---

// ServiceTaskNode handles the "service_task" node type.
// It calls an external API or executes business logic automatically.
// In v1, the actual service call is delegated to a callback registered by the business module.
type ServiceTaskNode struct {
	mu        sync.RWMutex
	Callbacks map[string]ServiceCallback
}

// ServiceCallback is a function that executes a service task.
type ServiceCallback func(ctx context.Context, inst *model.FlowInstance, node *engine.FlowNode, vars map[string]interface{}) (map[string]interface{}, error)

// RegisterService registers a callback for a service task.
// Safe for concurrent use.
func (n *ServiceTaskNode) RegisterService(name string, cb ServiceCallback) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.Callbacks[name] = cb
}

func (n *ServiceTaskNode) Type() string { return "service_task" }

func (n *ServiceTaskNode) Execute(ctx context.Context, inst *model.FlowInstance, node *engine.FlowNode, graph *engine.FlowGraph, vars map[string]interface{}) ([]string, error) {
	serviceName, _ := node.Config["service"].(string)
	if serviceName == "" {
		return nil, response.NewAppError(response.CodeWorkflowInvalidNode,
			"服务任务缺少 service 配置")
	}

	n.mu.RLock()
	cb, ok := n.Callbacks[serviceName]
	n.mu.RUnlock()

	if !ok {
		return nil, response.NewAppErrorf(response.CodeWorkflowNodeType,
			"服务任务 '%s' 没有注册回调函数", serviceName)
	}

	// Execute the service callback.
	result, err := cb(ctx, inst, node, vars)
	if err != nil {
		return nil, err
	}

	// Merge callback results into variables.
	for k, v := range result {
		vars[k] = v
	}

	// Proceed to next node.
	edges := graph.GetNextNodes(node.ID, "")
	resultIDs := make([]string, 0, len(edges))
	for _, e := range edges {
		resultIDs = append(resultIDs, e.Target)
	}
	return resultIDs, nil
}

func (n *ServiceTaskNode) OnEnter(ctx context.Context, inst *model.FlowInstance, node *engine.FlowNode, vars map[string]interface{}) error {
	return nil
}

func (n *ServiceTaskNode) OnLeave(ctx context.Context, inst *model.FlowInstance, node *engine.FlowNode, vars map[string]interface{}) error {
	return nil
}

func (n *ServiceTaskNode) Validate(node *engine.FlowNode) error {
	if node.Config == nil {
		return response.NewAppError(response.CodeWorkflowInvalidNode,
			"服务任务缺少配置")
	}
	if _, ok := node.Config["service"].(string); !ok {
		return response.NewAppError(response.CodeWorkflowInvalidNode,
			"服务任务缺少 'service' 字段配置")
	}
	return nil
}
