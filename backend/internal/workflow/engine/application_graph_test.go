package engine

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMemberApplicationGraph(t *testing.T) {
	graph, err := ParseGraph(MemberApplicationBPMN())
	require.NoError(t, err)
	require.NoError(t, ValidateBusinessDefinition(MemberApplicationDefinitionKey, graph))
	require.Equal(t, "start", graph.FindStartNode().ID)
}

func TestMemberApplicationGraphRejectsBypass(t *testing.T) {
	graph, err := ParseGraph(MemberApplicationBPMN())
	require.NoError(t, err)
	delete(graph.Nodes, "minister")
	graph.Edges = []*FlowEdge{{Source: "start", Target: "president"}, {Source: "president", Target: "end"}}
	require.Error(t, ValidateBusinessDefinition(MemberApplicationDefinitionKey, graph))
}

func TestMemberApplicationGraphRequiresRoleAssignees(t *testing.T) {
	graph, err := ParseGraph(MemberApplicationBPMN())
	require.NoError(t, err)
	graph.Nodes["minister"].Config["assigneeStrategy"] = "business_role"
	require.Error(t, ValidateBusinessDefinition(MemberApplicationDefinitionKey, graph))
}
