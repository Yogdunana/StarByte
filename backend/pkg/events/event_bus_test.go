package events

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeEvent is a minimal Event implementation for testing.
type fakeEvent struct {
	name       string
	instanceID uuid.UUID
}

func (e fakeEvent) EventName() string        { return e.name }
func (e fakeEvent) GetInstanceID() uuid.UUID { return e.instanceID }

func TestEventBus_SubscribeAndPublish(t *testing.T) {
	bus := NewEventBus()
	instID := uuid.New()
	called := false

	bus.Subscribe("test.event", func(ctx context.Context, e Event) error {
		called = true
		assert.Equal(t, "test.event", e.EventName())
		assert.Equal(t, instID, e.GetInstanceID())
		return nil
	})

	errs := bus.Publish(context.Background(), fakeEvent{name: "test.event", instanceID: instID})
	assert.Empty(t, errs)
	assert.True(t, called)
}

func TestEventBus_MultipleHandlers(t *testing.T) {
	bus := NewEventBus()
	instID := uuid.New()
	count := 0

	bus.Subscribe("test.event", func(ctx context.Context, e Event) error {
		count++
		return nil
	})
	bus.Subscribe("test.event", func(ctx context.Context, e Event) error {
		count++
		return nil
	})

	errs := bus.Publish(context.Background(), fakeEvent{name: "test.event", instanceID: instID})
	assert.Empty(t, errs)
	assert.Equal(t, 2, count)
}

func TestEventBus_HandlerError(t *testing.T) {
	bus := NewEventBus()
	instID := uuid.New()

	bus.Subscribe("test.event", func(ctx context.Context, e Event) error {
		return errors.New("handler error 1")
	})
	bus.Subscribe("test.event", func(ctx context.Context, e Event) error {
		return errors.New("handler error 2")
	})

	errs := bus.Publish(context.Background(), fakeEvent{name: "test.event", instanceID: instID})
	require.Len(t, errs, 2)
	assert.Contains(t, errs[0].Error(), "handler error 1")
	assert.Contains(t, errs[1].Error(), "handler error 2")
}

func TestEventBus_NoSubscribers(t *testing.T) {
	bus := NewEventBus()
	instID := uuid.New()

	errs := bus.Publish(context.Background(), fakeEvent{name: "unsubscribed.event", instanceID: instID})
	assert.Empty(t, errs)
}

func TestEventBus_HasSubscribers(t *testing.T) {
	bus := NewEventBus()

	assert.False(t, bus.HasSubscribers("test.event"))

	bus.Subscribe("test.event", func(ctx context.Context, e Event) error {
		return nil
	})

	assert.True(t, bus.HasSubscribers("test.event"))
	assert.False(t, bus.HasSubscribers("other.event"))
}

func TestEventBus_DifferentEventNames(t *testing.T) {
	bus := NewEventBus()
	instID := uuid.New()
	calledA := false
	calledB := false

	bus.Subscribe("event.a", func(ctx context.Context, e Event) error {
		calledA = true
		return nil
	})
	bus.Subscribe("event.b", func(ctx context.Context, e Event) error {
		calledB = true
		return nil
	})

	bus.Publish(context.Background(), fakeEvent{name: "event.a", instanceID: instID})
	assert.True(t, calledA)
	assert.False(t, calledB)
}

func TestFlowStartedEvent(t *testing.T) {
	instID := uuid.New()
	defID := uuid.New()
	initID := uuid.New()

	e := FlowStartedEvent{
		InstanceID:   instID,
		DefinitionID: defID,
		InitiatorID:  initID,
		BusinessKey:  "BK001",
		BusinessType: "leave",
	}
	assert.Equal(t, "flow.started", e.EventName())
	assert.Equal(t, instID, e.GetInstanceID())
}

func TestTaskCreatedEvent(t *testing.T) {
	instID := uuid.New()
	taskID := uuid.New()
	assigneeID := uuid.New()

	e := TaskCreatedEvent{
		InstanceID: instID,
		TaskID:     taskID,
		AssigneeID: assigneeID,
		NodeID:     "node-1",
		NodeName:   "审批节点",
		TaskType:   "approval",
	}
	assert.Equal(t, "task.created", e.EventName())
	assert.Equal(t, instID, e.GetInstanceID())
}
