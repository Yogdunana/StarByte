package search

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"
)

// Engine runs compiled search statements against PostgreSQL.
type Engine struct{}

func NewEngine() *Engine { return &Engine{} }

func (e *Engine) Search(ctx context.Context, db *gorm.DB, schema Schema, q Query) (*Result, error) {
	start := time.Now()
	stmt, err := Compile(schema, q)
	if err != nil {
		return nil, err
	}
	var total int64
	if err := db.WithContext(ctx).Raw(stmt.CountSQL, stmt.CountArgs...).Scan(&total).Error; err != nil {
		return nil, fmt.Errorf("search count: %w", err)
	}
	rows, err := scanMaps(ctx, db, stmt.SQL, stmt.Args)
	if err != nil {
		return nil, fmt.Errorf("search query: %w", err)
	}
	hasMore := len(rows) > stmt.PageSize
	if hasMore {
		rows = rows[:stmt.PageSize]
	}
	next := ""
	if hasMore && len(rows) > 0 {
		vals, id := cursorValues(rows[len(rows)-1], stmt.Sorts)
		next, _ = encodeCursor(vals, id)
	}
	aggs, err := runAggs(ctx, db, stmt.Aggs)
	if err != nil {
		return nil, err
	}
	return &Result{
		List: rows, Total: total, Page: stmt.Page, PageSize: stmt.PageSize,
		NextCursor: next, HasMore: hasMore, Aggregations: aggs,
		ElapsedMs: time.Since(start).Milliseconds(),
	}, nil
}

func scanMaps(ctx context.Context, db *gorm.DB, q string, args []any) ([]map[string]any, error) {
	rows, err := db.WithContext(ctx).Raw(q, args...).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	out := make([]map[string]any, 0, 16)
	for rows.Next() {
		raw := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range raw {
			ptrs[i] = &raw[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		m := make(map[string]any, len(cols))
		for i, c := range cols {
			m[c] = normalize(raw[i])
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func runAggs(ctx context.Context, db *gorm.DB, stmts []AggStatement) (map[string][]AggRow, error) {
	if len(stmts) == 0 {
		return nil, nil
	}
	out := make(map[string][]AggRow, len(stmts))
	for _, s := range stmts {
		rows, err := scanMaps(ctx, db, s.SQL, s.Args)
		if err != nil {
			return nil, fmt.Errorf("search agg %s: %w", s.Name, err)
		}
		list := make([]AggRow, 0, len(rows))
		for _, r := range rows {
			list = append(list, AggRow{
				Key:   r["key"],
				Row:   r["row"],
				Col:   r["col"],
				Value: asFloat(r["value"]),
			})
		}
		out[s.Name] = list
	}
	return out, nil
}

func normalize(v any) any {
	switch t := v.(type) {
	case nil:
		return nil
	case []byte:
		return string(t)
	case time.Time:
		return t.UTC().Format(time.RFC3339)
	case sql.NullTime:
		if !t.Valid {
			return nil
		}
		return t.Time.UTC().Format(time.RFC3339)
	case sql.NullString:
		if !t.Valid {
			return nil
		}
		return t.String
	case sql.NullInt64:
		if !t.Valid {
			return nil
		}
		return t.Int64
	case sql.NullFloat64:
		if !t.Valid {
			return nil
		}
		return t.Float64
	case sql.NullBool:
		if !t.Valid {
			return nil
		}
		return t.Bool
	default:
		return v
	}
}

func asFloat(v any) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case float32:
		return float64(t)
	case int64:
		return float64(t)
	case int32:
		return float64(t)
	case int:
		return float64(t)
	case []byte:
		f, _ := strconv.ParseFloat(string(t), 64)
		return f
	case string:
		f, _ := strconv.ParseFloat(t, 64)
		return f
	default:
		return 0
	}
}
