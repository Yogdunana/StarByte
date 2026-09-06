package cache

import "context"

// EnableKeyspaceEvents turns on Redis expired-key notifications (Ex).
func (s *Store) EnableKeyspaceEvents(ctx context.Context) error {
	return s.rdb.ConfigSet(ctx, "notify-keyspace-events", "Ex").Err()
}

// SubscribeExpired calls onExpired with the key name when Redis publishes
// a keyevent expired notification. The goroutine exits when ctx is cancelled.
func (s *Store) SubscribeExpired(ctx context.Context, onExpired func(key string)) error {
	if s == nil || s.rdb == nil {
		return nil
	}
	_ = s.EnableKeyspaceEvents(ctx)
	pubsub := s.rdb.PSubscribe(ctx, "__keyevent@*:expired")
	go func() {
		defer func() { _ = pubsub.Close() }()
		ch := pubsub.Channel()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				if msg != nil && onExpired != nil {
					onExpired(msg.Payload)
				}
			}
		}
	}()
	return nil
}
