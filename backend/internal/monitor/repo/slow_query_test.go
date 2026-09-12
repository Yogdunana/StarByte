package repo

import (
	"context"
	"testing"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/monitor/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListWithoutDBUsesInMemory(t *testing.T) {
	database.ResetSlowQueriesForTest()
	t.Cleanup(database.ResetSlowQueriesForTest)

	r := NewSlowQueryRepo(nil)
	out := r.List(context.Background(), 0)
	require.NotNil(t, out)
	assert.True(t, out.Available)
	assert.Equal(t, "in_memory", out.Source)
	assert.Empty(t, out.Queries)
	assert.Contains(t, out.Note, "pg_stat_statements")

	database.RecordSlowQueryForTest("SELECT * FROM members", time.Second, 3)
	filled := r.List(context.Background(), 10)
	require.Len(t, filled.Queries, 1)
	assert.Equal(t, "in_memory", filled.Queries[0].Source)
	assert.Equal(t, "SELECT * FROM members", filled.Queries[0].Query)
}

type fakePG struct {
	hasExt      bool
	hasErr      error
	stmts       []dto.SlowQuery
	stmtErr     error
	activity    []dto.SlowQuery
	activityErr error
}

func (f fakePG) hasExtension(context.Context, string) (bool, error) {
	return f.hasExt, f.hasErr
}
func (f fakePG) listStatements(context.Context, int) ([]dto.SlowQuery, error) {
	return f.stmts, f.stmtErr
}
func (f fakePG) listActivity(context.Context, int) ([]dto.SlowQuery, error) {
	return f.activity, f.activityErr
}

func TestListPrefersPgStatStatements(t *testing.T) {
	r := &slowQueryRepo{
		querier: fakePG{
			hasExt: true,
			stmts:  []dto.SlowQuery{{Query: "SELECT 1", Source: "pg_stat_statements", Calls: 9}},
		},
		now: time.Now,
	}
	out := r.List(context.Background(), 10)
	assert.Equal(t, "pg_stat_statements", out.Source)
	require.Len(t, out.Queries, 1)
	assert.Equal(t, int64(9), out.Queries[0].Calls)
}

func TestListFallsBackWhenExtensionMissing(t *testing.T) {
	r := &slowQueryRepo{
		querier: fakePG{
			activity: []dto.SlowQuery{{Query: "SELECT pg_sleep(1)", Source: "pg_stat_activity"}},
		},
		now: time.Now,
	}
	out := r.List(context.Background(), 10)
	assert.Contains(t, out.Source, "pg_stat_activity")
	require.NotEmpty(t, out.Queries)
	assert.Equal(t, "pg_stat_activity", out.Queries[0].Source)
}

func TestListFallsBackWhenStatementsFail(t *testing.T) {
	r := &slowQueryRepo{
		querier: fakePG{
			hasExt:   true,
			stmtErr:  assert.AnError,
			activity: []dto.SlowQuery{{Query: "SELECT long", Source: "pg_stat_activity"}},
		},
		now: time.Now,
	}
	out := r.List(context.Background(), 5)
	assert.Contains(t, out.Source, "pg_stat_activity")
}

func TestListInMemoryWhenActivityFails(t *testing.T) {
	r := &slowQueryRepo{
		querier: fakePG{activityErr: assert.AnError},
		now:     time.Now,
	}
	out := r.List(context.Background(), 5)
	assert.Equal(t, "in_memory", out.Source)
	assert.True(t, out.Available)
}

func TestMapStatementRowsAndMerge(t *testing.T) {
	rows := mapStatementRows([]statementRow{
		{Query: "  SELECT id FROM users  ", Calls: 3, MeanTimeMs: 12.345, TotalTimeMs: 36, MaxTimeMs: 20, Rows: 2},
		{Query: "   ", Calls: 1},
	}, "pg_stat_statements", "2026-09-12T08:00:00Z")
	require.Len(t, rows, 1)
	assert.Equal(t, "SELECT id FROM users", rows[0].Query)
	assert.Equal(t, 12.35, rows[0].MeanTimeMs)

	merged := mergeSlow(rows, []dto.SlowQuery{
		{Query: "SELECT id FROM users", Source: "pg_stat_statements"},
		{Query: "SELECT 1", Source: "in_memory"},
	}, 10)
	require.Len(t, merged, 2)
	assert.Equal(t, "in_memory", merged[1].Source)
	assert.Len(t, mergeSlow(rows, nil, 1), 1)
}
