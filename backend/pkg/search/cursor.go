package search

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

type cursorPayload struct {
	Values []any  `json:"v"`
	ID     string `json:"id"`
}

func encodeCursor(values []any, id string) (string, error) {
	raw, err := json.Marshal(cursorPayload{Values: values, ID: id})
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func decodeCursor(s string) (*cursorPayload, error) {
	raw, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(s))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidCursor, err)
	}
	var p cursorPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidCursor, err)
	}
	if p.ID == "" {
		return nil, fmt.Errorf("%w: missing id", ErrInvalidCursor)
	}
	return &p, nil
}

func compileCursor(schema Schema, sorts []Sort, cur *cursorPayload, args *[]any) (string, error) {
	if cur == nil {
		return "", nil
	}
	if len(cur.Values) != len(sorts) {
		return "", fmt.Errorf("%w: sort arity", ErrInvalidCursor)
	}
	parts := make([]string, 0, len(sorts))
	for i, sort := range sorts {
		f, ok := schema.field(sort.Field)
		if !ok {
			return "", fmt.Errorf("%w: %s", ErrUnknownField, sort.Field)
		}
		col, err := schema.colRef(f)
		if err != nil {
			return "", err
		}
		cmp := "<"
		if !sort.Desc {
			cmp = ">"
		}
		conds := make([]string, 0, i+1)
		for j := 0; j < i; j++ {
			fj, ok := schema.field(sorts[j].Field)
			if !ok {
				return "", fmt.Errorf("%w: %s", ErrUnknownField, sorts[j].Field)
			}
			cj, err := schema.colRef(fj)
			if err != nil {
				return "", err
			}
			conds = append(conds, cj+" = ?")
			*args = append(*args, cur.Values[j])
		}
		conds = append(conds, col+" "+cmp+" ?")
		*args = append(*args, cur.Values[i])
		parts = append(parts, "("+strings.Join(conds, " AND ")+")")
	}
	return strings.Join(parts, " OR "), nil
}

func cursorValues(row map[string]any, sorts []Sort) ([]any, string) {
	vals := make([]any, 0, len(sorts))
	for _, s := range sorts {
		vals = append(vals, row[s.Field])
	}
	id, _ := row["id"].(string)
	if id == "" {
		if v, ok := row["id"]; ok && v != nil {
			id = fmt.Sprint(v)
		}
	}
	return vals, id
}
