package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompile_NullBetweenPrefixAndEmptyIn(t *testing.T) {
	stmt, err := Compile(demoSchema(), Query{
		Filters: &Group{Logic: "or", Conditions: []Condition{
			{Field: "title", Operator: "is_null"},
			{Field: "title", Operator: "not_null"},
			{Field: "title", Operator: "prefix", Value: "ab_c"},
			{Field: "progress", Operator: "between", Value: []any{1, 9}},
			{Field: "status", Operator: "in", Value: []any{}},
			{Field: "status", Operator: "ne", Value: 2},
		}},
		PageSize: 5,
	})
	require.NoError(t, err)
	assert.Contains(t, stmt.SQL, "IS NULL")
	assert.Contains(t, stmt.SQL, "IS NOT NULL")
	assert.Contains(t, stmt.SQL, "FALSE")
	assert.Contains(t, stmt.SQL, "BETWEEN")
	assert.Contains(t, stmt.Args, `ab\_c%`)
}

func TestCompile_TooManyFiltersAndDepth(t *testing.T) {
	conds := make([]Condition, MaxFilters+1)
	for i := range conds {
		conds[i] = Condition{Field: "status", Operator: "eq", Value: i}
	}
	_, err := Compile(demoSchema(), Query{Filters: &Group{Conditions: conds}})
	require.ErrorIs(t, err, ErrInvalidQuery)

	g := &Group{Conditions: []Condition{{Field: "status", Operator: "eq", Value: 1}}}
	cur := g
	for i := 0; i < MaxGroupDepth+1; i++ {
		n := Group{Conditions: []Condition{{Field: "status", Operator: "eq", Value: 1}}}
		cur.Groups = []Group{n}
		cur = &cur.Groups[0]
	}
	_, err = Compile(demoSchema(), Query{Filters: g})
	require.ErrorIs(t, err, ErrInvalidQuery)
}

func TestCompile_BadTableAndSortAndBetween(t *testing.T) {
	bad := demoSchema()
	bad.Table = "Tasks"
	_, err := Compile(bad, Query{})
	require.ErrorIs(t, err, ErrInvalidIdent)

	_, err = Compile(demoSchema(), Query{Sorts: []Sort{{Field: "nope"}}})
	require.ErrorIs(t, err, ErrUnknownField)

	_, err = Compile(demoSchema(), Query{Filters: &Group{Conditions: []Condition{
		{Field: "progress", Operator: "between", Value: 1},
	}}})
	require.ErrorIs(t, err, ErrInvalidQuery)

	_, err = Compile(demoSchema(), Query{Cursor: "not-base64"})
	require.ErrorIs(t, err, ErrInvalidCursor)
}

func TestCompile_ScalarAndTimeAgg(t *testing.T) {
	stmt, err := Compile(demoSchema(), Query{Aggregations: []AggRequest{
		{Name: "all", Fn: "count"},
		{Name: "avg_p", Field: "progress", Fn: "avg"},
		{Name: "sum_p", Field: "progress", Fn: "sum"},
		{Name: "min_p", Field: "progress", Fn: "min"},
		{Name: "max_p", Field: "progress", Fn: "max"},
	}})
	require.NoError(t, err)
	assert.Contains(t, stmt.Aggs[0].SQL, "COUNT(*)")
	assert.Equal(t, "scalar", stmt.Aggs[0].Kind)
	assert.Contains(t, stmt.Aggs[1].SQL, "AVG(")
	assert.NotContains(t, stmt.Aggs[1].SQL, "GROUP BY")
	assert.Equal(t, "scalar", stmt.Aggs[1].Kind)
	assert.NotContains(t, stmt.Aggs[2].SQL, "GROUP BY")
	assert.NotContains(t, stmt.Aggs[3].SQL, "GROUP BY")
	assert.NotContains(t, stmt.Aggs[4].SQL, "GROUP BY")

	_, err = Compile(demoSchema(), Query{Aggregations: []AggRequest{
		{Field: "created_at", Fn: "count", Interval: "hour"},
	}})
	require.ErrorIs(t, err, ErrInvalidAgg)
}

func TestCompile_PageClamp(t *testing.T) {
	stmt, err := Compile(demoSchema(), Query{Page: 0, PageSize: 500})
	require.NoError(t, err)
	assert.Equal(t, DefaultPage, stmt.Page)
	assert.Equal(t, MaxPageSize, stmt.PageSize)
}

func TestCountFiltersNil(t *testing.T) {
	assert.Equal(t, 0, countFilters(nil))
}

func TestLikeNeedsString(t *testing.T) {
	_, err := Compile(demoSchema(), Query{Filters: &Group{Conditions: []Condition{
		{Field: "title", Operator: "like", Value: 3},
	}}})
	require.ErrorIs(t, err, ErrInvalidQuery)
}

func TestSkipEmptyFilter(t *testing.T) {
	stmt, err := Compile(demoSchema(), Query{Filters: &Group{Conditions: []Condition{
		{Field: "id", Operator: "eq", Value: ""},
		{Field: "status", Operator: "eq", Value: 1},
	}}})
	require.NoError(t, err)
	assert.NotContains(t, stmt.SQL, `t."id" =`)
	assert.Contains(t, stmt.SQL, `t."status"`)
}

func TestInTooLong(t *testing.T) {
	vals := make([]any, MaxINValues+1)
	for i := range vals {
		vals[i] = i
	}
	_, err := Compile(demoSchema(), Query{Filters: &Group{Conditions: []Condition{
		{Field: "status", Operator: "in", Value: vals},
	}}})
	require.ErrorIs(t, err, ErrInvalidQuery)
}

func TestCompile_ExtraArgsAndCountGroup(t *testing.T) {
	sch := demoSchema()
	sch.ExtraWhere = `t."department_id" = ?`
	sch.ExtraArgs = []any{"dept-1"}
	stmt, err := Compile(sch, Query{
		Aggregations: []AggRequest{{Name: "by_status", Field: "status", Fn: "count"}},
	})
	require.NoError(t, err)
	assert.Contains(t, stmt.CountSQL, `t."department_id" = ?`)
	assert.Contains(t, stmt.CountArgs, "dept-1")
	require.Len(t, stmt.Aggs, 1)
	assert.Contains(t, stmt.Aggs[0].SQL, "GROUP BY")
	assert.Equal(t, "group", stmt.Aggs[0].Kind)
}

func TestApplyDataScope(t *testing.T) {
	base := demoSchema()
	base.ScopeColumn = "department_id"
	base.SelfSQL = `t."creator_id" = ? OR t."assignee_id" = ?`
	base.ExtraWhere = "deleted_at IS NULL"

	open := base.ApplyDataScope("", nil, "me")
	assert.Equal(t, "deleted_at IS NULL", open.ExtraWhere)

	dept := base.ApplyDataScope("department_id = ?", []any{"d1"}, "me")
	assert.Contains(t, dept.ExtraWhere, `t."department_id" = ?`)
	assert.Equal(t, []any{"d1"}, dept.ExtraArgs)

	self := base.ApplyDataScope("1 = 0", nil, "me")
	assert.Contains(t, self.ExtraWhere, `t."creator_id" = ?`)
	assert.Equal(t, []any{"me", "me"}, self.ExtraArgs)

	audit := Schema{Code: "audit_logs", Table: "audit_logs"}
	closed := audit.ApplyDataScope("department_id = ?", []any{"d1"}, "me")
	assert.Equal(t, "1 = 0", closed.ExtraWhere)
	assert.Empty(t, closed.ExtraArgs)
}
