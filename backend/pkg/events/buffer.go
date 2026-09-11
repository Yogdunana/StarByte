package events

import (
	"context"
	"sync"
)

// Buffer collects events for a database unit of work. Flush only after commit;
// dropping the buffer on rollback prevents notifications for changes that failed.
type Buffer struct {
	mu     sync.Mutex
	events []Event
}

func NewBufferedBus() (*EventBus, *Buffer) {
	buffer := &Buffer{}
	bus := NewEventBus()
	bus.buffer = buffer
	return bus, buffer
}
func (b *Buffer) add(event Event) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.events = append(b.events, event)
}
func (b *Buffer) Flush(ctx context.Context, target *EventBus) []error {
	b.mu.Lock()
	pending := b.events
	b.events = nil
	b.mu.Unlock()
	var errs []error
	for _, event := range pending {
		errs = append(errs, target.Publish(ctx, event)...)
	}
	return errs
}
