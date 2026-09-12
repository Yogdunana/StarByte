package engine

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/model"
	"github.com/Yogdunana/StarByte/backend/pkg/events"
)

func taskLifecycleEngine(t *testing.T, tasks *mockTaskRepo) (*FlowEngine, *storingInstRepo) {
	t.Helper()
	defID, verID := uuid.New(), uuid.New()
	def := &model.FlowDefinition{ID: defID, Key: TaskDefinitionKey, Status: 1}
	ver := &model.FlowDefinitionVersion{ID: verID, DefinitionID: defID, BpmnData: TaskLifecycleBPMN(), Status: 1}
	insts := &storingInstRepo{insts: map[uuid.UUID]*model.FlowInstance{}}
	e := NewFlowEngine(&mockDefRepo{def: def, version: ver}, insts, tasks, newMockVarRepo(), nil, &mockRegistryForTest{handlers: map[string]NodeHandler{
		"start": stubStartHandler{}, "end": stubEndHandler{}, "approval": &waitingTestNode{},
	}}, NewExpressionEngine(), events.NewEventBus(), nil)
	e.businessTransaction = true
	return e, insts
}

func TestTaskLifecycleStartApproveSpineComplete(t *testing.T) {
	creator, assignee, reviewer, acceptor := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	tasks := newMockTaskRepo()
	e, insts := taskLifecycleEngine(t, tasks)

	inst, err := e.Start(context.Background(), TaskDefinitionKey, uuid.New().String(), TaskBusinessType, creator, map[string]interface{}{
		"task_id": uuid.New().String(),
	})
	require.NoError(t, err)
	require.Equal(t, 0, inst.Status)
	nodeID, done, err := e.RunningApprovalNode(context.Background(), inst.ID)
	require.NoError(t, err)
	require.False(t, done)
	require.Equal(t, "assignment", nodeID)

	addApprovalTask(tasks, inst.ID, "assignment", creator)
	require.NoError(t, e.CompleteTaskApproval(context.Background(), inst.ID, creator, ActionApprove, "指定执行人"))
	nodeID, done, err = e.RunningApprovalNode(context.Background(), inst.ID)
	require.NoError(t, err)
	require.False(t, done)
	require.Equal(t, "execution", nodeID)

	addApprovalTask(tasks, inst.ID, "execution", assignee)
	require.NoError(t, e.CompleteTaskApproval(context.Background(), inst.ID, assignee, ActionApprove, "提交交付"))
	nodeID, _, err = e.RunningApprovalNode(context.Background(), inst.ID)
	require.NoError(t, err)
	require.Equal(t, "review", nodeID)

	addApprovalTask(tasks, inst.ID, "review", reviewer)
	require.NoError(t, e.CompleteTaskApproval(context.Background(), inst.ID, reviewer, ActionApprove, "审核通过"))
	nodeID, _, err = e.RunningApprovalNode(context.Background(), inst.ID)
	require.NoError(t, err)
	require.Equal(t, "acceptance", nodeID)

	addApprovalTask(tasks, inst.ID, "acceptance", acceptor)
	require.NoError(t, e.CompleteTaskApproval(context.Background(), inst.ID, acceptor, ActionApprove, "验收通过"))
	_, done, err = e.RunningApprovalNode(context.Background(), inst.ID)
	require.NoError(t, err)
	require.True(t, done)
	require.Equal(t, 1, insts.insts[inst.ID].Status)
}

func TestTaskLifecycleRejectTerminatesPendingTodos(t *testing.T) {
	creator, reviewer, other := uuid.New(), uuid.New(), uuid.New()
	tasks := newMockTaskRepo()
	e, insts := taskLifecycleEngine(t, tasks)

	inst, err := e.Start(context.Background(), TaskDefinitionKey, uuid.New().String(), TaskBusinessType, creator, nil)
	require.NoError(t, err)
	addApprovalTask(tasks, inst.ID, "assignment", creator)
	require.NoError(t, e.CompleteTaskApproval(context.Background(), inst.ID, creator, ActionApprove, "分配"))
	addApprovalTask(tasks, inst.ID, "execution", creator)
	require.NoError(t, e.CompleteTaskApproval(context.Background(), inst.ID, creator, ActionApprove, "交付"))
	addApprovalTask(tasks, inst.ID, "review", reviewer)
	addApprovalTask(tasks, inst.ID, "review", other)
	require.NoError(t, e.CompleteTaskApproval(context.Background(), inst.ID, reviewer, ActionReject, "质量不符"))
	require.Equal(t, 2, insts.insts[inst.ID].Status)

	var canceled, rejected int
	for _, task := range tasks.tasks {
		if task.InstanceID != inst.ID {
			continue
		}
		if task.AssigneeID != nil && *task.AssigneeID == reviewer && task.NodeID == "review" {
			require.Equal(t, 2, task.Status)
			rejected++
		}
		if task.AssigneeID != nil && *task.AssigneeID == other && task.NodeID == "review" {
			require.Equal(t, 5, task.Status)
			require.Equal(t, "cancel", task.Action)
			canceled++
		}
	}
	require.Equal(t, 1, rejected)
	require.Equal(t, 1, canceled)
	require.Error(t, e.CompleteTaskApproval(context.Background(), inst.ID, other, ActionApprove, "不能救活"))
	terminated, err := e.InstanceTerminated(context.Background(), inst.ID)
	require.NoError(t, err)
	require.True(t, terminated)
}

func TestClaimTaskAssignmentAdvancesToExecution(t *testing.T) {
	creator, claimer := uuid.New(), uuid.New()
	tasks := newMockTaskRepo()
	e, _ := taskLifecycleEngine(t, tasks)

	inst, err := e.Start(context.Background(), TaskDefinitionKey, uuid.New().String(), TaskBusinessType, creator, nil)
	require.NoError(t, err)
	addApprovalTask(tasks, inst.ID, "assignment", creator)
	require.NoError(t, e.ClaimTaskAssignment(context.Background(), inst.ID, claimer, "我来认领"))
	nodeID, done, err := e.RunningApprovalNode(context.Background(), inst.ID)
	require.NoError(t, err)
	require.False(t, done)
	require.Equal(t, "execution", nodeID)
	for _, task := range tasks.tasks {
		if task.InstanceID == inst.ID && task.NodeID == "assignment" && task.AssigneeID != nil && *task.AssigneeID == creator {
			require.Equal(t, 5, task.Status)
			require.Equal(t, "cancel", task.Action)
		}
	}
}

func TestReassignStageMovesPendingExecutionTodo(t *testing.T) {
	creator, previous, next, operator := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	tasks := newMockTaskRepo()
	e, _ := taskLifecycleEngine(t, tasks)
	inst, err := e.Start(context.Background(), TaskDefinitionKey, uuid.New().String(), TaskBusinessType, creator, nil)
	require.NoError(t, err)
	addApprovalTask(tasks, inst.ID, "assignment", creator)
	require.NoError(t, e.CompleteTaskApproval(context.Background(), inst.ID, creator, ActionApprove, "指定执行人"))
	addApprovalTask(tasks, inst.ID, "execution", previous)
	require.NoError(t, e.ReassignStage(context.Background(), inst.ID, "execution", previous, next, operator, "超时升级重新分配"))
	moved := false
	for _, task := range tasks.tasks {
		if task.InstanceID == inst.ID && task.NodeID == "execution" && task.Status == 0 && task.AssigneeID != nil && *task.AssigneeID == next {
			moved = true
			require.NotNil(t, task.DueDate)
		}
	}
	require.True(t, moved)
}

func TestTaskLifecycleBPMNMatchesDefinitionKey(t *testing.T) {
	require.Equal(t, "task_lifecycle", TaskDefinitionKey)
	graph, err := ParseGraph(TaskLifecycleBPMN())
	require.NoError(t, err)
	require.NoError(t, ValidateBusinessDefinition(TaskDefinitionKey, graph))
	require.Equal(t, "start", graph.FindStartNode().ID)
}
