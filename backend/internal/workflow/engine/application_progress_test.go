package engine

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestApplicationProgressOffPathBranchIsSkipped(t *testing.T) {
	graph, err := ParseGraph(branchMembershipBPMN())
	require.NoError(t, err)
	require.Equal(t, []string{"start", "fork", "path_a", "path_b", "end"}, displayNodeOrder(graph))

	current := map[string]bool{"path_a": true}
	future := reachableNodeIDs(graph, []string{"path_a"})
	states := map[string]string{}
	for _, id := range displayNodeOrder(graph) {
		states[id] = applicationStepState(0, graph.GetNode(id), current[id], "", false, future[id])
	}
	require.Equal(t, "done", states["start"])
	require.Equal(t, "done", states["fork"])
	require.Equal(t, "current", states["path_a"])
	require.Equal(t, "skipped", states["path_b"])
	require.Equal(t, "pending", states["end"])
}

func TestLastApplicationApproval(t *testing.T) {
	graph, err := ParseGraph(MemberApplicationBPMN())
	require.NoError(t, err)
	require.False(t, lastApplicationApproval(graph, []string{"officer"}, "officer"))
	require.False(t, lastApplicationApproval(graph, []string{"minister"}, "minister"))
	require.True(t, lastApplicationApproval(graph, []string{"president"}, "president"))
	require.False(t, lastApplicationApproval(graph, nil, "missing"))

	branch, err := ParseGraph(branchMembershipBPMN())
	require.NoError(t, err)
	require.True(t, lastApplicationApproval(branch, []string{"path_a"}, "path_a"))
	require.True(t, lastApplicationApproval(branch, []string{"path_b"}, "path_b"))
	require.False(t, lastApplicationApproval(branch, []string{"fork"}, "fork"))

	parallel, err := ParseGraph(parallelMembershipBPMN())
	require.NoError(t, err)
	require.False(t, lastApplicationApproval(parallel, []string{"path_a", "path_b"}, "path_a"))
	require.False(t, lastApplicationApproval(parallel, []string{"path_a", "path_b"}, "path_b"))
	require.True(t, lastApplicationApproval(parallel, []string{"path_b"}, "path_b"))
}

func branchMembershipBPMN() []byte {
	raw, err := json.Marshal(map[string]interface{}{
		"nodes": []map[string]interface{}{
			node("start", "start", "提交申请", 20, nil),
			node("fork", "condition", "分流", 80, nil),
			node("path_a", "approval", "路径A", 160, map[string]interface{}{
				"assigneeStrategy": "role", "roleCode": "officer", "approvalType": "any",
			}),
			node("path_b", "approval", "路径B", 240, map[string]interface{}{
				"assigneeStrategy": "role", "roleCode": "minister", "approvalType": "any",
			}),
			node("end", "end", "结束", 320, nil),
		},
		"edges": []map[string]string{
			{"id": "start-fork", "source": "start", "target": "fork"},
			{"id": "fork-a", "source": "fork", "target": "path_a"},
			{"id": "fork-b", "source": "fork", "target": "path_b"},
			{"id": "a-end", "source": "path_a", "target": "end"},
			{"id": "b-end", "source": "path_b", "target": "end"},
		},
	})
	if err != nil {
		panic(err)
	}
	return raw
}

func parallelMembershipBPMN() []byte {
	raw, err := json.Marshal(map[string]interface{}{
		"nodes": []map[string]interface{}{
			node("start", "start", "提交申请", 20, nil),
			node("fork", "parallel_gateway", "并行", 80, nil),
			node("path_a", "approval", "路径A", 160, map[string]interface{}{
				"assigneeStrategy": "role", "roleCode": "officer", "approvalType": "any",
			}),
			node("path_b", "approval", "路径B", 240, map[string]interface{}{
				"assigneeStrategy": "role", "roleCode": "minister", "approvalType": "any",
			}),
			node("join", "merge", "汇合", 300, nil),
			node("end", "end", "结束", 360, nil),
		},
		"edges": []map[string]string{
			{"id": "start-fork", "source": "start", "target": "fork"},
			{"id": "fork-a", "source": "fork", "target": "path_a"},
			{"id": "fork-b", "source": "fork", "target": "path_b"},
			{"id": "a-join", "source": "path_a", "target": "join"},
			{"id": "b-join", "source": "path_b", "target": "join"},
			{"id": "join-end", "source": "join", "target": "end"},
		},
	})
	if err != nil {
		panic(err)
	}
	return raw
}
