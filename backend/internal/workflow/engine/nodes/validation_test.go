package nodes

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/engine"
)

func TestPublicationRejectsBrokenGraphs(t *testing.T) {
	registry := NewDefaultRegistry(engine.NewExpressionEngine(), nil, nil)
	graph := func() *engine.FlowGraph {
		return &engine.FlowGraph{
			Nodes: map[string]*engine.FlowNode{"s": {ID: "s", Type: "start"}, "a": {ID: "a", Type: "approval", Config: map[string]interface{}{"assigneeStrategy": "initiator"}}, "e": {ID: "e", Type: "end"}},
			Edges: []*engine.FlowEdge{{Source: "s", Target: "a"}, {Source: "a", Target: "e"}},
		}
	}
	require.NoError(t, engine.ValidateGraph(graph(), registry))
	cases := []struct {
		name   string
		change func(*engine.FlowGraph)
	}{
		{"missing target", func(g *engine.FlowGraph) { g.Edges[0].Target = "missing" }},
		{"multiple starts", func(g *engine.FlowGraph) { g.Nodes["x"] = &engine.FlowNode{ID: "x", Type: "start"} }},
		{"isolated node", func(g *engine.FlowGraph) { g.Nodes["x"] = &engine.FlowNode{ID: "x", Type: "end"} }},
		{"approval without assignee", func(g *engine.FlowGraph) { g.Nodes["a"].Config = nil }},
		{"approval without exit", func(g *engine.FlowGraph) { g.Edges = g.Edges[:1] }},
		{"duplicate edge", func(g *engine.FlowGraph) { g.Edges = append(g.Edges, g.Edges[0]) }},
		{"unknown node", func(g *engine.FlowGraph) { g.Nodes["a"].Type = "unknown" }},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) { g := graph(); tt.change(g); require.Error(t, engine.ValidateGraph(g, registry)) })
	}
}
