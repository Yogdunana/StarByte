package search

import (
	"fmt"
	"reflect"
	"strings"
)

func compileGroup(schema Schema, g *Group, args *[]any, depth int) (string, error) {
	if g == nil {
		return "", nil
	}
	if depth > MaxGroupDepth {
		return "", fmt.Errorf("%w: filter nesting too deep", ErrInvalidQuery)
	}
	logic := "AND"
	if strings.EqualFold(strings.TrimSpace(g.Logic), LogicOr) {
		logic = "OR"
	}
	n := len(g.Conditions) + len(g.Groups)
	if n == 0 {
		return "", nil
	}
	parts := make([]string, 0, n)
	for _, c := range g.Conditions {
		p, err := compileCond(schema, c, args)
		if err != nil {
			return "", err
		}
		parts = append(parts, p)
	}
	for i := range g.Groups {
		p, err := compileGroup(schema, &g.Groups[i], args, depth+1)
		if err != nil {
			return "", err
		}
		if p != "" {
			parts = append(parts, "("+p+")")
		}
	}
	if len(parts) == 0 {
		return "", nil
	}
	return strings.Join(parts, " "+logic+" "), nil
}

func compileCond(schema Schema, c Condition, args *[]any) (string, error) {
	f, ok := schema.field(c.Field)
	if !ok || !f.Filterable {
		return "", fmt.Errorf("%w: %s", ErrUnknownField, c.Field)
	}
	op := strings.ToLower(strings.TrimSpace(c.Operator))
	if !allowedOp(f, op) {
		return "", fmt.Errorf("%w: %s on %s", ErrInvalidOp, c.Operator, c.Field)
	}
	col, err := schema.colRef(f)
	if err != nil {
		return "", err
	}
	switch op {
	case OpIsNull:
		return col + " IS NULL", nil
	case OpNotNull:
		return col + " IS NOT NULL", nil
	case OpIn:
		vals, err := asSlice(c.Value)
		if err != nil {
			return "", err
		}
		if len(vals) == 0 {
			return "FALSE", nil
		}
		if len(vals) > MaxINValues {
			return "", fmt.Errorf("%w: in-list too long", ErrInvalidQuery)
		}
		ph := make([]string, len(vals))
		for i, v := range vals {
			ph[i] = "?"
			*args = append(*args, v)
		}
		return col + " IN (" + strings.Join(ph, ",") + ")", nil
	case OpBetween:
		vals, err := asSlice(c.Value)
		if err != nil || len(vals) != 2 {
			return "", fmt.Errorf("%w: between needs [min,max]", ErrInvalidQuery)
		}
		*args = append(*args, vals[0], vals[1])
		return col + " BETWEEN ? AND ?", nil
	case OpLike:
		s, err := asString(c.Value)
		if err != nil {
			return "", err
		}
		*args = append(*args, "%"+escapeLike(s)+"%")
		return col + ` ILIKE ? ESCAPE '\'`, nil
	case OpPrefix:
		s, err := asString(c.Value)
		if err != nil {
			return "", err
		}
		*args = append(*args, escapeLike(s)+"%")
		return col + ` ILIKE ? ESCAPE '\'`, nil
	}
	sym, ok := cmpOp(op)
	if !ok {
		return "", fmt.Errorf("%w: %s", ErrInvalidOp, op)
	}
	*args = append(*args, c.Value)
	return col + " " + sym + " ?", nil
}

func cmpOp(op string) (string, bool) {
	switch op {
	case OpEq:
		return "=", true
	case OpNe:
		return "<>", true
	case OpGt:
		return ">", true
	case OpGte:
		return ">=", true
	case OpLt:
		return "<", true
	case OpLte:
		return "<=", true
	}
	return "", false
}

func escapeLike(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}

func asString(v any) (string, error) {
	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("%w: expected string", ErrInvalidQuery)
	}
	return s, nil
}

func asSlice(v any) ([]any, error) {
	if v == nil {
		return nil, fmt.Errorf("%w: expected array", ErrInvalidQuery)
	}
	if s, ok := v.([]any); ok {
		return s, nil
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return nil, fmt.Errorf("%w: expected array", ErrInvalidQuery)
	}
	out := make([]any, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		out[i] = rv.Index(i).Interface()
	}
	return out, nil
}

func countFilters(g *Group) int {
	if g == nil {
		return 0
	}
	n := len(g.Conditions)
	for i := range g.Groups {
		n += countFilters(&g.Groups[i])
	}
	return n
}
