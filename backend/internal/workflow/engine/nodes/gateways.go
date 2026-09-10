package nodes

import (
	"context"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/engine"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

// --- ExclusiveGatewayNode ---

// ExclusiveGatewayNode handles the "exclusive_gateway" node type.
// It evaluates branch expressions and selects exactly one outgoing path.
type ExclusiveGatewayNode struct {
	ExprEngine *engine.ExpressionEngine
}

func (n *ExclusiveGatewayNode) Type() string { return "exclusive_gateway" }

func (n *ExclusiveGatewayNode) Execute(ctx context.Context, inst *model.FlowInstance, node *engine.FlowNode, graph *engine.FlowGraph, vars map[string]interface{}) ([]string, error) {
	branches, err := engine.ParseBranches(node.Config)
	if err != nil {
		return nil, err
	}

	branchID, err := n.ExprEngine.EvaluateBranch(branches, vars)
	if err != nil {
		return nil, err
	}

	// Find the edge whose sourceHandle matches the selected branch ID.
	edges := graph.GetNextNodes(node.ID, branchID)

	result := make([]string, 0, len(edges))
	for _, e := range edges {
		result = append(result, e.Target)
	}

	if len(result) == 0 {
		return nil, response.NewAppErrorf(response.CodeWorkflowNodeNotFound,
			"排他网关 '%s' 没有分支 '%s' 的出线", node.ID, branchID)
	}

	return result, nil
}

func (n *ExclusiveGatewayNode) OnEnter(ctx context.Context, inst *model.FlowInstance, node *engine.FlowNode, vars map[string]interface{}) error {
	return nil
}

func (n *ExclusiveGatewayNode) OnLeave(ctx context.Context, inst *model.FlowInstance, node *engine.FlowNode, vars map[string]interface{}) error {
	return nil
}

func (n *ExclusiveGatewayNode) Validate(node *engine.FlowNode) error {
	if node.Config == nil {
		return response.NewAppError(response.CodeWorkflowInvalidNode,
			"排他网关缺少配置")
	}

	branches, err := engine.ParseBranches(node.Config)
	if err != nil {
		return err
	}

	if len(branches) < 2 {
		return response.NewAppError(response.CodeWorkflowInvalidNode,
			"排他网关至少需要 2 个分支")
	}

	return nil
}

// --- ParallelGatewayNode ---

// ParallelGatewayNode handles the "parallel_gateway" node type.
// All outgoing branches are activated simultaneously.
type ParallelGatewayNode struct{}

func (ParallelGatewayNode) Type() string { return "parallel_gateway" }

func (ParallelGatewayNode) Execute(ctx context.Context, inst *model.FlowInstance, node *engine.FlowNode, graph *engine.FlowGraph, vars map[string]interface{}) ([]string, error) {
	// Activate all outgoing edges.
	edges := graph.GetNextNodes(node.ID, "")
	result := make([]string, 0, len(edges))
	for _, e := range edges {
		result = append(result, e.Target)
	}

	if len(result) == 0 {
		return nil, response.NewAppErrorf(response.CodeWorkflowNodeNotFound,
			"并行网关 '%s' 没有出线", node.ID)
	}

	return result, nil
}

func (ParallelGatewayNode) OnEnter(ctx context.Context, inst *model.FlowInstance, node *engine.FlowNode, vars map[string]interface{}) error {
	return nil
}

func (ParallelGatewayNode) OnLeave(ctx context.Context, inst *model.FlowInstance, node *engine.FlowNode, vars map[string]interface{}) error {
	return nil
}

func (ParallelGatewayNode) Validate(node *engine.FlowNode) error {
	// Parallel gateway validation is done at graph level (must have >= 2 outgoing edges).
	return nil
}
