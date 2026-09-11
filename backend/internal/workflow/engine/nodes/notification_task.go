package nodes

import (
	"context"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/engine"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/model"
	"github.com/Yogdunana/StarByte/backend/pkg/events"
)

// --- NotificationTaskNode ---

// NotificationTaskNode handles the "notification_task" node type.
// It publishes a notification event that the notification service can subscribe to.
type NotificationTaskNode struct {
	EventBus *events.EventBus
}

func (n *NotificationTaskNode) Type() string { return "notification_task" }

func (n *NotificationTaskNode) Execute(ctx context.Context, inst *model.FlowInstance, node *engine.FlowNode, graph *engine.FlowGraph, vars map[string]interface{}) ([]string, error) {
	// Extract notification config and publish a dedicated event.
	notifType, _ := node.Config["notificationType"].(string)
	if notifType == "" {
		notifType = "default"
	}

	n.EventBus.Publish(ctx, events.NotificationTaskTriggeredEvent{
		InstanceID:       inst.ID,
		NodeID:           node.ID,
		NodeName:         node.Label,
		NotificationType: notifType,
	})

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
