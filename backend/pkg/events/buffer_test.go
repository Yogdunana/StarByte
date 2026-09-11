package events

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBufferedEventsAreInvisibleUntilCommitFlush(t *testing.T) {
	target := NewEventBus()
	count := 0
	target.Subscribe("flow.started", func(context.Context, Event) error { count++; return nil })
	bus, buffer := NewBufferedBus()
	bus.Publish(context.Background(), FlowStartedEvent{})
	require.Zero(t, count)
	require.Empty(t, buffer.Flush(context.Background(), target))
	require.Equal(t, 1, count)
	require.Empty(t, buffer.Flush(context.Background(), target))
	require.Equal(t, 1, count)
	// An abandoned transaction never flushes its separate buffer.
	abandoned, _ := NewBufferedBus()
	abandoned.Publish(context.Background(), FlowStartedEvent{})
	require.Equal(t, 1, count)
}
