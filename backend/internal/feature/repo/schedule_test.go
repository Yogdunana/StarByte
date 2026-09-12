package repo

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/feature/model"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestScheduledStateCompareAndSwapPostgres(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		if os.Getenv("TEST_DATABASE_REQUIRED") == "1" {
			t.Fatal("TEST_DATABASE_URL required")
		}
		t.Skip("set TEST_DATABASE_URL to a migrated test database")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer raw.Close()
	tx := db.Begin()
	if tx.Error != nil {
		t.Fatal(tx.Error)
	}
	defer tx.Rollback()
	r := New(tx)
	ctx := context.Background()
	now := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)
	flag := &model.Flag{ID: uuid.New(), FlagKey: "test.cas." + uuid.NewString(),
		Name: "Original", FlagType: model.TypeBoolean, CreatedAt: now, UpdatedAt: now}
	if err := r.Create(ctx, flag); err != nil {
		t.Fatal(err)
	}
	// A manual edit lands after the scheduler's read.
	flag.Name, flag.UpdatedAt = "Manual", now.Add(time.Second)
	if err := r.Update(ctx, flag); err != nil {
		t.Fatal(err)
	}
	if changed, err := r.UpdateScheduledState(ctx, flag.ID, now, true, now.Add(2*time.Second)); err != nil || changed {
		t.Fatalf("stale write accepted: %v %v", changed, err)
	}
	fresh, err := r.GetByID(ctx, flag.ID)
	if err != nil || fresh == nil {
		t.Fatalf("load: %v", err)
	}
	if fresh.Enabled || fresh.Name != "Manual" {
		t.Fatalf("manual edit lost: %+v", fresh)
	}
	if changed, err := r.UpdateScheduledState(ctx, flag.ID, fresh.UpdatedAt, true, now.Add(3*time.Second)); err != nil || !changed {
		t.Fatalf("fresh write skipped: %v %v", changed, err)
	}
	if changed, err := r.UpdateScheduledState(ctx, flag.ID, fresh.UpdatedAt, true, now.Add(4*time.Second)); err != nil || changed {
		t.Fatalf("duplicate write accepted: %v %v", changed, err)
	}
	got, err := r.GetByID(ctx, flag.ID)
	if err != nil || got == nil || !got.Enabled || got.Name != "Manual" {
		t.Fatalf("state or unrelated field incorrect: %+v %v", got, err)
	}
}
