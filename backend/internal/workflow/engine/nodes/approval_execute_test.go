package nodes

import (
	"context"
	"testing"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/engine"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/model"
	"github.com/Yogdunana/StarByte/backend/pkg/events"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- ApprovalNode Execute Tests ---
// (Validate is already covered; these test Execute and OnEnter flows)

func TestApprovalNode_Type(t *testing.T) {
	n := ApprovalNode{}
	assert.Equal(t, "approval", n.Type())
}

func TestApprovalNode_Execute_NextNodes(t *testing.T) {
	n := ApprovalNode{}
	graph := &engine.FlowGraph{
		Nodes: map[string]*engine.FlowNode{
			"approval": {ID: "approval", Type: "approval"},
			"end":      {ID: "end", Type: "end"},
		},
		Edges: []*engine.FlowEdge{
			{ID: "e1", Source: "approval", Target: "end"},
		},
	}

	next, err := n.Execute(context.Background(), &model.FlowInstance{}, graph.GetNode("approval"), graph, nil)
	require.NoError(t, err)
	assert.Equal(t, []string{"end"}, next)
}

func TestApprovalNode_Execute_NoOutgoing(t *testing.T) {
	n := ApprovalNode{}
	graph := &engine.FlowGraph{
		Nodes: map[string]*engine.FlowNode{
			"approval": {ID: "approval", Type: "approval"},
		},
		Edges: []*engine.FlowEdge{},
	}

	next, err := n.Execute(context.Background(), &model.FlowInstance{}, graph.GetNode("approval"), graph, nil)
	require.NoError(t, err)
	assert.Empty(t, next)
}

func TestApprovalNode_OnEnter_StaticAssignees(t *testing.T) {
	bus := events.NewEventBus()
	taskRepo := &mockTaskRepo{}

	n := &ApprovalNode{TaskRepo: taskRepo, EventBus: bus}

	assigneeID := uuid.New()
	inst := &model.FlowInstance{ID: uuid.New(), InitiatorID: uuid.New()}
	config := map[string]interface{}{
		"assigneeStrategy": "static",
		"assignees":        []interface{}{assigneeID.String()},
	}
	node := &engine.FlowNode{ID: "approval", Type: "approval", Config: config, Label: "Manager Approval"}

	err := n.OnEnter(context.Background(), inst, node, map[string]interface{}{})
	require.NoError(t, err)

	// Verify task was created.
	require.NotNil(t, taskRepo.createdTask)
	assert.Equal(t, inst.ID, taskRepo.createdTask.InstanceID)
	assert.Equal(t, "approval", taskRepo.createdTask.NodeID)
	assert.Equal(t, "Manager Approval", taskRepo.createdTask.NodeName)
	assert.Equal(t, "approval", taskRepo.createdTask.TaskType)
	require.NotNil(t, taskRepo.createdTask.AssigneeID)
	assert.Equal(t, assigneeID, *taskRepo.createdTask.AssigneeID)
	assert.Equal(t, 0, taskRepo.createdTask.Status) // pending
}

func TestApprovalNode_OnEnter_InitiatorStrategy(t *testing.T) {
	bus := events.NewEventBus()
	taskRepo := &mockTaskRepo{}

	n := &ApprovalNode{TaskRepo: taskRepo, EventBus: bus}

	initiatorID := uuid.New()
	inst := &model.FlowInstance{ID: uuid.New(), InitiatorID: initiatorID}
	config := map[string]interface{}{
		"assigneeStrategy": "initiator",
	}
	node := &engine.FlowNode{ID: "approval", Type: "approval", Config: config}

	err := n.OnEnter(context.Background(), inst, node, map[string]interface{}{})
	require.NoError(t, err)

	require.NotNil(t, taskRepo.createdTask)
	assert.Equal(t, initiatorID, *taskRepo.createdTask.AssigneeID)
}

func TestApprovalNode_OnEnter_WithDueDate(t *testing.T) {
	bus := events.NewEventBus()
	taskRepo := &mockTaskRepo{}

	n := &ApprovalNode{TaskRepo: taskRepo, EventBus: bus}

	inst := &model.FlowInstance{ID: uuid.New(), InitiatorID: uuid.New()}
	assigneeID := uuid.New()
	config := map[string]interface{}{
		"assigneeStrategy": "static",
		"assignees":        []interface{}{assigneeID.String()},
		"dueDays":          float64(3),
	}
	node := &engine.FlowNode{ID: "approval", Type: "approval", Config: config}

	err := n.OnEnter(context.Background(), inst, node, map[string]interface{}{})
	require.NoError(t, err)

	require.NotNil(t, taskRepo.createdTask)
	require.NotNil(t, taskRepo.createdTask.DueDate)
}

func TestApprovalNode_OnEnter_PublishesTaskCreatedEvent(t *testing.T) {
	bus := events.NewEventBus()
	taskRepo := &mockTaskRepo{}

	var receivedEvent events.TaskCreatedEvent
	bus.Subscribe("task.created", func(ctx context.Context, e events.Event) error {
		te, ok := e.(events.TaskCreatedEvent)
		if ok {
			receivedEvent = te
		}
		return nil
	})

	n := &ApprovalNode{TaskRepo: taskRepo, EventBus: bus}

	assigneeID := uuid.New()
	inst := &model.FlowInstance{ID: uuid.New(), InitiatorID: uuid.New()}
	config := map[string]interface{}{
		"assigneeStrategy": "static",
		"assignees":        []interface{}{assigneeID.String()},
	}
	node := &engine.FlowNode{ID: "approval", Type: "approval", Config: config, Label: "Manager Approval"}

	err := n.OnEnter(context.Background(), inst, node, map[string]interface{}{})
	require.NoError(t, err)

	assert.Equal(t, inst.ID, receivedEvent.InstanceID)
	assert.Equal(t, assigneeID, receivedEvent.AssigneeID)
	assert.Equal(t, "approval", receivedEvent.NodeID)
	assert.Equal(t, "Manager Approval", receivedEvent.NodeName)
	assert.Equal(t, "approval", receivedEvent.TaskType)
}

func TestApprovalNode_OnEnter_NoAssignees(t *testing.T) {
	bus := events.NewEventBus()
	taskRepo := &mockTaskRepo{}

	n := &ApprovalNode{TaskRepo: taskRepo, EventBus: bus}

	inst := &model.FlowInstance{ID: uuid.New(), InitiatorID: uuid.New()}
	config := map[string]interface{}{
		"assigneeStrategy": "static",
		"assignees":        []interface{}{}, // empty
	}
	node := &engine.FlowNode{ID: "approval", Type: "approval", Config: config}

	err := n.OnEnter(context.Background(), inst, node, map[string]interface{}{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "没有处理人")
}

func TestApprovalNode_OnLeave(t *testing.T) {
	n := ApprovalNode{}
	err := n.OnLeave(context.Background(), &model.FlowInstance{}, &engine.FlowNode{}, nil)
	require.NoError(t, err)
}

func TestApprovalNode_Validate_MissingConfig(t *testing.T) {
	n := ApprovalNode{}
	node := &engine.FlowNode{ID: "approval", Type: "approval", Config: nil}
	err := n.Validate(node)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "缺少配置")
}

func TestApprovalNode_Validate_StaticMissingAssignees(t *testing.T) {
	n := ApprovalNode{}
	config := map[string]interface{}{
		"assigneeStrategy": "static",
		"assignees":        []interface{}{}, // empty
	}
	node := &engine.FlowNode{ID: "approval", Type: "approval", Config: config}
	err := n.Validate(node)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "assignees")
}

func TestApprovalNode_Validate_InvalidStrategy(t *testing.T) {
	n := ApprovalNode{}
	config := map[string]interface{}{
		"assigneeStrategy": "nonexistent_strategy",
	}
	node := &engine.FlowNode{ID: "approval", Type: "approval", Config: config}
	err := n.Validate(node)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "不支持的处理人策略")
}
