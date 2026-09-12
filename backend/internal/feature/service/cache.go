package service

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/feature/model"
	"github.com/Yogdunana/StarByte/backend/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	snapshotKey = "feature:snapshot"
	channelName = "feature:invalidate"
	snapshotTTL = 10 * time.Minute
)

// Broadcaster notifies other processes that flags changed.
type Broadcaster interface {
	Publish(ctx context.Context, payload string) error
	Subscribe(ctx context.Context, onMsg func(string)) error
}

// SnapshotStore is the Redis (or memory) copy of all flags.
type SnapshotStore interface {
	Get(ctx context.Context) ([]model.Flag, error)
	Set(ctx context.Context, flags []model.Flag) error
	Delete(ctx context.Context) error
}

type memorySnapshot struct {
	mu    sync.RWMutex
	flags []model.Flag
}

func NewMemorySnapshot() SnapshotStore { return &memorySnapshot{} }

func (m *memorySnapshot) Get(context.Context) ([]model.Flag, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return cloneFlags(m.flags), nil
}

func (m *memorySnapshot) Set(_ context.Context, flags []model.Flag) error {
	m.mu.Lock()
	m.flags = cloneFlags(flags)
	m.mu.Unlock()
	return nil
}

func (m *memorySnapshot) Delete(context.Context) error {
	m.mu.Lock()
	m.flags = nil
	m.mu.Unlock()
	return nil
}

type MemoryBus struct {
	mu   sync.Mutex
	subs []chan string
}

func NewMemoryBus() *MemoryBus { return &MemoryBus{} }

func (b *MemoryBus) Publish(_ context.Context, payload string) error {
	b.mu.Lock()
	subs := append([]chan string(nil), b.subs...)
	b.mu.Unlock()
	for _, ch := range subs {
		select {
		case ch <- payload:
		default:
		}
	}
	return nil
}

func (b *MemoryBus) Subscribe(ctx context.Context, onMsg func(string)) error {
	ch := make(chan string, 8)
	b.mu.Lock()
	b.subs = append(b.subs, ch)
	b.mu.Unlock()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-ch:
				if onMsg != nil {
					onMsg(msg)
				}
			}
		}
	}()
	return nil
}

type redisSnapshot struct{ rdb *redis.Client }

func NewRedisSnapshot(rdb *redis.Client) SnapshotStore {
	if rdb == nil {
		return NewMemorySnapshot()
	}
	return &redisSnapshot{rdb: rdb}
}

func (s *redisSnapshot) Get(ctx context.Context) ([]model.Flag, error) {
	raw, err := s.rdb.Get(ctx, snapshotKey).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var flags []model.Flag
	if err := json.Unmarshal(raw, &flags); err != nil {
		return nil, err
	}
	return flags, nil
}

func (s *redisSnapshot) Set(ctx context.Context, flags []model.Flag) error {
	raw, err := json.Marshal(flags)
	if err != nil {
		return err
	}
	return s.rdb.Set(ctx, snapshotKey, raw, snapshotTTL).Err()
}

func (s *redisSnapshot) Delete(ctx context.Context) error {
	return s.rdb.Del(ctx, snapshotKey).Err()
}

type redisBus struct{ rdb *redis.Client }

func NewRedisBus(rdb *redis.Client) Broadcaster {
	if rdb == nil {
		return NewMemoryBus()
	}
	return &redisBus{rdb: rdb}
}

func (b *redisBus) Publish(ctx context.Context, payload string) error {
	return b.rdb.Publish(ctx, channelName, payload).Err()
}

func (b *redisBus) Subscribe(ctx context.Context, onMsg func(string)) error {
	if b.rdb == nil {
		return nil
	}
	pubsub := b.rdb.Subscribe(ctx, channelName)
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
				if msg != nil && onMsg != nil {
					onMsg(msg.Payload)
				}
			}
		}
	}()
	return nil
}

func cloneFlags(in []model.Flag) []model.Flag {
	if in == nil {
		return nil
	}
	out := make([]model.Flag, len(in))
	copy(out, in)
	return out
}

func logCacheErr(op string, err error) {
	if err == nil {
		return
	}
	logger.Warn("feature cache "+op, zap.Error(err))
}
