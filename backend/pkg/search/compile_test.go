package search

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenize_CJKBigrams(t *testing.T) {
	got := Tokenize("招新系统")
	assert.Contains(t, got, "招")
	assert.Contains(t, got, "新")
	assert.Contains(t, got, "招新")
	assert.Contains(t, got, "系统")
}

func TestTokenize_LatinAndMixed(t *testing.T) {
	got := Tokenize("Hello 招新-world")
	assert.Contains(t, got, "hello")
	assert.Contains(t, got, "world")
	assert.Contains(t, got, "招新")
	q := ToTSQuery("Hello 招新")
	assert.Contains(t, q, "hello")
	assert.Contains(t, q, " & ")
	assert.True(t, tokenSafe("招新"))
	assert.False(t, tokenSafe("a&b"))
}

func TestTokenize_Empty(t *testing.T) {
	assert.Empty(t, Tokenize("   "))
	assert.Empty(t, ToTSQuery("!!!"))
}

func demoSchema() Schema {
	return Schema{
		Code: "tasks", Name: "任务", Table: "tasks", IDColumn: "id",
		FTSExpr:  "starbyte_cjk_tokens(coalesce(t.title,'') || ' ' || coalesce(t.description,''))",
		Headline: "coalesce(t.title,'') || ' ' || coalesce(t.description,'')",
		Fields: []Field{
			{Name: "id", Column: "id", Kind: KindString, Label: "ID", Sortable: true, Filterable: true},
			{Name: "title", Column: "title", Kind: KindString, Label: "标题", Searchable: true, Filterable: true, Sortable: true},
			{Name: "status", Column: "status", Kind: KindNumber, Label: "状态", Filterable: true, Sortable: true, Agg: true},
			{Name: "progress", Column: "progress", Kind: KindNumber, Label: "进度", Filterable: true, Agg: true},
			{Name: "created_at", Column: "created_at", Kind: KindTime, Label: "创建时间", Filterable: true, Sortable: true, Agg: true},
		},
	}
}

func TestCompile_RejectsInjectionField(t *testing.T) {
	_, err := Compile(demoSchema(), Query{
		Filters: &Group{Logic: "and", Conditions: []Condition{
			{Field: "title;drop", Operator: "eq", Value: "x"},
		}},
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnknownField)
}

func TestCompile_NestedAndOrAndLike(t *testing.T) {
	stmt, err := Compile(demoSchema(), Query{
		Keyword: "招新",
		Filters: &Group{
			Logic: "and",
			Conditions: []Condition{
				{Field: "status", Operator: "in", Value: []any{0, 1}},
				{Field: "title", Operator: "like", Value: "100%_off"},
			},
			Groups: []Group{{
				Logic: "or",
				Conditions: []Condition{
					{Field: "progress", Operator: "gte", Value: 50},
					{Field: "status", Operator: "eq", Value: 3},
				},
			}},
		},
		Sorts:    []Sort{{Field: "created_at", Desc: true}},
		Page:     1,
		PageSize: 10,
		Aggregations: []AggRequest{
			{Name: "by_status", Field: "status", Fn: "count"},
			{Name: "by_day", Field: "created_at", Fn: "count", Interval: "day"},
			{Name: "cross", Fn: "count", Row: "status", Col: "progress"},
		},
	})
	require.NoError(t, err)
	assert.Contains(t, stmt.SQL, "to_tsvector")
	assert.Contains(t, stmt.SQL, "ILIKE")
	assert.Contains(t, stmt.SQL, "ESCAPE")
	assert.Contains(t, stmt.SQL, " IN (")
	assert.Contains(t, stmt.SQL, " OR ")
	assert.NotContains(t, stmt.SQL, "100%_off")
	assert.Contains(t, stmt.Args, `%100\%\_off%`)
	assert.Len(t, stmt.Aggs, 3)
	assert.Contains(t, stmt.Aggs[1].SQL, "date_trunc('day'")
	assert.Contains(t, stmt.Aggs[2].SQL, "GROUP BY 1, 2")
	assert.True(t, strings.HasPrefix(stmt.SQL, "SELECT "))
}

func TestCompile_DeepOffsetRejected(t *testing.T) {
	_, err := Compile(demoSchema(), Query{Page: 200, PageSize: 100})
	require.ErrorIs(t, err, ErrDeepPagination)
}

func TestCompile_CursorNotInCount(t *testing.T) {
	cur, err := encodeCursor([]any{"2026-01-01T00:00:00Z", "abc"}, "abc")
	require.NoError(t, err)
	stmt, err := Compile(demoSchema(), Query{
		Cursor: cur, PageSize: 5,
		Sorts: []Sort{{Field: "created_at", Desc: true}, {Field: "id", Desc: true}},
	})
	require.NoError(t, err)
	assert.Contains(t, stmt.SQL, " OR ")
	assert.NotContains(t, stmt.CountSQL, " OR ")
}

func TestQuoteIdent(t *testing.T) {
	_, err := quoteIdent("tasks;drop")
	assert.Error(t, err)
	got, err := quoteIdent("created_at")
	require.NoError(t, err)
	assert.Equal(t, `"created_at"`, got)
}

func TestInvalidAggAndOp(t *testing.T) {
	_, err := Compile(demoSchema(), Query{Aggregations: []AggRequest{{Fn: "median", Field: "status"}}})
	require.ErrorIs(t, err, ErrInvalidAgg)
	_, err = Compile(demoSchema(), Query{Filters: &Group{Conditions: []Condition{
		{Field: "title", Operator: "gte", Value: 1},
	}}})
	require.ErrorIs(t, err, ErrInvalidOp)
}
