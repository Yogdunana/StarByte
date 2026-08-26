package events

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Event is the base interface for all workflow events.
type Event interface {
	EventName() string
	GetInstanceID() uuid.UUID
}

// FlowStartedEvent is published when a flow instance is started.
type FlowStartedEvent struct {
	InstanceID   uuid.UUID
	DefinitionID uuid.UUID
	InitiatorID  uuid.UUID
	BusinessKey  string
	BusinessType string
	StartedAt    time.Time
}

func (e FlowStartedEvent) EventName() string        { return "flow.started" }
func (e FlowStartedEvent) GetInstanceID() uuid.UUID { return e.InstanceID }

// FlowCompletedEvent is published when a flow instance completes normally.
type FlowCompletedEvent struct {
	InstanceID uuid.UUID
	EndedAt    time.Time
}

func (e FlowCompletedEvent) EventName() string        { return "flow.completed" }
func (e FlowCompletedEvent) GetInstanceID() uuid.UUID { return e.InstanceID }

// FlowTerminatedEvent is published when a flow instance is terminated.
type FlowTerminatedEvent struct {
	InstanceID uuid.UUID
	Reason     string
	OperatorID uuid.UUID
	EndedAt    time.Time
}

func (e FlowTerminatedEvent) EventName() string        { return "flow.terminated" }
func (e FlowTerminatedEvent) GetInstanceID() uuid.UUID { return e.InstanceID }

// TaskCreatedEvent is published when a new flow task is created.
type TaskCreatedEvent struct {
	InstanceID uuid.UUID
	TaskID     uuid.UUID
	AssigneeID uuid.UUID
	NodeID     string
	NodeName   string
	TaskType   string
}

func (e TaskCreatedEvent) EventName() string        { return "task.created" }
func (e TaskCreatedEvent) GetInstanceID() uuid.UUID { return e.InstanceID }

// TaskCompletedEvent is published when a flow task is completed.
type TaskCompletedEvent struct {
	InstanceID uuid.UUID
	TaskID     uuid.UUID
	OperatorID uuid.UUID
	Action     string
	Comment    string
}

func (e TaskCompletedEvent) EventName() string        { return "task.completed" }
func (e TaskCompletedEvent) GetInstanceID() uuid.UUID { return e.InstanceID }

// TaskAssignedEvent is published when a task is transferred to a new assignee.
type TaskAssignedEvent struct {
	InstanceID    uuid.UUID
	TaskID        uuid.UUID
	OldAssigneeID uuid.UUID
	NewAssigneeID uuid.UUID
}

func (e TaskAssignedEvent) EventName() string        { return "task.assigned" }
func (e TaskAssignedEvent) GetInstanceID() uuid.UUID { return e.InstanceID }

// NotificationTaskTriggeredEvent is published when a notification task node is executed.
type NotificationTaskTriggeredEvent struct {
	InstanceID       uuid.UUID
	NodeID           string
	NodeName         string
	NotificationType string
}

func (e NotificationTaskTriggeredEvent) EventName() string        { return "notification.triggered" }
func (e NotificationTaskTriggeredEvent) GetInstanceID() uuid.UUID { return e.InstanceID }

// NodeEnteredEvent is published when the flow enters a node.
type NodeEnteredEvent struct {
	InstanceID uuid.UUID
	NodeID     string
	NodeType   string
}

func (e NodeEnteredEvent) EventName() string        { return "node.entered" }
func (e NodeEnteredEvent) GetInstanceID() uuid.UUID { return e.InstanceID }

// NodeLeftEvent is published when the flow leaves a node.
type NodeLeftEvent struct {
	InstanceID uuid.UUID
	NodeID     string
	NodeType   string
}

func (e NodeLeftEvent) EventName() string        { return "node.left" }
func (e NodeLeftEvent) GetInstanceID() uuid.UUID { return e.InstanceID }

// Handler is a function that processes an event.
type Handler func(ctx context.Context, event Event) error

// EventBus provides publish/subscribe functionality for workflow events.
// In v1, dispatch is synchronous. v2 may add async dispatch with a worker pool.
type EventBus struct {
	mu       sync.RWMutex
	handlers map[string][]Handler
}

// NewEventBus creates a new EventBus.
func NewEventBus() *EventBus {
	return &EventBus{
		handlers: make(map[string][]Handler),
	}
}

// Subscribe registers a handler for a specific event name.
// Multiple handlers can be registered for the same event.
func (b *EventBus) Subscribe(eventName string, handler Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventName] = append(b.handlers[eventName], handler)
}

// Publish dispatches an event to all registered handlers synchronously.
// If a handler returns an error, subsequent handlers are still called.
// Errors are collected and returned as a slice.
func (b *EventBus) Publish(ctx context.Context, event Event) []error {
	b.mu.RLock()
	handlers := b.handlers[event.EventName()]
	b.mu.RUnlock()

	var errs []error
	for _, h := range handlers {
		if err := h(ctx, event); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

// HasSubscribers returns true if any handler is registered for the event.
func (b *EventBus) HasSubscribers(eventName string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.handlers[eventName]) > 0
}
