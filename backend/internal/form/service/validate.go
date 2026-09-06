package service

import (
	"fmt"
	"math"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/Yogdunana/StarByte/backend/internal/form/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

var namePattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]{0,63}$`)

func validateSchema(fields []model.FormField) error {
	if len(fields) > 80 {
		return response.NewError(response.CodeFormInvalidSchema, "字段数量不能超过 80")
	}
	seen := make(map[string]struct{}, len(fields))
	for i, f := range fields {
		name := strings.TrimSpace(f.Name)
		label := strings.TrimSpace(f.Label)
		if !namePattern.MatchString(name) {
			return response.NewError(response.CodeFormInvalidSchema, fmt.Sprintf("第 %d 个字段 name 不合法", i+1))
		}
		if label == "" || utf8.RuneCountInString(label) > 50 {
			return response.NewError(response.CodeFormInvalidSchema, fmt.Sprintf("字段 %s 的标签不合法", name))
		}
		if _, ok := model.AllowedFieldTypes[f.Type]; !ok {
			return response.NewError(response.CodeFormInvalidSchema, fmt.Sprintf("字段 %s 类型不支持", name))
		}
		if _, dup := seen[name]; dup {
			return response.NewError(response.CodeFormInvalidSchema, fmt.Sprintf("字段名重复: %s", name))
		}
		seen[name] = struct{}{}
		if f.Validation != nil && strings.TrimSpace(f.Validation.Pattern) != "" {
			if _, err := regexp.Compile(f.Validation.Pattern); err != nil {
				return response.NewError(response.CodeFormInvalidSchema, fmt.Sprintf("字段 %s 的 pattern 不是合法正则", name))
			}
		}
		if needsOptions(f.Type) && len(f.Options) == 0 {
			return response.NewError(response.CodeFormInvalidSchema, fmt.Sprintf("字段 %s 需要 options", name))
		}
		if f.VisibleWhen != nil {
			if strings.TrimSpace(f.VisibleWhen.Field) == "" {
				return response.NewError(response.CodeFormInvalidSchema, fmt.Sprintf("字段 %s 的 visible_when.field 为空", name))
			}
			switch f.VisibleWhen.Operator {
			case "==", "!=", ">", "<", "in":
			default:
				return response.NewError(response.CodeFormInvalidSchema, fmt.Sprintf("字段 %s 的 visible_when.operator 不支持", name))
			}
		}
	}
	return nil
}

func needsOptions(t string) bool {
	return t == "select" || t == "radio" || t == "checkbox" || t == "cascader"
}

func isVisible(f model.FormField, values map[string]interface{}) bool {
	cond := f.VisibleWhen
	if cond == nil {
		return true
	}
	actual := values[cond.Field]
	switch cond.Operator {
	case "==":
		return valuesEqual(actual, cond.Value)
	case "!=":
		return !valuesEqual(actual, cond.Value)
	case ">":
		return compareNum(actual, cond.Value) > 0
	case "<":
		return compareNum(actual, cond.Value) < 0
	case "in":
		return valueIn(actual, cond.Value)
	default:
		return true
	}
}

func valuesEqual(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if fa, ok := toFloat(a); ok {
		if fb, ok2 := toFloat(b); ok2 {
			return fa == fb
		}
	}
	return fmt.Sprint(a) == fmt.Sprint(b)
}

func compareNum(a, b interface{}) int {
	fa, ok1 := toFloat(a)
	fb, ok2 := toFloat(b)
	if !ok1 || !ok2 {
		return 0
	}
	if fa > fb {
		return 1
	}
	if fa < fb {
		return -1
	}
	return 0
}

func valueIn(actual, expected interface{}) bool {
	rv := reflect.ValueOf(expected)
	if !rv.IsValid() {
		return false
	}
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return valuesEqual(actual, expected)
	}
	for i := 0; i < rv.Len(); i++ {
		if valuesEqual(actual, rv.Index(i).Interface()) {
			return true
		}
	}
	return false
}

func toFloat(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case jsonNumber:
		f, err := n.Float64()
		return f, err == nil
	case string:
		f, err := strconv.ParseFloat(n, 64)
		return f, err == nil
	default:
		return 0, false
	}
}

type jsonNumber interface{ Float64() (float64, error) }

func fieldMessage(f model.FormField, fallback string) string {
	if f.Validation != nil && f.Validation.Message != "" {
		return f.Validation.Message
	}
	return fallback
}

func isEmpty(v interface{}) bool {
	if v == nil {
		return true
	}
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t) == ""
	case []interface{}:
		return len(t) == 0
	case []string:
		return len(t) == 0
	default:
		return false
	}
}

func asString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprint(v)
}

func validateValue(f model.FormField, v interface{}) error {
	if isEmpty(v) {
		if f.Required {
			return response.NewError(response.CodeFormFieldRequired, fieldMessage(f, f.Label+" 为必填项"))
		}
		return nil
	}
	switch f.Type {
	case "text", "textarea":
		s := asString(v)
		if f.Validation != nil {
			n := utf8.RuneCountInString(s)
			if f.Validation.MinLength != nil && n < *f.Validation.MinLength {
				return response.NewError(response.CodeFormFieldInvalid, fieldMessage(f, f.Label+" 长度不足"))
			}
			if f.Validation.MaxLength != nil && n > *f.Validation.MaxLength {
				return response.NewError(response.CodeFormFieldInvalid, fieldMessage(f, f.Label+" 超出长度"))
			}
			if f.Validation.Pattern != "" {
				re, err := regexp.Compile(f.Validation.Pattern)
				if err == nil && !re.MatchString(s) {
					return response.NewError(response.CodeFormFieldInvalid, fieldMessage(f, f.Label+" 格式不正确"))
				}
			}
		}
	case "number", "rating":
		n, ok := toFloat(v)
		if !ok || math.IsNaN(n) {
			return response.NewError(response.CodeFormFieldInvalid, fieldMessage(f, f.Label+" 必须是数字"))
		}
		if f.Validation != nil {
			if f.Validation.MinValue != nil && n < *f.Validation.MinValue {
				return response.NewError(response.CodeFormFieldInvalid, fieldMessage(f, f.Label+" 小于最小值"))
			}
			if f.Validation.MaxValue != nil && n > *f.Validation.MaxValue {
				return response.NewError(response.CodeFormFieldInvalid, fieldMessage(f, f.Label+" 大于最大值"))
			}
		}
	}
	return nil
}
