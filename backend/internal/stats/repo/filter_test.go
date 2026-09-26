package repo

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// 分组必须先把 UTC 墙钟折成北京时间，否则「每天」的分界会落在北京时间 08:00。
// 括号同样是语法的一部分，去掉就会变成把时区名强转成 date。
const (
	shanghaiCreatedAt = "(created_at AT TIME ZONE 'UTC' AT TIME ZONE 'Asia/Shanghai')"
	shanghaiStartAt   = "(m.start_time AT TIME ZONE 'UTC' AT TIME ZONE 'Asia/Shanghai')"
)

func TestTruncExpr(t *testing.T) {
	if got := truncExpr("created_at", "day"); got != "to_char(date_trunc('day', "+shanghaiCreatedAt+"), 'YYYY-MM-DD')" {
		t.Fatalf("day: %s", got)
	}
	if got := truncExpr("m.start_time", "week"); got != "to_char(date_trunc('week', "+shanghaiStartAt+"), 'YYYY-MM-DD')" {
		t.Fatalf("week: %s", got)
	}
	if got := truncExpr("created_at", "MONTH"); got != "to_char(date_trunc('month', "+shanghaiCreatedAt+"), 'YYYY-MM-DD')" {
		t.Fatalf("month: %s", got)
	}
	if got := truncExpr("created_at", ""); got != "to_char(date_trunc('month', "+shanghaiCreatedAt+"), 'YYYY-MM-DD')" {
		t.Fatalf("default: %s", got)
	}
}

// 调用方会在后面直接拼 ::date（见 ops.go 的会议日历），所以返回值必须是完整的
// 括号表达式，不能是裸的 AT TIME ZONE 片段。
func TestShanghaiWallClock_Parenthesized(t *testing.T) {
	got := shanghaiWallClock("m.start_time")
	assert.Equal(t, shanghaiStartAt, got)
	assert.True(t, strings.HasPrefix(got, "(") && strings.HasSuffix(got, ")"),
		"缺括号时拼 ::date 会被解析成 AT TIME ZONE 'Asia/Shanghai'::date: %s", got)
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
