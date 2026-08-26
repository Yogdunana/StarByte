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

// TestEndToEnd_ApprovalFlow simulates a complete approval workflow:
//
//   start → approval → end
//
// It verifies:
// 1. StartNode moves to the approval node
// 2. ApprovalNode.OnEnter creates a task and publishes an event
// 3. ApprovalNode.Execute (after task completion) moves to end
// 4. EndNode.Execute terminates the flow

func TestEndToEnd_ApprovalFlow(t *testing.T) {
	// --- Setup ---
	bus := events.NewEventBus()
	taskRepo := &mockTaskRepo{}
	registry := NewNodeRegistry()

	// Register all nodes.
	startNode := &StartNode{}
	endNode := &EndNode{}
	approvalNode := &ApprovalNode{TaskRepo: taskRepo, EventBus: bus}
	registry.Register(startNode)
	registry.Register(endNode)
	registry.Register(approvalNode)

	// Build the graph: start → approval → end
	assigneeID := uuid.New()
	graph := &engine.FlowGraph{
		Nodes: map[string]*engine.FlowNode{
			"start":    {ID: "start", Type: "start", Label: "Start"},
			"approval": {ID: "approval", Type: "approval", Label: "Manager Approval", Config: map[string]interface{}{
				"assigneeStrategy": "static",
				"assignees":        []interface{}{assigneeID.String()},
			}},
			"end": {ID: "end", Type: "end", Label: "End"},
		},
		Edges: []*engine.FlowEdge{
			{ID: "e1", Source: "start", Target: "approval"},
			{ID: "e2", Source: "approval", Target: "end"},
		},
	}

	// Track events.
	var taskCreatedEvents []events.TaskCreatedEvent
	bus.Subscribe("task.created", func(ctx context.Context, e events.Event) error {
		if te, ok := e.(events.TaskCreatedEvent); ok {
			taskCreatedEvents = append(taskCreatedEvents, te)
		}
		return nil
	})

	inst := &model.FlowInstance{
		ID:         uuid.New(),
		InitiatorID: uuid.New(),
		Status:     0, // running
	}
	vars := map[string]interface{}{}

	// --- Step 1: StartNode Execute → should go to approval ---
	startHandler, err := registry.Get("start")
	require.NoError(t, err)

	next, err := startHandler.Execute(context.Background(), inst, graph.GetNode("start"), graph, vars)
	require.NoError(t, err)
	assert.Equal(t, []string{"approval"}, next)

	// --- Step 2: ApprovalNode OnEnter → creates task ---
	approvalHandler, err := registry.Get("approval")
	require.NoError(t, err)

	err = approvalHandler.OnEnter(context.Background(), inst, graph.GetNode("approval"), vars)
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

	// Verify event was published.
	require.Len(t, taskCreatedEvents, 1)
	assert.Equal(t, assigneeID, taskCreatedEvents[0].AssigneeID)
	assert.Equal(t, "Manager Approval", taskCreatedEvents[0].NodeName)

	// --- Step 3: ApprovalNode Execute (task completed) → should go to end ---
	next, err = approvalHandler.Execute(context.Background(), inst, graph.GetNode("approval"), graph, vars)
	require.NoError(t, err)
	assert.Equal(t, []string{"end"}, next)

	// --- Step 4: EndNode Execute → flow ends ---
	endHandler, err := registry.Get("end")
	require.NoError(t, err)

	next, err = endHandler.Execute(context.Background(), inst, graph.GetNode("end"), graph, vars)
	require.NoError(t, err)
	assert.Empty(t, next) // end node has no outgoing
}

// TestEndToEnd_ExclusiveGatewayFlow simulates a branching workflow:
//
//   start → exclusive_gateway → (amount > 10000 → high_approval → end)
//                              (amount <= 10000 → low_approval → end)

func TestEndToEnd_ExclusiveGatewayFlow(t *testing.T) {
	bus := events.NewEventBus()
	taskRepo := &mockTaskRepo{}
	registry := NewNodeRegistry()

	exprEngine := engine.NewExpressionEngine()
	registry.Register(&StartNode{})
	registry.Register(&EndNode{})
	registry.Register(&ApprovalNode{TaskRepo: taskRepo, EventBus: bus})
	registry.Register(&ExclusiveGatewayNode{ExprEngine: exprEngine})

	// Build the graph.
	graph := &engine.FlowGraph{
		Nodes: map[string]*engine.FlowNode{
			"start":    {ID: "start", Type: "start"},
			"gw":       {ID: "gw", Type: "exclusive_gateway", Config: map[string]interface{}{
				"branches": []interface{}{
					map[string]interface{}{"id": "high", "expression": "amount > 10000"},
					map[string]interface{}{"id": "low", "expression": "amount <= 10000"},
				},
			}},
			"high_approval": {ID: "high_approval", Type: "approval", Config: map[string]interface{}{
				"assigneeStrategy": "initiator",
			}, Label: "High Value Approval"},
			"low_approval": {ID: "low_approval", Type: "approval", Config: map[string]interface{}{
				"assigneeStrategy": "initiator",
			}, Label: "Low Value Approval"},
			"end": {ID: "end", Type: "end"},
		},
		Edges: []*engine.FlowEdge{
			{ID: "e1", Source: "start", Target: "gw"},
			{ID: "e2", Source: "gw", Target: "high_approval", SourceHandle: "high"},
			{ID: "e3", Source: "gw", Target: "low_approval", SourceHandle: "low"},
			{ID: "e4", Source: "high_approval", Target: "end"},
			{ID: "e5", Source: "low_approval", Target: "end"},
		},
	}

	inst := &model.FlowInstance{
		ID:          uuid.New(),
		InitiatorID: uuid.New(),
		Status:      0,
	}

	// --- High amount path ---
	vars := map[string]interface{}{"amount": 15000}

	startHandler, err := registry.Get("start")
	require.NoError(t, err)
	next, err := startHandler.Execute(context.Background(), inst, graph.GetNode("start"), graph, vars)
	require.NoError(t, err)
	assert.Equal(t, []string{"gw"}, next)

	// Gateway should route to high_approval.
	gwHandler, err := registry.Get("exclusive_gateway")
	require.NoError(t, err)
	next, err = gwHandler.Execute(context.Background(), inst, graph.GetNode("gw"), graph, vars)
	require.NoError(t, err)
	assert.Equal(t, []string{"high_approval"}, next)

	// High approval OnEnter should create a task.
	approvalHandler, err := registry.Get("approval")
	require.NoError(t, err)
	err = approvalHandler.OnEnter(context.Background(), inst, graph.GetNode("high_approval"), vars)
	require.NoError(t, err)
	require.NotNil(t, taskRepo.createdTask)
	assert.Equal(t, inst.InitiatorID, *taskRepo.createdTask.AssigneeID)

	// High approval Execute should go to end.
	next, err = approvalHandler.Execute(context.Background(), inst, graph.GetNode("high_approval"), graph, vars)
	require.NoError(t, err)
	assert.Equal(t, []string{"end"}, next)

	// End node.
	endHandler, err := registry.Get("end")
	require.NoError(t, err)
	next, err = endHandler.Execute(context.Background(), inst, graph.GetNode("end"), graph, vars)
	require.NoError(t, err)
	assert.Empty(t, next)
}

// TestEndToEnd_LowAmountPath tests the low-amount branch of the exclusive gateway.

func TestEndToEnd_LowAmountPath(t *testing.T) {
	registry := NewNodeRegistry()

	exprEngine := engine.NewExpressionEngine()
	registry.Register(&StartNode{})
	registry.Register(&EndNode{})
	registry.Register(&ExclusiveGatewayNode{ExprEngine: exprEngine})

	graph := &engine.FlowGraph{
		Nodes: map[string]*engine.FlowNode{
			"gw":  {ID: "gw", Type: "exclusive_gateway", Config: map[string]interface{}{
				"branches": []interface{}{
					map[string]interface{}{"id": "high", "expression": "amount > 10000"},
					map[string]interface{}{"id": "low", "expression": "amount <= 10000"},
				},
			}},
			"high_end": {ID: "high_end", Type: "end"},
			"low_end":  {ID: "low_end", Type: "end"},
		},
		Edges: []*engine.FlowEdge{
			{ID: "e1", Source: "gw", Target: "high_end", SourceHandle: "high"},
			{ID: "e2", Source: "gw", Target: "low_end", SourceHandle: "low"},
		},
	}

	inst := &model.FlowInstance{ID: uuid.New(), InitiatorID: uuid.New()}
	vars := map[string]interface{}{"amount": 5000}

	gwHandler, err := registry.Get("exclusive_gateway")
	require.NoError(t, err)
	next, err := gwHandler.Execute(context.Background(), inst, graph.GetNode("gw"), graph, vars)
	require.NoError(t, err)
	assert.Equal(t, []string{"low_end"}, next)
}

// TestEndToEnd_ParallelGatewayFlow simulates a parallel split:
//
//   start → parallel_gateway → [branch_a → end, branch_b → end]

func TestEndToEnd_ParallelGatewayFlow(t *testing.T) {
	registry := NewNodeRegistry()
	registry.Register(&StartNode{})
	registry.Register(&EndNode{})
	registry.Register(&ParallelGatewayNode{})

	graph := &engine.FlowGraph{
		Nodes: map[string]*engine.FlowNode{
			"start": {ID: "start", Type: "start"},
			"gw":    {ID: "gw", Type: "parallel_gateway"},
			"a":     {ID: "a", Type: "end"},
			"b":     {ID: "b", Type: "end"},
		},
		Edges: []*engine.FlowEdge{
			{ID: "e1", Source: "start", Target: "gw"},
			{ID: "e2", Source: "gw", Target: "a"},
			{ID: "e3", Source: "gw", Target: "b"},
		},
	}

	inst := &model.FlowInstance{ID: uuid.New(), InitiatorID: uuid.New()}

	// Start → gw.
	startHandler, err := registry.Get("start")
	require.NoError(t, err)
	next, err := startHandler.Execute(context.Background(), inst, graph.GetNode("start"), graph, nil)
	require.NoError(t, err)
	assert.Equal(t, []string{"gw"}, next)

	// gw → both branches.
	gwHandler, err := registry.Get("parallel_gateway")
	require.NoError(t, err)
	next, err = gwHandler.Execute(context.Background(), inst, graph.GetNode("gw"), graph, nil)
	require.NoError(t, err)
	assert.Len(t, next, 2)
	assert.Contains(t, next, "a")
	assert.Contains(t, next, "b")
}

// TestEndToEnd_ServiceTaskFlow simulates:
//
//   start → service_task → end

func TestEndToEnd_ServiceTaskFlow(t *testing.T) {
	registry := NewNodeRegistry()

	executed := false
	serviceNode := &ServiceTaskNode{
		Callbacks: map[string]ServiceCallback{
			"sendEmail": func(ctx context.Context, inst *model.FlowInstance, node *engine.FlowNode, vars map[string]interface{}) (map[string]interface{}, error) {
				executed = true
				return map[string]interface{}{"sent": true}, nil
			},
		},
	}
	registry.Register(&StartNode{})
	registry.Register(&EndNode{})
	registry.Register(serviceNode)

	graph := &engine.FlowGraph{
		Nodes: map[string]*engine.FlowNode{
			"start": {ID: "start", Type: "start"},
			"task":  {ID: "task", Type: "service_task", Config: map[string]interface{}{"service": "sendEmail"}},
			"end":   {ID: "end", Type: "end"},
		},
		Edges: []*engine.FlowEdge{
			{ID: "e1", Source: "start", Target: "task"},
			{ID: "e2", Source: "task", Target: "end"},
		},
	}

	inst := &model.FlowInstance{ID: uuid.New(), InitiatorID: uuid.New()}
	vars := map[string]interface{}{}

	// Start → task.
	startHandler, err := registry.Get("start")
	require.NoError(t, err)
	next, err := startHandler.Execute(context.Background(), inst, graph.GetNode("start"), graph, vars)
	require.NoError(t, err)
	assert.Equal(t, []string{"task"}, next)

	// Task executes callback and → end.
	svcHandler, err := registry.Get("service_task")
	require.NoError(t, err)
	next, err = svcHandler.Execute(context.Background(), inst, graph.GetNode("task"), graph, vars)
	require.NoError(t, err)
	assert.Equal(t, []string{"end"}, next)
	assert.True(t, executed)
	assert.Equal(t, true, vars["sent"])
}
