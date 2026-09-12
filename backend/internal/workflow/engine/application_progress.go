package engine

import (
	"context"

	"github.com/google/uuid"

	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

// ApplicationStep is one visible node on the membership approval timeline.
type ApplicationStep struct {
	ID            string `json:"id"`
	Label         string `json:"label"`
	Type          string `json:"type"`
	Role          string `json:"role,omitempty"`
	State         string `json:"state"`
	ApprovalType  string `json:"approval_type,omitempty"`
	AllowTransfer bool   `json:"allow_transfer,omitempty"`
}

// ApplicationProgress is the instance projection used by the front-end panel.
type ApplicationProgress struct {
	InstanceID     uuid.UUID         `json:"instance_id"`
	Status         int               `json:"status"`
	CurrentNodeIDs []string          `json:"current_node_ids"`
	Completed      bool              `json:"completed"`
	Terminated     bool              `json:"terminated"`
	Steps          []ApplicationStep `json:"steps"`
}

// ApplicationProgress loads the published graph for an instance and marks each
// visible node as done / current / pending / skipped.
func (e *FlowEngine) ApplicationProgress(ctx context.Context, instanceID uuid.UUID) (*ApplicationProgress, error) {
	inst, err := e.instRepo.GetByID(ctx, instanceID)
	if err != nil {
		return nil, err
	}
	if inst == nil || inst.BusinessType != "member_application" {
		return nil, response.NewError(response.CodeWorkflowInstNotFound, "入会流程实例不存在")
	}
	version, err := e.defRepo.GetVersionByID(ctx, inst.DefinitionVersionID)
	if err != nil {
		return nil, err
	}
	if version == nil {
		return nil, response.NewError(response.CodeWorkflowVerNotFound, "审批流程版本不存在")
	}
	graph, err := ParseGraph(version.BpmnData)
	if err != nil {
		return nil, err
	}
	history, err := e.taskRepo.ListHistory(ctx, instanceID)
	if err != nil {
		return nil, err
	}
	tasks, err := e.taskRepo.ListTasksByInstance(ctx, instanceID)
	if err != nil {
		return nil, err
	}
	current := map[string]bool{}
	for _, id := range GetCurrentNodeIDs(inst.CurrentNodeIDs) {
		current[id] = true
	}
	acted := map[string]string{}
	hasTask := map[string]bool{}
	for _, item := range history {
		switch item.Action {
		case "approve", "complete":
			acted[item.NodeID] = "done"
		case "skip":
			if acted[item.NodeID] == "" {
				acted[item.NodeID] = "skipped"
			}
		case "reject", "terminate":
			acted[item.NodeID] = "done"
		}
	}
	for _, task := range tasks {
		hasTask[task.NodeID] = true
		if task.Status == 1 || task.Action == "approve" {
			acted[task.NodeID] = "done"
		}
	}
	future := reachableNodeIDs(graph, GetCurrentNodeIDs(inst.CurrentNodeIDs))
	steps := make([]ApplicationStep, 0, len(graph.Nodes))
	for _, id := range displayNodeOrder(graph) {
		item := graph.GetNode(id)
		if item == nil {
			continue
		}
		step := ApplicationStep{
			ID: item.ID, Label: item.Label, Type: item.Type,
			Role: approvalRoleCode(item), ApprovalType: approvalTypeOf(item),
			AllowTransfer: item.Type == "approval" && nodeAllowsTransfer(item),
		}
		step.State = applicationStepState(inst.Status, item, current[id], acted[id], hasTask[id], future[id])
		steps = append(steps, step)
	}
	return &ApplicationProgress{
		InstanceID: inst.ID, Status: inst.Status, CurrentNodeIDs: GetCurrentNodeIDs(inst.CurrentNodeIDs),
		Completed: inst.Status == 1, Terminated: inst.Status == 2, Steps: steps,
	}, nil
}

func applicationStepState(status int, item *FlowNode, isCurrent bool, acted string, hasTask, future bool) string {
	if item == nil {
		return "pending"
	}
	switch {
	case status == 1:
		if acted == "skipped" || (item.Type == "approval" && !hasTask && acted != "done") {
			return "skipped"
		}
		return "done"
	case isCurrent:
		return "current"
	case acted != "":
		return acted
	case future:
		return "pending"
	case item.Type == "approval":
		return "skipped"
	default:
		return "done"
	}
}

func displayNodeOrder(g *FlowGraph) []string {
	start := g.FindStartNode()
	if start == nil {
		return nil
	}
	order := []string{}
	for _, id := range walkNodeIDs(g, []string{start.ID}) {
		node := g.GetNode(id)
		if node != nil && visibleProgressNode(node.Type) {
			order = append(order, id)
		}
	}
	return order
}

func visibleProgressNode(kind string) bool {
	switch kind {
	case "start", "end", "approval", "condition", "exclusive_gateway", "parallel_gateway", "parallel", "merge":
		return true
	default:
		return false
	}
}

func reachableNodeIDs(g *FlowGraph, starts []string) map[string]bool {
	seen := map[string]bool{}
	for _, id := range walkNodeIDs(g, starts) {
		seen[id] = true
	}
	return seen
}

func walkNodeIDs(g *FlowGraph, starts []string) []string {
	if g == nil {
		return nil
	}
	seen := map[string]bool{}
	order := []string{}
	queue := append([]string{}, starts...)
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		order = append(order, id)
		for _, edge := range g.GetNextNodes(id, "") {
			queue = append(queue, edge.Target)
		}
	}
	return order
}

// lastApplicationApproval is true when no other approval node is reachable
// after nodeID (condition / gateway branches included).
func lastApplicationApproval(g *FlowGraph, nodeID string) bool {
	if g == nil || nodeID == "" || g.GetNode(nodeID) == nil {
		return false
	}
	starts := []string{}
	for _, edge := range g.GetNextNodes(nodeID, "") {
		starts = append(starts, edge.Target)
	}
	for _, id := range walkNodeIDs(g, starts) {
		node := g.GetNode(id)
		if node != nil && node.Type == "approval" {
			return false
		}
	}
	return true
}
