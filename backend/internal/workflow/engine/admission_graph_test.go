package engine

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func admissionGraphForTest(officer bool) *FlowGraph {
	g := &FlowGraph{Nodes: map[string]*FlowNode{}}
	add := func(id, kind, stage, role string) {
		g.Nodes[id] = &FlowNode{ID: id, Type: kind, Config: map[string]interface{}{"assigneeStrategy": "business_role", "businessType": "member_application", "admissionStage": stage, "admissionRole": role, "approvalType": "any"}}
	}
	add("s", "start", "", "")
	add("e", "end", "", "")
	add("m", "approval", "materials", "materials")
	edges := [][2]string{{"s", "m"}, {"m", "e"}}
	if officer {
		add("f", "parallel_gateway", "", "")
		add("j", "parallel_gateway", "", "")
		add("a", "approval", "round1", "minister")
		add("b", "approval", "round1", "center")
		add("c", "approval", "round2", "center")
		add("p", "approval", "president", "president")
		edges = [][2]string{{"s", "m"}, {"m", "f"}, {"f", "a"}, {"f", "b"}, {"a", "j"}, {"b", "j"}, {"j", "c"}, {"c", "p"}, {"p", "e"}}
	}
	for _, pair := range edges {
		g.Edges = append(g.Edges, &FlowEdge{ID: pair[0] + pair[1], Source: pair[0], Target: pair[1]})
	}
	return g
}
func TestAdmissionPolicyProtectsMandatorySignatures(t *testing.T) {
	require.NoError(t, ValidateBusinessDefinition("member_admission", admissionGraphForTest(false)))
	require.NoError(t, ValidateBusinessDefinition("officer_interview", admissionGraphForTest(true)))
	t.Run("member cannot be routed through interviews", func(t *testing.T) {
		require.Error(t, ValidateBusinessDefinition("member_admission", admissionGraphForTest(true)))
	})
	t.Run("officer cannot omit interview signatures", func(t *testing.T) {
		require.Error(t, ValidateBusinessDefinition("officer_interview", admissionGraphForTest(false)))
	})
	t.Run("direct edge bypass fails", func(t *testing.T) {
		g := admissionGraphForTest(true)
		g.Edges = append(g.Edges, &FlowEdge{Source: "m", Target: "e"})
		require.Error(t, ValidateBusinessDefinition("officer_interview", g))
	})
	t.Run("arbitrary assignee fails", func(t *testing.T) {
		g := admissionGraphForTest(true)
		g.Nodes["p"].Config["assigneeStrategy"] = "initiator"
		require.Error(t, ValidateBusinessDefinition("officer_interview", g))
	})
	t.Run("visual labels remain editable", func(t *testing.T) {
		g := admissionGraphForTest(true)
		g.Nodes["p"].Label = "会长核准录用"
		require.NoError(t, ValidateBusinessDefinition("officer_interview", g))
	})
	t.Run("sequential first round still requires both signatures", func(t *testing.T) {
		g := admissionGraphForTest(true)
		delete(g.Nodes, "f")
		delete(g.Nodes, "j")
		g.Edges = nil
		for _, pair := range [][2]string{{"s", "m"}, {"m", "a"}, {"a", "b"}, {"b", "c"}, {"c", "p"}, {"p", "e"}} {
			g.Edges = append(g.Edges, &FlowEdge{Source: pair[0], Target: pair[1]})
		}
		require.NoError(t, ValidateBusinessDefinition("officer_interview", g))
	})
}
