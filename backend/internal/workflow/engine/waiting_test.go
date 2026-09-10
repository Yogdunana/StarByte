package engine

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/model"
	"github.com/Yogdunana/StarByte/backend/pkg/events"
)

type waitingTestNode struct {
	stubStartHandler
	entered, executed, left int
}

func (n *waitingTestNode) Type() string            { return "approval" }
func (n *waitingTestNode) WaitForCompletion() bool { return true }
func (n *waitingTestNode) OnEnter(context.Context, *model.FlowInstance, *FlowNode, map[string]interface{}) error {
	n.entered++
	return nil
}
func (n *waitingTestNode) OnLeave(context.Context, *model.FlowInstance, *FlowNode, map[string]interface{}) error {
	n.left++
	return nil
}
func (n *waitingTestNode) Execute(ctx context.Context, inst *model.FlowInstance, node *FlowNode, graph *FlowGraph, vars map[string]interface{}) ([]string, error) {
	n.executed++
	return n.stubStartHandler.Execute(ctx, inst, node, graph, vars)
}
func waitingEngine(n *waitingTestNode) *FlowEngine {
	return NewFlowEngine(&mockDefRepo{}, &mockInstRepo{}, newMockTaskRepo(), newMockVarRepo(), nil, &mockRegistryForTest{handlers: map[string]NodeHandler{"start": stubStartHandler{}, "approval": n, "end": stubEndHandler{}}}, NewExpressionEngine(), events.NewEventBus(), nil)
}
func TestApprovalEntryDoesNotExecuteOutgoingEdges(t *testing.T) {
	node := &waitingTestNode{}
	e := waitingEngine(node)
	inst := &model.FlowInstance{ID: uuid.New()}
	require.NoError(t, e.executeFromNodes(context.Background(), inst, buildGraphWithApproval(t), []string{"start"}, nil))
	require.Equal(t, 0, inst.Status)
	require.Equal(t, []string{"approve"}, GetCurrentNodeIDs(inst.CurrentNodeIDs))
	require.Equal(t, 1, node.entered)
	require.Zero(t, node.executed)
	require.Zero(t, node.left)
}
func TestAnEndedBranchDoesNotCompleteAWaitingParallelBranch(t *testing.T) {
	node := &waitingTestNode{}
	e := waitingEngine(node)
	inst := &model.FlowInstance{ID: uuid.New()}
	graph := buildGraphWithApproval(t)
	require.NoError(t, e.executeFromNodes(context.Background(), inst, graph, []string{"end", "approve"}, nil))
	require.Zero(t, inst.Status)
	require.Equal(t, []string{"approve"}, GetCurrentNodeIDs(inst.CurrentNodeIDs))
	// A later unrelated branch reaching the end must retain the existing wait.
	require.NoError(t, e.executeFromNodes(context.Background(), inst, graph, []string{"end"}, nil))
	require.Zero(t, inst.Status)
	require.Equal(t, []string{"approve"}, GetCurrentNodeIDs(inst.CurrentNodeIDs))
}

func TestTerminationClosesPendingBranchesAndPreservesCompletedDecisions(t *testing.T) {
	e := waitingEngine(&waitingTestNode{})
	id, user := uuid.New(), uuid.New()
	inst := &model.FlowInstance{ID: id, Status: 3, CurrentNodeIDs: []byte(`["a","b"]`)}
	completed := &model.FlowTask{ID: uuid.New(), InstanceID: id, Status: 1, Action: "approve"}
	pending := &model.FlowTask{ID: uuid.New(), InstanceID: id, Status: 0}
	unrelated := &model.FlowTask{ID: uuid.New(), InstanceID: uuid.New(), Status: 0}
	tasks := e.taskRepo.(*mockTaskRepo)
	tasks.tasks[completed.ID] = completed
	tasks.tasks[pending.ID] = pending
	tasks.tasks[unrelated.ID] = unrelated
	require.NoError(t, e.terminateInstance(context.Background(), inst, user, "管理员终止"))
	require.Equal(t, 2, inst.Status)
	require.Empty(t, GetCurrentNodeIDs(inst.CurrentNodeIDs))
	require.Equal(t, 5, tasks.tasks[pending.ID].Status)
	require.Equal(t, "cancel", tasks.tasks[pending.ID].Action)
	require.NotNil(t, tasks.tasks[pending.ID].CompletedAt)
	require.Equal(t, "approve", tasks.tasks[completed.ID].Action)
	require.Zero(t, tasks.tasks[unrelated.ID].Status)
}
