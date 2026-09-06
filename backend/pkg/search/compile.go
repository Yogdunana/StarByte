package search

import (
	"fmt"
	"strings"
)

func defaultSorts(schema Schema, sorts []Sort) ([]Sort, error) {
	out := make([]Sort, 0, len(sorts)+1)
	seen := map[string]bool{}
	for _, s := range sorts {
		f, ok := schema.field(s.Field)
		if !ok || !f.Sortable {
			return nil, fmt.Errorf("%w: %s", ErrUnknownField, s.Field)
		}
		if seen[s.Field] {
			continue
		}
		seen[s.Field] = true
		out = append(out, Sort{Field: s.Field, Desc: s.Desc})
	}
	idName := schema.IDColumn
	if idName == "" {
		idName = "id"
	}
	if !seen[idName] {
		if f, ok := schema.field(idName); ok && f.Sortable {
			out = append(out, Sort{Field: idName, Desc: true})
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("%w: no sortable fields", ErrInvalidQuery)
	}
	return out, nil
}

func compileWhere(schema Schema, q Query) (string, []any, []Sort, int, int, error) {
	if countFilters(q.Filters) > MaxFilters {
		return "", nil, nil, 0, 0, fmt.Errorf("%w: too many filters", ErrInvalidQuery)
	}
	page := q.Page
	size := q.PageSize
	if page <= 0 {
		page = DefaultPage
	}
	if size <= 0 {
		size = DefaultSize
	}
	if size > MaxPageSize {
		size = MaxPageSize
	}
	if q.Cursor == "" && (page-1)*size > MaxOffset {
		return "", nil, nil, 0, 0, ErrDeepPagination
	}
	sorts, err := defaultSorts(schema, q.Sorts)
	if err != nil {
		return "", nil, nil, 0, 0, err
	}

	var args []any
	var parts []string
	if schema.ExtraWhere != "" {
		parts = append(parts, "("+schema.ExtraWhere+")")
	}
	if ts := ToTSQuery(q.Keyword); ts != "" && schema.FTSExpr != "" {
		parts = append(parts, `to_tsvector('simple', `+schema.FTSExpr+`) @@ to_tsquery('simple', ?)`)
		args = append(args, ts)
	}
	if q.Filters != nil {
		fw, err := compileGroup(schema, q.Filters, &args, 1)
		if err != nil {
			return "", nil, nil, 0, 0, err
		}
		if fw != "" {
			parts = append(parts, "("+fw+")")
		}
	}
	where := "TRUE"
	if len(parts) > 0 {
		where = strings.Join(parts, " AND ")
	}
	return where, args, sorts, page, size, nil
}

func compileListWhere(schema Schema, q Query, sorts []Sort, base string, args []any) (string, []any, error) {
	if q.Cursor == "" {
		return base, args, nil
	}
	cur, err := decodeCursor(q.Cursor)
	if err != nil {
		return "", nil, err
	}
	listArgs := append([]any{}, args...)
	cw, err := compileCursor(schema, sorts, cur, &listArgs)
	if err != nil {
		return "", nil, err
	}
	if cw == "" {
		return base, listArgs, nil
	}
	return base + " AND (" + cw + ")", listArgs, nil
}

func orderSQL(schema Schema, sorts []Sort) (string, error) {
	parts := make([]string, 0, len(sorts))
	for _, s := range sorts {
		f, ok := schema.field(s.Field)
		if !ok {
			return "", fmt.Errorf("%w: %s", ErrUnknownField, s.Field)
		}
		col, err := schema.colRef(f)
		if err != nil {
			return "", err
		}
		dir := "ASC"
		if s.Desc {
			dir = "DESC"
		}
		parts = append(parts, col+" "+dir)
	}
	return strings.Join(parts, ", "), nil
}

func selectSQL(schema Schema, keyword string) (string, error) {
	cols := make([]string, 0, len(schema.Fields)+1)
	for _, f := range schema.Fields {
		col, err := schema.colRef(f)
		if err != nil {
			return "", err
		}
		alias, err := quoteIdent(f.Name)
		if err != nil {
			return "", err
		}
		cols = append(cols, col+" AS "+alias)
	}
	if keyword != "" && schema.Headline != "" && ToTSQuery(keyword) != "" {
		cols = append(cols, `ts_headline('simple', `+schema.Headline+`, to_tsquery('simple', ?), 'MaxFragments=2, MaxWords=16') AS "_headline"`)
	}
	return strings.Join(cols, ", "), nil
}

// Compile builds parameterized list/count/aggregation SQL for a query.
func Compile(schema Schema, q Query) (*Statement, error) {
	if _, err := schema.tableRef(); err != nil {
		return nil, err
	}
	baseWhere, args, sorts, page, size, err := compileWhere(schema, q)
	if err != nil {
		return nil, err
	}
	where, listArgs, err := compileListWhere(schema, q, sorts, baseWhere, args)
	if err != nil {
		return nil, err
	}
	sel, err := selectSQL(schema, q.Keyword)
	if err != nil {
		return nil, err
	}
	ord, err := orderSQL(schema, sorts)
	if err != nil {
		return nil, err
	}
	tbl, _ := schema.tableRef()
	if q.Keyword != "" && schema.Headline != "" && ToTSQuery(q.Keyword) != "" {
		listArgs = append([]any{ToTSQuery(q.Keyword)}, listArgs...)
	}
	offset := 0
	if q.Cursor == "" {
		offset = (page - 1) * size
	}
	sql := "SELECT " + sel + " FROM " + tbl + " WHERE " + where + " ORDER BY " + ord + " LIMIT ? OFFSET ?"
	listArgs = append(listArgs, size+1, offset)

	countSQL := "SELECT COUNT(*) FROM " + tbl + " WHERE " + baseWhere
	aggs, err := compileAggs(schema, q.Aggregations, tbl, baseWhere, args)
	if err != nil {
		return nil, err
	}
	return &Statement{
		SQL: sql, Args: listArgs,
		CountSQL: countSQL, CountArgs: append([]any{}, args...),
		Aggs: aggs, Limit: size + 1, Page: page, PageSize: size, Sorts: sorts,
	}, nil
}
