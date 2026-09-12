package engine

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLeaveApprovalGraph(t *testing.T) {
	graph, err := ParseGraph(LeaveApprovalBPMN())
	require.NoError(t, err)
	require.NoError(t, ValidateBusinessDefinition(LeaveDefinitionKey, graph))
	require.Equal(t, "start", graph.FindStartNode().ID)
	require.Equal(t, LeaveBusinessType, "leave_application")
	require.True(t, IsProtectedBusiness(LeaveBusinessType))
}

func TestLeaveApprovalGraphRejectsBypass(t *testing.T) {
	graph, err := ParseGraph(LeaveApprovalBPMN())
	require.NoError(t, err)
	delete(graph.Nodes, "minister")
	graph.Edges = []*FlowEdge{{Source: "start", Target: "president"}, {Source: "president", Target: "end"}}
	require.Error(t, ValidateBusinessDefinition(LeaveDefinitionKey, graph))
}

func TestLeaveApprovalGraphRequiresBusinessRole(t *testing.T) {
	graph, err := ParseGraph(LeaveApprovalBPMN())
	require.NoError(t, err)
	graph.Nodes["minister"].Config["assigneeStrategy"] = "role"
	require.Error(t, ValidateBusinessDefinition(LeaveDefinitionKey, graph))
}
