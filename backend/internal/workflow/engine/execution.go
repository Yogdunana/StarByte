package engine

import (
	"context"
	"sort"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/model"
	"github.com/Yogdunana/StarByte/backend/pkg/events"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

// executeFromNodes advances automatic nodes and retains every waiting branch.
func (e *FlowEngine) executeFromNodes(ctx context.Context, inst *model.FlowInstance, graph *FlowGraph, nodeIDs []string, vars map[string]interface{}, source ...string) error {
	if vars == nil {
		vars = map[string]interface{}{}
	}
	active := map[string]bool{}
	for _, id := range GetCurrentNodeIDs(inst.CurrentNodeIDs) {
		active[id] = true
	}
	type arrival struct{ id, source string }
	queue := []arrival{}
	origin := ""
	if len(source) > 0 {
		origin = source[0]
	}
	for _, id := range nodeIDs {
		queue = append(queue, arrival{id, origin})
	}
	for steps := 0; len(queue) > 0; steps++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if steps >= 1000 {
			return response.NewAppError(response.CodeWorkflowInvalidNode, "自动节点流转次数超限，请检查流程循环")
		}
		entry := queue[0]
		id := entry.id
		queue = queue[1:]
		node := graph.GetNode(id)
		if node == nil {
			return response.NewAppErrorf(response.CodeWorkflowNodeNotFound, "流程节点不存在: %s", id)
		}
		if node.Type == "approval" && active[id] {
			return response.NewAppError(response.CodeWorkflowInvalidNode, "多个并行分支不能直接进入同一审批节点，请先使用并行汇合网关")
		}
		delete(active, id)
		ready, err := arriveParallelJoin(graph, node, entry.source, vars)
		if err != nil {
			return response.NewAppError(response.CodeWorkflowInvalidNode, err.Error())
		}
		if !ready {
			active[id] = true
			continue
		}
		handler, err := e.registry.Get(node.Type)
		if err != nil {
			return err
		}
		e.eventBus.Publish(ctx, events.NodeEnteredEvent{InstanceID: inst.ID, NodeID: node.ID, NodeType: node.Type})
		if skipMinisterApproval(node, vars) {
			next, err := handler.Execute(ctx, inst, node, graph, vars)
			if err != nil {
				return err
			}
			if err = handler.OnLeave(ctx, inst, node, vars); err != nil {
				return err
			}
			e.eventBus.Publish(ctx, events.NodeLeftEvent{InstanceID: inst.ID, NodeID: node.ID, NodeType: node.Type})
			for _, nextID := range next {
				queue = append(queue, arrival{nextID, id})
			}
			continue
		}
		if err = handler.OnEnter(ctx, inst, node, vars); err != nil {
			return err
		}
		if waiting, ok := handler.(WaitingNode); ok && waiting.WaitForCompletion() {
			active[id] = true
			continue
		}
		next, err := handler.Execute(ctx, inst, node, graph, vars)
		if err != nil {
			return err
		}
		if err = handler.OnLeave(ctx, inst, node, vars); err != nil {
			return err
		}
		e.eventBus.Publish(ctx, events.NodeLeftEvent{InstanceID: inst.ID, NodeID: node.ID, NodeType: node.Type})
		if len(next) == 0 && node.Type != "end" {
			active[id] = true
		}
		for _, nextID := range next {
			queue = append(queue, arrival{nextID, id})
		}
	}
	current := make([]string, 0, len(active))
	for id := range active {
		current = append(current, id)
	}
	sort.Strings(current)
	if err := e.updateCurrentNodes(ctx, inst, current); err != nil {
		return err
	}
	if err := e.varRepo.SetMap(ctx, nil, inst.ID, vars); err != nil {
		return err
	}
	if len(current) == 0 {
		return e.completeInstance(ctx, inst)
	}
	return nil
}

// executeNode runs a single node and returns the next node IDs.
// Used by CompleteTask to continue after task approval.
// OnEnter is skipped because it was already called when the task was created.
// Lifecycle order: OnLeave → Execute → NodeLeftEvent
func (e *FlowEngine) executeNode(ctx context.Context, inst *model.FlowInstance, graph *FlowGraph, node *FlowNode, vars map[string]interface{}) ([]string, error) {
	handler, err := e.registry.Get(node.Type)
	if err != nil {
		return nil, err
	}

	// OnLeave.
	if err := handler.OnLeave(ctx, inst, node, vars); err != nil {
		return nil, err
	}

	// Execute to get next nodes.
	nextNodeIDs, err := handler.Execute(ctx, inst, node, graph, vars)
	if err != nil {
		return nil, err
	}

	// Publish NodeLeftEvent for the completed node.
	e.eventBus.Publish(ctx, events.NodeLeftEvent{
		InstanceID: inst.ID,
		NodeID:     node.ID,
		NodeType:   node.Type,
	})

	return nextNodeIDs, nil
}
