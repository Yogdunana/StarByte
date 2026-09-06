package search

import (
	"fmt"
	"strings"
)

func compileAggs(schema Schema, reqs []AggRequest, tbl, where string, whereArgs []any) ([]AggStatement, error) {
	if len(reqs) == 0 {
		return nil, nil
	}
	out := make([]AggStatement, 0, len(reqs))
	for i, r := range reqs {
		name := strings.TrimSpace(r.Name)
		if name == "" {
			name = fmt.Sprintf("agg_%d", i)
		}
		fn := strings.ToLower(strings.TrimSpace(r.Fn))
		if fn == "" {
			fn = FnCount
		}
		if fn != FnCount && fn != FnSum && fn != FnAvg && fn != FnMin && fn != FnMax {
			return nil, fmt.Errorf("%w: fn %s", ErrInvalidAgg, r.Fn)
		}
		if r.Row != "" && r.Col != "" {
			stmt, err := crossAgg(schema, name, fn, r, tbl, where, whereArgs)
			if err != nil {
				return nil, err
			}
			out = append(out, stmt)
			continue
		}
		if r.Interval != "" {
			stmt, err := timeAgg(schema, name, fn, r, tbl, where, whereArgs)
			if err != nil {
				return nil, err
			}
			out = append(out, stmt)
			continue
		}
		stmt, err := groupAgg(schema, name, fn, r, tbl, where, whereArgs)
		if err != nil {
			return nil, err
		}
		out = append(out, stmt)
	}
	return out, nil
}

func aggFnSQL(schema Schema, fn, field string) (string, error) {
	if fn == FnCount && field == "" {
		return "COUNT(*)", nil
	}
	f, ok := schema.field(field)
	if !ok || !f.Agg {
		return "", fmt.Errorf("%w: field %s", ErrInvalidAgg, field)
	}
	if fn != FnCount && f.Kind != KindNumber {
		return "", fmt.Errorf("%w: %s needs a number field", ErrInvalidAgg, fn)
	}
	col, err := schema.colRef(f)
	if err != nil {
		return "", err
	}
	switch fn {
	case FnCount:
		return "COUNT(" + col + ")", nil
	case FnSum:
		return "SUM(" + col + ")", nil
	case FnAvg:
		return "AVG(" + col + ")", nil
	case FnMin:
		return "MIN(" + col + ")", nil
	case FnMax:
		return "MAX(" + col + ")", nil
	}
	return "", fmt.Errorf("%w: fn %s", ErrInvalidAgg, fn)
}

func groupAgg(schema Schema, name, fn string, r AggRequest, tbl, where string, whereArgs []any) (AggStatement, error) {
	expr, err := aggFnSQL(schema, fn, r.Field)
	if err != nil {
		return AggStatement{}, err
	}
	args := append([]any{}, whereArgs...)
	if fn != FnCount || r.Field == "" {
		sql := "SELECT " + expr + " AS value FROM " + tbl + " WHERE " + where
		return AggStatement{Name: name, SQL: sql, Args: args, Kind: "scalar"}, nil
	}
	f, ok := schema.field(r.Field)
	if !ok {
		return AggStatement{}, fmt.Errorf("%w: field %s", ErrInvalidAgg, r.Field)
	}
	col, err := schema.colRef(f)
	if err != nil {
		return AggStatement{}, err
	}
	sql := "SELECT " + col + " AS key, " + expr + " AS value FROM " + tbl + " WHERE " + where + " GROUP BY " + col + " ORDER BY " + col
	return AggStatement{Name: name, SQL: sql, Args: args, Kind: "group"}, nil
}

func timeAgg(schema Schema, name, fn string, r AggRequest, tbl, where string, whereArgs []any) (AggStatement, error) {
	iv := strings.ToLower(r.Interval)
	if iv != "day" && iv != "week" && iv != "month" && iv != "year" {
		return AggStatement{}, fmt.Errorf("%w: interval %s", ErrInvalidAgg, r.Interval)
	}
	f, ok := schema.field(r.Field)
	if !ok || f.Kind != KindTime {
		return AggStatement{}, fmt.Errorf("%w: time field %s", ErrInvalidAgg, r.Field)
	}
	col, err := schema.colRef(f)
	if err != nil {
		return AggStatement{}, err
	}
	if fn != FnCount {
		return AggStatement{}, fmt.Errorf("%w: time buckets only support count", ErrInvalidAgg)
	}
	bucket := "date_trunc('" + iv + "', " + col + ")"
	sql := "SELECT " + bucket + " AS key, COUNT(*) AS value FROM " + tbl + " WHERE " + where + " GROUP BY 1 ORDER BY 1"
	return AggStatement{Name: name, SQL: sql, Args: append([]any{}, whereArgs...), Kind: "time"}, nil
}

func crossAgg(schema Schema, name, fn string, r AggRequest, tbl, where string, whereArgs []any) (AggStatement, error) {
	rf, ok := schema.field(r.Row)
	if !ok || !rf.Agg {
		return AggStatement{}, fmt.Errorf("%w: row %s", ErrInvalidAgg, r.Row)
	}
	cf, ok := schema.field(r.Col)
	if !ok || !cf.Agg {
		return AggStatement{}, fmt.Errorf("%w: col %s", ErrInvalidAgg, r.Col)
	}
	rc, err := schema.colRef(rf)
	if err != nil {
		return AggStatement{}, err
	}
	cc, err := schema.colRef(cf)
	if err != nil {
		return AggStatement{}, err
	}
	metric := r.Field
	if metric == "" {
		metric = r.Row
	}
	expr, err := aggFnSQL(schema, fn, metric)
	if fn == FnCount {
		expr, err = "COUNT(*)", nil
	}
	if err != nil {
		return AggStatement{}, err
	}
	sql := "SELECT " + rc + " AS row, " + cc + " AS col, " + expr + " AS value FROM " + tbl +
		" WHERE " + where + " GROUP BY 1, 2 ORDER BY 1, 2"
	return AggStatement{Name: name, SQL: sql, Args: append([]any{}, whereArgs...), Kind: "cross"}, nil
}
