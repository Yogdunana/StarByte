package search

import (
	"fmt"
	"regexp"
	"strings"
)

var identRe = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

func quoteIdent(name string) (string, error) {
	if !identRe.MatchString(name) {
		return "", fmt.Errorf("%w: %s", ErrInvalidIdent, name)
	}
	return `"` + name + `"`, nil
}

func (s Schema) field(name string) (Field, bool) {
	for _, f := range s.Fields {
		if f.Name == name {
			return f, true
		}
	}
	return Field{}, false
}

func (s Schema) colRef(f Field) (string, error) {
	col, err := quoteIdent(f.Column)
	if err != nil {
		return "", err
	}
	return `t.` + col, nil
}

func (s Schema) tableRef() (string, error) {
	tbl, err := quoteIdent(s.Table)
	if err != nil {
		return "", err
	}
	return tbl + ` AS t`, nil
}

func operatorsFor(f Field) []string {
	base := []string{OpIsNull, OpNotNull}
	switch f.Kind {
	case KindNumber, KindTime:
		return append([]string{OpEq, OpNe, OpGt, OpGte, OpLt, OpLte, OpIn, OpBetween}, base...)
	case KindBool:
		return append([]string{OpEq, OpNe}, base...)
	default:
		return append([]string{OpEq, OpNe, OpLike, OpPrefix, OpIn}, base...)
	}
}

func (s Schema) OperatorsFor(f Field) []string {
	return operatorsFor(f)
}

func allowedOp(f Field, op string) bool {
	op = strings.ToLower(strings.TrimSpace(op))
	for _, a := range operatorsFor(f) {
		if a == op {
			return true
		}
	}
	return false
}

// ApplyDataScope ANDs a data-range predicate onto ExtraWhere.
// Empty where is a no-op (super-admin / all). Tables without ScopeColumn
// fail closed unless isSelf and SelfSQL can express "only me".
// A "1 = 0" predicate is deny-all unless isSelf is true.
func (s Schema) ApplyDataScope(where string, args []any, userID any, isSelf bool) Schema {
	where = strings.TrimSpace(where)
	if where == "" {
		return s
	}
	var extraArgs []any
	switch {
	case isSelf && strings.TrimSpace(s.SelfSQL) != "":
		where = s.SelfSQL
		n := strings.Count(s.SelfSQL, "?")
		extraArgs = make([]any, n)
		for i := range extraArgs {
			extraArgs[i] = userID
		}
	case s.ScopeColumn == "":
		where, extraArgs = "1 = 0", nil
	default:
		col, err := quoteIdent(s.ScopeColumn)
		if err != nil {
			where, extraArgs = "1 = 0", nil
			break
		}
		where = strings.ReplaceAll(where, "department_id", "t."+col)
		extraArgs = append([]any{}, args...)
	}
	out := s
	if out.ExtraWhere != "" {
		out.ExtraWhere = "(" + out.ExtraWhere + ") AND (" + where + ")"
	} else {
		out.ExtraWhere = where
	}
	out.ExtraArgs = append(append([]any{}, s.ExtraArgs...), extraArgs...)
	return out
}

func (s Schema) PublicFields() []map[string]any {
	out := make([]map[string]any, 0, len(s.Fields))
	for _, f := range s.Fields {
		out = append(out, map[string]any{
			"name":       f.Name,
			"label":      f.Label,
			"type":       f.Kind,
			"searchable": f.Searchable,
			"filterable": f.Filterable,
			"sortable":   f.Sortable,
			"agg":        f.Agg,
			"operators":  s.OperatorsFor(f),
		})
	}
	return out
}
