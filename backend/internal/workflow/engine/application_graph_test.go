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
	require.NotNil(t, graph.GetNode("officer"))
	require.NotNil(t, graph.GetNode("minister"))
	require.NotNil(t, graph.GetNode("president"))
}

func TestMemberApplicationGraphRejectsEmptyApprovals(t *testing.T) {
	graph, err := ParseGraph(MemberApplicationBPMN())
	require.NoError(t, err)
	delete(graph.Nodes, "officer")
	delete(graph.Nodes, "minister")
	delete(graph.Nodes, "president")
	graph.Edges = []*FlowEdge{{Source: "start", Target: "end"}}
	require.Error(t, ValidateBusinessDefinition(MemberApplicationDefinitionKey, graph))
}

func TestMemberApplicationGraphAllowsEditedSpine(t *testing.T) {
	graph, err := ParseGraph(MemberApplicationBPMN())
	require.NoError(t, err)
	delete(graph.Nodes, "officer")
	graph.Edges = []*FlowEdge{
		{Source: "start", Target: "minister"},
		{Source: "minister", Target: "president"},
		{Source: "president", Target: "end"},
	}
	require.NoError(t, ValidateBusinessDefinition(MemberApplicationDefinitionKey, graph))
}

func TestMemberApplicationGraphAllowsCountersignAndCondition(t *testing.T) {
	graph, err := ParseGraph(MemberApplicationBPMN())
	require.NoError(t, err)
	graph.Nodes["minister"].Config["approvalType"] = "all"
	graph.Nodes["route"] = &FlowNode{ID: "route", Type: "condition", Label: "是否干事", Config: map[string]interface{}{}}
	graph.Edges = []*FlowEdge{
		{Source: "start", Target: "route"},
		{Source: "route", Target: "officer"},
		{Source: "route", Target: "minister"},
		{Source: "officer", Target: "minister"},
		{Source: "minister", Target: "president"},
		{Source: "president", Target: "end"},
	}
	require.NoError(t, ValidateBusinessDefinition(MemberApplicationDefinitionKey, graph))
}

func TestMemberApplicationGraphRequiresRoleAssignees(t *testing.T) {
	graph, err := ParseGraph(MemberApplicationBPMN())
	require.NoError(t, err)
	graph.Nodes["minister"].Config["assigneeStrategy"] = "business_role"
	require.Error(t, ValidateBusinessDefinition(MemberApplicationDefinitionKey, graph))
}
