package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseGraph_ValidGraph(t *testing.T) {
	bpmnData := []byte(`{
		"nodes": [
			{"id": "n1", "type": "start", "data": {"label": "开始"}},
			{"id": "n2", "type": "approval", "data": {"label": "审批", "config": {"assigneeStrategy": "static", "assignees": ["user-1"]}}},
			{"id": "n3", "type": "end", "data": {"label": "结束"}}
		],
		"edges": [
			{"id": "e1", "source": "n1", "target": "n2"},
			{"id": "e2", "source": "n2", "target": "n3"}
		]
	}`)

	graph, err := ParseGraph(bpmnData)
	require.NoError(t, err)
	assert.Len(t, graph.Nodes, 3)
	assert.Len(t, graph.Edges, 2)

	startNode := graph.Nodes["n1"]
	assert.Equal(t, "start", startNode.Type)
	assert.Equal(t, "开始", startNode.Label)

	approvalNode := graph.Nodes["n2"]
	assert.Equal(t, "approval", approvalNode.Type)
	require.NotNil(t, approvalNode.Config)
	assert.Equal(t, "static", approvalNode.Config["assigneeStrategy"])
}

func TestParseGraph_InvalidJSON(t *testing.T) {
	bpmnData := []byte(`{invalid json}`)

	_, err := ParseGraph(bpmnData)
	require.Error(t, err)
}

func TestParseGraph_EmptyGraph(t *testing.T) {
	bpmnData := []byte(`{"nodes": [], "edges": []}`)

	graph, err := ParseGraph(bpmnData)
	require.NoError(t, err)
	assert.Empty(t, graph.Nodes)
	assert.Empty(t, graph.Edges)
}

func TestParseGraph_WithSourceHandle(t *testing.T) {
	bpmnData := []byte(`{
		"nodes": [
			{"id": "n1", "type": "exclusive_gateway", "data": {"label": "分支"}},
			{"id": "n2", "type": "end", "data": {"label": "结束A"}},
			{"id": "n3", "type": "end", "data": {"label": "结束B"}}
		],
		"edges": [
			{"id": "e1", "source": "n1", "target": "n2", "sourceHandle": "branch_a"},
			{"id": "e2", "source": "n1", "target": "n3", "sourceHandle": "branch_b"}
		]
	}`)

	graph, err := ParseGraph(bpmnData)
	require.NoError(t, err)

	edges := graph.GetNextNodes("n1", "branch_a")
	assert.Len(t, edges, 1)
	assert.Equal(t, "n2", edges[0].Target)

	edges = graph.GetNextNodes("n1", "branch_b")
	assert.Len(t, edges, 1)
	assert.Equal(t, "n3", edges[0].Target)
}

func TestFlowGraph_GetNextNodes_AllEdges(t *testing.T) {
	graph := &FlowGraph{
		Nodes: map[string]*FlowNode{
			"n1": {ID: "n1", Type: "start"},
			"n2": {ID: "n2", Type: "end"},
			"n3": {ID: "n3", Type: "end"},
		},
		Edges: []*FlowEdge{
			{ID: "e1", Source: "n1", Target: "n2"},
			{ID: "e2", Source: "n1", Target: "n3"},
		},
	}

	edges := graph.GetNextNodes("n1", "")
	assert.Len(t, edges, 2)
}

func TestFlowGraph_GetNextNodes_NoOutgoing(t *testing.T) {
	graph := &FlowGraph{
		Nodes: map[string]*FlowNode{
			"n1": {ID: "n1", Type: "end"},
		},
		Edges: []*FlowEdge{},
	}

	edges := graph.GetNextNodes("n1", "")
	assert.Empty(t, edges)
}

func TestFlowGraph_GetNode(t *testing.T) {
	graph := &FlowGraph{
		Nodes: map[string]*FlowNode{
			"n1": {ID: "n1", Type: "start", Label: "Start"},
		},
	}

	node := graph.GetNode("n1")
	require.NotNil(t, node)
	assert.Equal(t, "start", node.Type)

	node = graph.GetNode("nonexistent")
	assert.Nil(t, node)
}

func TestFlowGraph_FindStartNode(t *testing.T) {
	graph := &FlowGraph{
		Nodes: map[string]*FlowNode{
			"n1": {ID: "n1", Type: "approval"},
			"n2": {ID: "n2", Type: "start"},
			"n3": {ID: "n3", Type: "end"},
		},
	}

	start := graph.FindStartNode()
	require.NotNil(t, start)
	assert.Equal(t, "start", start.Type)
	assert.Equal(t, "n2", start.ID)
}

func TestFlowGraph_FindStartNode_NoneFound(t *testing.T) {
	graph := &FlowGraph{
		Nodes: map[string]*FlowNode{
			"n1": {ID: "n1", Type: "approval"},
			"n2": {ID: "n2", Type: "end"},
		},
	}

	start := graph.FindStartNode()
	assert.Nil(t, start)
}

func TestTaskAction_Constants(t *testing.T) {
	assert.Equal(t, TaskAction("approve"), ActionApprove)
	assert.Equal(t, TaskAction("reject"), ActionReject)
	assert.Equal(t, TaskAction("transfer"), ActionTransfer)
	assert.Equal(t, TaskAction("rollback"), ActionRollback)
	assert.Equal(t, TaskAction("withdraw"), ActionWithdraw)
}
