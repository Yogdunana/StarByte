package service

import (
	"context"
	"testing"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/feature/model"
	"github.com/google/uuid"
)

func TestMemorySnapshotAndBus(t *testing.T) {
	store := NewMemorySnapshot()
	ctx := context.Background()
	if err := store.Set(ctx, []model.Flag{{ID: uuid.New(), FlagKey: "cms.public"}}); err != nil {
		t.Fatal(err)
	}
	got, err := store.Get(ctx)
	if err != nil || len(got) != 1 || got[0].FlagKey != "cms.public" {
		t.Fatalf("get: %+v %v", got, err)
	}
	if err := store.Delete(ctx); err != nil {
		t.Fatal(err)
	}
	got, err = store.Get(ctx)
	if err != nil || got != nil {
		t.Fatalf("deleted: %+v %v", got, err)
	}

	bus := NewMemoryBus()
	hit := make(chan string, 1)
	subCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	if err := bus.Subscribe(subCtx, func(msg string) { hit <- msg }); err != nil {
		t.Fatal(err)
	}
	if err := bus.Publish(ctx, "ping"); err != nil {
		t.Fatal(err)
	}
	select {
	case msg := <-hit:
		if msg != "ping" {
			t.Fatalf("msg=%s", msg)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for pubsub")
	}
}

func TestNilRedisFallsBackToMemory(t *testing.T) {
	if _, ok := NewRedisSnapshot(nil).(*memorySnapshot); !ok {
		t.Fatal("snapshot fallback")
	}
	if _, ok := NewRedisBus(nil).(*MemoryBus); !ok {
		t.Fatal("bus fallback")
	}
}
