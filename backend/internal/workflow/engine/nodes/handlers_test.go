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
	"gorm.io/gorm"
)

// --- StartNode Tests ---

func TestStartNode_Type(t *testing.T) {
	n := StartNode{}
	assert.Equal(t, "start", n.Type())
}

func TestStartNode_Execute(t *testing.T) {
	n := StartNode{}
	graph := &engine.FlowGraph{
		Nodes: map[string]*engine.FlowNode{
			"n1": {ID: "n1", Type: "start"},
			"n2": {ID: "n2", Type: "end"},
		},
		Edges: []*engine.FlowEdge{
			{ID: "e1", Source: "n1", Target: "n2"},
		},
	}

	next, err := n.Execute(context.Background(), &model.FlowInstance{}, graph.GetNode("n1"), graph, nil)
	require.NoError(t, err)
	assert.Equal(t, []string{"n2"}, next)
}

func TestStartNode_Execute_NoOutgoing(t *testing.T) {
	n := StartNode{}
	graph := &engine.FlowGraph{
		Nodes: map[string]*engine.FlowNode{
			"n1": {ID: "n1", Type: "start"},
		},
		Edges: []*engine.FlowEdge{},
	}

	next, err := n.Execute(context.Background(), &model.FlowInstance{}, graph.GetNode("n1"), graph, nil)
	require.NoError(t, err)
	assert.Empty(t, next)
}

// --- EndNode Tests ---

func TestEndNode_Type(t *testing.T) {
	n := EndNode{}
	assert.Equal(t, "end", n.Type())
}

func TestEndNode_Execute(t *testing.T) {
	n := EndNode{}
	graph := &engine.FlowGraph{
		Nodes: map[string]*engine.FlowNode{
			"n1": {ID: "n1", Type: "end"},
		},
	}

	next, err := n.Execute(context.Background(), &model.FlowInstance{}, graph.GetNode("n1"), graph, nil)
	require.NoError(t, err)
	assert.Empty(t, next)
}

// --- ExclusiveGatewayNode Tests ---

func TestExclusiveGatewayNode_Type(t *testing.T) {
	n := ExclusiveGatewayNode{ExprEngine: engine.NewExpressionEngine()}
	assert.Equal(t, "exclusive_gateway", n.Type())
}

func TestExclusiveGatewayNode_Execute_MatchBranch(t *testing.T) {
	n := ExclusiveGatewayNode{ExprEngine: engine.NewExpressionEngine()}

	config := map[string]interface{}{
		"branches": []interface{}{
			map[string]interface{}{
				"id":         "high",
				"expression": "amount > 10000",
			},
			map[string]interface{}{
				"id":         "low",
				"expression": "amount <= 10000",
			},
		},
	}

	graph := &engine.FlowGraph{
		Nodes: map[string]*engine.FlowNode{
			"gw":  {ID: "gw", Type: "exclusive_gateway", Config: config},
			"end": {ID: "end", Type: "end"},
		},
		Edges: []*engine.FlowEdge{
			{ID: "e1", Source: "gw", Target: "end", SourceHandle: "high"},
			{ID: "e2", Source: "gw", Target: "end", SourceHandle: "low"},
		},
	}

	vars := map[string]interface{}{"amount": 5000}
	next, err := n.Execute(context.Background(), &model.FlowInstance{}, graph.GetNode("gw"), graph, vars)
	require.NoError(t, err)
	assert.Equal(t, []string{"end"}, next)
}

func TestExclusiveGatewayNode_Execute_NoOutgoingEdge(t *testing.T) {
	n := ExclusiveGatewayNode{ExprEngine: engine.NewExpressionEngine()}

	config := map[string]interface{}{
		"branches": []interface{}{
			map[string]interface{}{
				"id":         "b1",
				"expression": "amount > 1000",
			},
		},
	}

	graph := &engine.FlowGraph{
		Nodes: map[string]*engine.FlowNode{
			"gw": {ID: "gw", Type: "exclusive_gateway", Config: config},
		},
		Edges: []*engine.FlowEdge{},
	}

	vars := map[string]interface{}{"amount": 5000}
	_, err := n.Execute(context.Background(), &model.FlowInstance{}, graph.GetNode("gw"), graph, vars)
	require.Error(t, err)
}

func TestExclusiveGatewayNode_Validate(t *testing.T) {
	n := ExclusiveGatewayNode{ExprEngine: engine.NewExpressionEngine()}

	config := map[string]interface{}{
		"branches": []interface{}{
			map[string]interface{}{
				"id":         "b1",
				"expression": "amount > 1000",
			},
			map[string]interface{}{
				"id":         "b2",
				"expression": "amount <= 1000",
			},
		},
	}

	node := &engine.FlowNode{ID: "gw", Type: "exclusive_gateway", Config: config}
	err := n.Validate(node)
	require.NoError(t, err)
}

func TestExclusiveGatewayNode_Validate_MissingConfig(t *testing.T) {
	n := ExclusiveGatewayNode{ExprEngine: engine.NewExpressionEngine()}
	node := &engine.FlowNode{ID: "gw", Type: "exclusive_gateway", Config: nil}

	err := n.Validate(node)
	require.Error(t, err)
}

func TestExclusiveGatewayNode_Validate_TooFewBranches(t *testing.T) {
	n := ExclusiveGatewayNode{ExprEngine: engine.NewExpressionEngine()}

	config := map[string]interface{}{
		"branches": []interface{}{
			map[string]interface{}{
				"id":         "b1",
				"expression": "amount > 1000",
			},
		},
	}

	node := &engine.FlowNode{ID: "gw", Type: "exclusive_gateway", Config: config}
	err := n.Validate(node)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at least 2 branches")
}

// --- ParallelGatewayNode Tests ---

func TestParallelGatewayNode_Type(t *testing.T) {
	n := ParallelGatewayNode{}
	assert.Equal(t, "parallel_gateway", n.Type())
}

func TestParallelGatewayNode_Execute(t *testing.T) {
	n := ParallelGatewayNode{}
	graph := &engine.FlowGraph{
		Nodes: map[string]*engine.FlowNode{
			"gw": {ID: "gw", Type: "parallel_gateway"},
			"a":  {ID: "a", Type: "end"},
			"b":  {ID: "b", Type: "end"},
		},
		Edges: []*engine.FlowEdge{
			{ID: "e1", Source: "gw", Target: "a"},
			{ID: "e2", Source: "gw", Target: "b"},
		},
	}

	next, err := n.Execute(context.Background(), &model.FlowInstance{}, graph.GetNode("gw"), graph, nil)
	require.NoError(t, err)
	assert.Len(t, next, 2)
	assert.Contains(t, next, "a")
	assert.Contains(t, next, "b")
}

func TestParallelGatewayNode_Execute_NoOutgoing(t *testing.T) {
	n := ParallelGatewayNode{}
	graph := &engine.FlowGraph{
		Nodes: map[string]*engine.FlowNode{
			"gw": {ID: "gw", Type: "parallel_gateway"},
		},
		Edges: []*engine.FlowEdge{},
	}

	_, err := n.Execute(context.Background(), &model.FlowInstance{}, graph.GetNode("gw"), graph, nil)
	require.Error(t, err)
}

// --- NodeRegistry Tests ---

func TestNodeRegistry_RegisterAndGet(t *testing.T) {
	r := NewNodeRegistry()
	r.Register(StartNode{})

	handler, err := r.Get("start")
	require.NoError(t, err)
	assert.Equal(t, "start", handler.Type())
}

func TestNodeRegistry_Get_NotFound(t *testing.T) {
	r := NewNodeRegistry()

	_, err := r.Get("nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not supported")
}

func TestNodeRegistry_Register_DuplicatePanics(t *testing.T) {
	r := NewNodeRegistry()
	r.Register(StartNode{})

	assert.Panics(t, func() {
		r.Register(StartNode{})
	})
}

func TestNewDefaultRegistry_AllRegistered(t *testing.T) {
	taskRepo := &mockTaskRepo{}
	registry := NewDefaultRegistry(engine.NewExpressionEngine(), events.NewEventBus(), taskRepo)

	types := []string{"start", "end", "approval", "exclusive_gateway", "parallel_gateway", "service_task", "notification_task"}
	for _, typ := range types {
		handler, err := registry.Get(typ)
		require.NoError(t, err, "handler for %s not found", typ)
		assert.Equal(t, typ, handler.Type())
	}
}

// --- ServiceTaskNode Tests ---

func TestServiceTaskNode_Execute_WithCallback(t *testing.T) {
	called := false
	cb := func(ctx context.Context, inst *model.FlowInstance, node *engine.FlowNode, vars map[string]interface{}) (map[string]interface{}, error) {
		called = true
		return map[string]interface{}{"result": "done"}, nil
	}

	n := &ServiceTaskNode{
		Callbacks: map[string]ServiceCallback{"myService": cb},
	}

	config := map[string]interface{}{"service": "myService"}
	graph := &engine.FlowGraph{
		Nodes: map[string]*engine.FlowNode{
			"task": {ID: "task", Type: "service_task", Config: config},
			"end":  {ID: "end", Type: "end"},
		},
		Edges: []*engine.FlowEdge{
			{ID: "e1", Source: "task", Target: "end"},
		},
	}

	vars := map[string]interface{}{}
	next, err := n.Execute(context.Background(), &model.FlowInstance{}, graph.GetNode("task"), graph, vars)
	require.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, "done", vars["result"])
	assert.Equal(t, []string{"end"}, next)
}

func TestServiceTaskNode_Execute_NoServiceName(t *testing.T) {
	n := &ServiceTaskNode{
		Callbacks: map[string]ServiceCallback{},
	}

	config := map[string]interface{}{}
	graph := &engine.FlowGraph{
		Nodes: map[string]*engine.FlowNode{
			"task": {ID: "task", Type: "service_task", Config: config},
		},
	}

	_, err := n.Execute(context.Background(), &model.FlowInstance{}, graph.GetNode("task"), graph, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing 'service'")
}

func TestServiceTaskNode_Execute_NoCallback(t *testing.T) {
	n := &ServiceTaskNode{
		Callbacks: map[string]ServiceCallback{},
	}

	config := map[string]interface{}{"service": "unknown"}
	graph := &engine.FlowGraph{
		Nodes: map[string]*engine.FlowNode{
			"task": {ID: "task", Type: "service_task", Config: config},
		},
	}

	_, err := n.Execute(context.Background(), &model.FlowInstance{}, graph.GetNode("task"), graph, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no callback")
}

// --- NotificationTaskNode Tests ---

func TestNotificationTaskNode_Execute(t *testing.T) {
	n := &NotificationTaskNode{EventBus: events.NewEventBus()}

	config := map[string]interface{}{"notificationType": "email"}
	graph := &engine.FlowGraph{
		Nodes: map[string]*engine.FlowNode{
			"notif": {ID: "notif", Type: "notification_task", Config: config},
			"end":   {ID: "end", Type: "end"},
		},
		Edges: []*engine.FlowEdge{
			{ID: "e1", Source: "notif", Target: "end"},
		},
	}

	vars := map[string]interface{}{}
	next, err := n.Execute(context.Background(), &model.FlowInstance{}, graph.GetNode("notif"), graph, vars)
	require.NoError(t, err)
	assert.Equal(t, []string{"end"}, next)
	assert.Equal(t, "email", vars["_notification_type"])
}

func TestNotificationTaskNode_Execute_DefaultType(t *testing.T) {
	n := &NotificationTaskNode{EventBus: events.NewEventBus()}

	config := map[string]interface{}{}
	graph := &engine.FlowGraph{
		Nodes: map[string]*engine.FlowNode{
			"notif": {ID: "notif", Type: "notification_task", Config: config},
		},
		Edges: []*engine.FlowEdge{},
	}

	vars := map[string]interface{}{}
	_, err := n.Execute(context.Background(), &model.FlowInstance{}, graph.GetNode("notif"), graph, vars)
	require.NoError(t, err)
	assert.Equal(t, "default", vars["_notification_type"])
}

// --- Mocks ---

// mockTaskRepo implements repo.TaskRepo for testing.
type mockTaskRepo struct {
	createdTask *model.FlowTask
}

func (m *mockTaskRepo) CreateTask(ctx context.Context, tx *gorm.DB, task *model.FlowTask) error {
	m.createdTask = task
	return nil
}

func (m *mockTaskRepo) GetTaskByID(ctx context.Context, id uuid.UUID) (*model.FlowTask, error) {
	return nil, nil
}

func (m *mockTaskRepo) UpdateTask(ctx context.Context, tx *gorm.DB, task *model.FlowTask) error {
	return nil
}

func (m *mockTaskRepo) ListTodoTasks(ctx context.Context, assigneeID uuid.UUID, page, pageSize int) ([]model.FlowTask, int64, error) {
	return nil, 0, nil
}

func (m *mockTaskRepo) ListDoneTasks(ctx context.Context, assigneeID uuid.UUID, page, pageSize int) ([]model.FlowTask, int64, error) {
	return nil, 0, nil
}

func (m *mockTaskRepo) ListTasksByInstance(ctx context.Context, instanceID uuid.UUID) ([]model.FlowTask, error) {
	return nil, nil
}

func (m *mockTaskRepo) CreateHistory(ctx context.Context, tx *gorm.DB, hist *model.FlowHistory) error {
	return nil
}

func (m *mockTaskRepo) ListHistory(ctx context.Context, instanceID uuid.UUID) ([]model.FlowHistory, error) {
	return nil, nil
}
