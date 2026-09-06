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
	assert.Contains(t, stmt.Aggs[1].SQL, "AVG(")

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
