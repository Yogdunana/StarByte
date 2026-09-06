package repo

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestTruncExpr(t *testing.T) {
	if got := truncExpr("created_at", "day"); got != "to_char(date_trunc('day', created_at), 'YYYY-MM-DD')" {
		t.Fatalf("day: %s", got)
	}
	if got := truncExpr("m.start_time", "week"); got != "to_char(date_trunc('week', m.start_time), 'YYYY-MM-DD')" {
		t.Fatalf("week: %s", got)
	}
	if got := truncExpr("created_at", "MONTH"); got != "to_char(date_trunc('month', created_at), 'YYYY-MM-DD')" {
		t.Fatalf("month: %s", got)
	}
	if got := truncExpr("created_at", ""); got != "to_char(date_trunc('month', created_at), 'YYYY-MM-DD')" {
		t.Fatalf("default: %s", got)
	}
}

func TestQueryFlags(t *testing.T) {
	id := uuid.New()
	q := Query{Denied: true, DepartmentID: &id, DeptIDs: []uuid.UUID{id}, AllScope: false, HideRanking: true}
	assert.True(t, q.Denied)
	assert.True(t, q.HideRanking)
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
	overlap := Query{Start: &start, End: &end}
	assert.False(t, overlap.End.Before(*overlap.Start))
}
