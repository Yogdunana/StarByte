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
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
	q := Query{Denied: true, DepartmentID: &id, DeptIDs: []uuid.UUID{id}, HideRanking: true, Start: &start, End: &end}
	assert.True(t, q.Denied)
	assert.Equal(t, id, *q.DepartmentID)
	assert.Equal(t, []uuid.UUID{id}, q.DeptIDs)
	assert.True(t, q.HideRanking)
	assert.Contains(t, clippedDaysSQL(q), "2026-01-01")
	assert.Contains(t, clippedDaysSQL(q), "2026-02-01")
	assert.NotContains(t, clippedDaysSQL(q), "DATE '2026-01-31'")
	endOfDay := time.Date(2026, 1, 31, 23, 59, 59, 999999999, time.UTC)
	today := time.Date(2026, 9, 6, 23, 59, 59, 999999999, time.Local)
	assert.Contains(t, clippedDaysSQL(Query{Start: &endOfDay, End: &endOfDay}), "2026-02-01")
	assert.Contains(t, clippedDaysSQL(Query{Start: &today, End: &today}), "2026-09-07")
	assert.Equal(t, "GREATEST(0, ((COALESCE(i.end_date, CURRENT_DATE) + 1) - i.start_date))", clippedDaysSQL(Query{}))
}
