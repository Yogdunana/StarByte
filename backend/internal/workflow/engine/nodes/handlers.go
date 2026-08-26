package nodes

import (
	"context"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/engine"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/model"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/events"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
)

// ApprovalNode handles the "approval" node type.
// When entered, it creates a FlowTask for the assignee(s) and pauses the flow.
// When the task is completed, the flow resumes to the next node.
type ApprovalNode struct {
	TaskRepo repo.TaskRepo
	EventBus *events.EventBus
}

func (n *ApprovalNode) Type() string { return "approval" }

func (n *ApprovalNode) Execute(ctx context.Context, inst *model.FlowInstance, node *engine.FlowNode, graph *engine.FlowGraph, vars map[string]interface{}) ([]string, error) {
	// Execute is called after task completion. Return next edges.
	edges := graph.GetNextNodes(node.ID, "")
	result := make([]string, 0, len(edges))
	for _, e := range edges {
		result = append(result, e.Target)
	}
	return result, nil
}

func (n *ApprovalNode) OnEnter(ctx context.Context, inst *model.FlowInstance, node *engine.FlowNode, vars map[string]interface{}) error {
	config := n.parseConfig(node)
	assignees := n.resolveAssignees(config, inst, vars)
	if len(assignees) == 0 {
		return response.NewAppError(response.CodeWorkflowInvalidNode,
			"approval node has no assignees")
	}

	// In v1, we support single approval only. Multi-assignee (all/any/ratio) is v2.
	for _, assigneeID := range assignees {
		task := &model.FlowTask{
			ID:         uuid.New(),
			InstanceID: inst.ID,
			NodeID:     node.ID,
			NodeName:   node.Label,
			TaskType:   "approval",
			AssigneeID: &assigneeID,
			Status:     0,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		if dueDays, ok := config["dueDays"].(float64); ok && dueDays > 0 {
			due := time.Now().Add(time.Duration(dueDays) * 24 * time.Hour)
			task.DueDate = &due
		}

		if err := n.TaskRepo.CreateTask(ctx, nil, task); err != nil {
			return err
		}

		// Publish TaskCreatedEvent.
		n.EventBus.Publish(ctx, events.TaskCreatedEvent{
			InstanceID: inst.ID,
			TaskID:     task.ID,
			AssigneeID: assigneeID,
			NodeID:     node.ID,
			NodeName:   node.Label,
			TaskType:   "approval",
		})
	}

	return nil
}

func (n *ApprovalNode) OnLeave(ctx context.Context, inst *model.FlowInstance, node *engine.FlowNode, vars map[string]interface{}) error {
	return nil
}

func (n *ApprovalNode) Validate(node *engine.FlowNode) error {
	config := n.parseConfig(node)
	if config == nil {
		return response.NewAppError(response.CodeWorkflowInvalidNode,
			"approval node missing config")
	}

	strategy, ok := config["assigneeStrategy"].(string)
	if !ok || strategy == "" {
		return response.NewAppError(response.CodeWorkflowInvalidNode,
			"approval node missing assigneeStrategy")
	}

	// Validate strategy-specific fields.
	switch strategy {
	case "static":
		assignees, ok := config["assignees"].([]interface{})
		if !ok || len(assignees) == 0 {
			return response.NewAppError(response.CodeWorkflowInvalidNode,
				"static assignee strategy requires non-empty assignees list")
		}
	case "role":
		if _, ok := config["roleId"].(string); !ok {
			return response.NewAppError(response.CodeWorkflowInvalidNode,
				"role assignee strategy requires roleId")
		}
	case "dept_leader", "initiator":
		// No additional fields required.
	default:
		return response.NewAppErrorf(response.CodeWorkflowNodeType,
			"unsupported assigneeStrategy: %s", strategy)
	}

	return nil
}

func (n *ApprovalNode) parseConfig(node *engine.FlowNode) map[string]interface{} {
	if node.Config == nil {
		return nil
	}
	return node.Config
}

func (n *ApprovalNode) resolveAssignees(config map[string]interface{}, inst *model.FlowInstance, vars map[string]interface{}) []uuid.UUID {
	var result []uuid.UUID

	strategy, _ := config["assigneeStrategy"].(string)
	switch strategy {
	case "static":
		if assignees, ok := config["assignees"].([]interface{}); ok {
			for _, a := range assignees {
				if s, ok := a.(string); ok {
					if id, err := uuid.Parse(s); err == nil {
						result = append(result, id)
					}
				}
			}
		}
	case "initiator":
		result = append(result, inst.InitiatorID)
	case "role", "dept_leader":
		// These strategies require external services (RBAC) to resolve.
		// In v1, we pass the roleId/dept config via variables for the
		// service layer to resolve. For now, return empty — the service
		// layer will handle this before calling OnEnter.
	}

	return result
}

// --- ExclusiveGatewayNode ---

// ExclusiveGatewayNode handles the "exclusive_gateway" node type.
// It evaluates branch expressions and selects exactly one outgoing path.
type ExclusiveGatewayNode struct {
	ExprEngine *engine.ExpressionEngine
}

func (n *ExclusiveGatewayNode) Type() string { return "exclusive_gateway" }

func (n *ExclusiveGatewayNode) Execute(ctx context.Context, inst *model.FlowInstance, node *engine.FlowNode, graph *engine.FlowGraph, vars map[string]interface{}) ([]string, error) {
	branches, err := engine.ParseBranches(node.Config)
	if err != nil {
		return nil, err
	}

	branchID, err := n.ExprEngine.EvaluateBranch(branches, vars)
	if err != nil {
		return nil, err
	}

	// Find the edge whose sourceHandle matches the selected branch ID.
	edges := graph.GetNextNodes(node.ID, branchID)
	if len(edges) == 0 {
		// Fall back to any outgoing edge if sourceHandle matching fails.
		edges = graph.GetNextNodes(node.ID, "")
	}

	result := make([]string, 0, len(edges))
	for _, e := range edges {
		result = append(result, e.Target)
	}

	if len(result) == 0 {
		return nil, response.NewAppErrorf(response.CodeWorkflowNodeNotFound,
			"exclusive gateway '%s' has no outgoing edge for branch '%s'", node.ID, branchID)
	}

	return result, nil
}

func (n *ExclusiveGatewayNode) OnEnter(ctx context.Context, inst *model.FlowInstance, node *engine.FlowNode, vars map[string]interface{}) error {
	return nil
}

func (n *ExclusiveGatewayNode) OnLeave(ctx context.Context, inst *model.FlowInstance, node *engine.FlowNode, vars map[string]interface{}) error {
	return nil
}

func (n *ExclusiveGatewayNode) Validate(node *engine.FlowNode) error {
	if node.Config == nil {
		return response.NewAppError(response.CodeWorkflowInvalidNode,
			"exclusive gateway missing config")
	}

	branches, err := engine.ParseBranches(node.Config)
	if err != nil {
		return err
	}

	if len(branches) < 2 {
		return response.NewAppError(response.CodeWorkflowInvalidNode,
			"exclusive gateway must have at least 2 branches")
	}

	return nil
}

// --- ParallelGatewayNode ---

// ParallelGatewayNode handles the "parallel_gateway" node type.
// All outgoing branches are activated simultaneously.
type ParallelGatewayNode struct{}

func (ParallelGatewayNode) Type() string { return "parallel_gateway" }

func (ParallelGatewayNode) Execute(ctx context.Context, inst *model.FlowInstance, node *engine.FlowNode, graph *engine.FlowGraph, vars map[string]interface{}) ([]string, error) {
	// Activate all outgoing edges.
	edges := graph.GetNextNodes(node.ID, "")
	result := make([]string, 0, len(edges))
	for _, e := range edges {
		result = append(result, e.Target)
	}

	if len(result) == 0 {
		return nil, response.NewAppErrorf(response.CodeWorkflowNodeNotFound,
			"parallel gateway '%s' has no outgoing edges", node.ID)
	}

	return result, nil
}

func (ParallelGatewayNode) OnEnter(ctx context.Context, inst *model.FlowInstance, node *engine.FlowNode, vars map[string]interface{}) error {
	return nil
}

func (ParallelGatewayNode) OnLeave(ctx context.Context, inst *model.FlowInstance, node *engine.FlowNode, vars map[string]interface{}) error {
	return nil
}

func (ParallelGatewayNode) Validate(node *engine.FlowNode) error {
	// Parallel gateway validation is done at graph level (must have >= 2 outgoing edges).
	return nil
}

// --- ServiceTaskNode ---

// ServiceTaskNode handles the "service_task" node type.
// It calls an external API or executes business logic automatically.
// In v1, the actual service call is delegated to a callback registered by the business module.
type ServiceTaskNode struct {
	// Callbacks maps node config "service" to a handler function.
	// Business modules register callbacks via RegisterService.
	Callbacks map[string]ServiceCallback
}

// ServiceCallback is a function that executes a service task.
type ServiceCallback func(ctx context.Context, inst *model.FlowInstance, node *engine.FlowNode, vars map[string]interface{}) (map[string]interface{}, error)

func (n *ServiceTaskNode) Type() string { return "service_task" }

func (n *ServiceTaskNode) Execute(ctx context.Context, inst *model.FlowInstance, node *engine.FlowNode, graph *engine.FlowGraph, vars map[string]interface{}) ([]string, error) {
	serviceName, _ := node.Config["service"].(string)
	if serviceName == "" {
		return nil, response.NewAppError(response.CodeWorkflowInvalidNode,
			"service_task missing 'service' in config")
	}

	cb, ok := n.Callbacks[serviceName]
	if !ok {
		return nil, response.NewAppErrorf(response.CodeWorkflowNodeType,
			"no callback registered for service '%s'", serviceName)
	}

	// Execute the service callback.
	result, err := cb(ctx, inst, node, vars)
	if err != nil {
		return nil, err
	}

	// Merge callback results into variables.
	for k, v := range result {
		vars[k] = v
	}

	// Proceed to next node.
	edges := graph.GetNextNodes(node.ID, "")
	resultIDs := make([]string, 0, len(edges))
	for _, e := range edges {
		resultIDs = append(resultIDs, e.Target)
	}
	return resultIDs, nil
}

func (n *ServiceTaskNode) OnEnter(ctx context.Context, inst *model.FlowInstance, node *engine.FlowNode, vars map[string]interface{}) error {
	return nil
}

func (n *ServiceTaskNode) OnLeave(ctx context.Context, inst *model.FlowInstance, node *engine.FlowNode, vars map[string]interface{}) error {
	return nil
}

func (n *ServiceTaskNode) Validate(node *engine.FlowNode) error {
	if node.Config == nil {
		return response.NewAppError(response.CodeWorkflowInvalidNode,
			"service_task missing config")
	}
	if _, ok := node.Config["service"].(string); !ok {
		return response.NewAppError(response.CodeWorkflowInvalidNode,
			"service_task missing 'service' field in config")
	}
	return nil
}

// --- NotificationTaskNode ---

// NotificationTaskNode handles the "notification_task" node type.
// It publishes a notification event that the notification service can subscribe to.
type NotificationTaskNode struct {
	EventBus *events.EventBus
}

func (n *NotificationTaskNode) Type() string { return "notification_task" }

func (n *NotificationTaskNode) Execute(ctx context.Context, inst *model.FlowInstance, node *engine.FlowNode, graph *engine.FlowGraph, vars map[string]interface{}) ([]string, error) {
	// The notification logic is handled by the notification service
	// subscribing to NodeEnteredEvent. Here we just proceed.
	// Alternatively, we could publish a dedicated notification event here.

	// Extract notification config.
	notifType, _ := node.Config["notificationType"].(string)
	if notifType == "" {
		notifType = "default"
	}

	// Store the notification type in variables for the notification service.
	vars["_notification_type"] = notifType
	vars["_notification_node_id"] = node.ID

	// Proceed to next node.
	edges := graph.GetNextNodes(node.ID, "")
	result := make([]string, 0, len(edges))
	for _, e := range edges {
		result = append(result, e.Target)
	}
	return result, nil
}

func (n *NotificationTaskNode) OnEnter(ctx context.Context, inst *model.FlowInstance, node *engine.FlowNode, vars map[string]interface{}) error {
	return nil
}

func (n *NotificationTaskNode) OnLeave(ctx context.Context, inst *model.FlowInstance, node *engine.FlowNode, vars map[string]interface{}) error {
	return nil
}

func (n *NotificationTaskNode) Validate(node *engine.FlowNode) error {
	// Notification task doesn't require strict config validation in v1.
	return nil
}

// --- DefaultRegistry ---

// NewDefaultRegistry creates a NodeRegistry with all built-in node handlers registered.
// Callbacks for service tasks and the event bus are wired by the caller.
func NewDefaultRegistry(exprEngine *engine.ExpressionEngine, eventBus *events.EventBus, taskRepo repo.TaskRepo) *NodeRegistry {
	registry := NewNodeRegistry()
	registry.Register(StartNode{})
	registry.Register(EndNode{})
	registry.Register(&ApprovalNode{
		TaskRepo: taskRepo,
		EventBus: eventBus,
	})
	registry.Register(&ExclusiveGatewayNode{
		ExprEngine: exprEngine,
	})
	registry.Register(ParallelGatewayNode{})
	registry.Register(&ServiceTaskNode{
		Callbacks: make(map[string]ServiceCallback),
	})
	registry.Register(&NotificationTaskNode{
		EventBus: eventBus,
	})
	return registry
}
