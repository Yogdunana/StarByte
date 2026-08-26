package engine

import (
	"context"
	"encoding/json"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/model"
)

// TaskAction represents the action taken on a flow task.
type TaskAction string

const (
	ActionApprove  TaskAction = "approve"
	ActionReject   TaskAction = "reject"
	ActionTransfer TaskAction = "transfer"
	ActionRollback TaskAction = "rollback"
	ActionWithdraw TaskAction = "withdraw"
)

// FlowNode represents a parsed node from the graph definition.
type FlowNode struct {
	ID     string                 `json:"id"`
	Type   string                 `json:"type"`
	Label  string                 `json:"label"`
	Config map[string]interface{} `json:"config"`
}

// FlowEdge represents a parsed edge from the graph definition.
type FlowEdge struct {
	ID           string `json:"id"`
	Source       string `json:"source"`
	Target       string `json:"target"`
	SourceHandle string `json:"source_handle"` // For gateway branches
}

// FlowGraph is the parsed representation of a flow definition version's bpmn_data.
type FlowGraph struct {
	Nodes map[string]*FlowNode `json:"nodes"`
	Edges []*FlowEdge          `json:"edges"`
}

// ParseGraph deserializes bpmn_data (JSON) into a FlowGraph.
// The bpmn_data format follows the React Flow structure: { nodes: [], edges: [] }.
func ParseGraph(bpmnData []byte) (*FlowGraph, error) {
	var raw struct {
		Nodes []struct {
			ID   string                 `json:"id"`
			Type string                 `json:"type"`
			Data map[string]interface{} `json:"data"`
		} `json:"nodes"`
		Edges []struct {
			ID           string `json:"id"`
			Source       string `json:"source"`
			Target       string `json:"target"`
			SourceHandle string `json:"sourceHandle"`
		} `json:"edges"`
	}

	if err := json.Unmarshal(bpmnData, &raw); err != nil {
		return nil, err
	}

	graph := &FlowGraph{
		Nodes: make(map[string]*FlowNode, len(raw.Nodes)),
		Edges: make([]*FlowEdge, 0, len(raw.Edges)),
	}

	for _, n := range raw.Nodes {
		label, _ := n.Data["label"].(string)
		config, _ := n.Data["config"].(map[string]interface{})
		graph.Nodes[n.ID] = &FlowNode{
			ID:     n.ID,
			Type:   n.Type,
			Label:  label,
			Config: config,
		}
	}

	for _, e := range raw.Edges {
		graph.Edges = append(graph.Edges, &FlowEdge{
			ID:           e.ID,
			Source:       e.Source,
			Target:       e.Target,
			SourceHandle: e.SourceHandle,
		})
	}

	return graph, nil
}

// GetNextNodes returns the target nodes of all outgoing edges from the given node ID.
// If sourceHandle is non-empty, only edges with matching sourceHandle are returned.
func (g *FlowGraph) GetNextNodes(nodeID string, sourceHandle string) []*FlowEdge {
	var result []*FlowEdge
	for _, e := range g.Edges {
		if e.Source != nodeID {
			continue
		}
		if sourceHandle != "" && e.SourceHandle != sourceHandle {
			continue
		}
		result = append(result, e)
	}
	return result
}

// GetNode returns the node with the given ID, or nil if not found.
func (g *FlowGraph) GetNode(nodeID string) *FlowNode {
	return g.Nodes[nodeID]
}

// FindStartNode returns the first node of type "start" in the graph.
func (g *FlowGraph) FindStartNode() *FlowNode {
	for _, n := range g.Nodes {
		if n.Type == "start" {
			return n
		}
	}
	return nil
}

// GetCurrentNodeIDs extracts the current node IDs from the instance's CurrentNodeIDs JSONB field.
func GetCurrentNodeIDs(currentNodeIDsJSON []byte) []string {
	if len(currentNodeIDsJSON) == 0 {
		return nil
	}
	var nodeIDs []string
	if err := json.Unmarshal(currentNodeIDsJSON, &nodeIDs); err != nil {
		return nil
	}
	return nodeIDs
}

// NodeHandler is the interface that all node type handlers must implement.
type NodeHandler interface {
	// Type returns the node type identifier (e.g., "start", "approval").
	Type() string

	// Execute runs the node's logic and returns the IDs of outgoing edges to follow.
	// For automatic nodes (start, end, gateway), this returns the next edges.
	// For wait nodes (approval), this returns empty (flow waits for task completion).
	Execute(ctx context.Context, instance *model.FlowInstance, node *FlowNode, graph *FlowGraph, vars map[string]interface{}) ([]string, error)

	// OnEnter is called when the flow enters the node.
	OnEnter(ctx context.Context, instance *model.FlowInstance, node *FlowNode, vars map[string]interface{}) error

	// OnLeave is called when the flow leaves the node.
	OnLeave(ctx context.Context, instance *model.FlowInstance, node *FlowNode, vars map[string]interface{}) error

	// Validate checks that the node's configuration is valid.
	Validate(node *FlowNode) error
}
