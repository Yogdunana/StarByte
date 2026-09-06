package audit

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// FieldChange 单字段 before/after 变更。
type FieldChange struct {
	Path   string `json:"path"`
	Before any    `json:"before"`
	After  any    `json:"after"`
}

// DiffJSON 比较两段 JSON，返回字段级 diff 的 JSON 数组（已脱敏输入应先走 DesensitizeJSON）。
func DiffJSON(before, after string) string {
	if strings.TrimSpace(before) == "" && strings.TrimSpace(after) == "" {
		return ""
	}
	var b, a any
	if strings.TrimSpace(before) != "" {
		if err := json.Unmarshal([]byte(before), &b); err != nil {
			b = before
		}
	}
	if strings.TrimSpace(after) != "" {
		if err := json.Unmarshal([]byte(after), &a); err != nil {
			a = after
		}
	}
	changes := diffValue("", b, a)
	if len(changes) == 0 {
		return "[]"
	}
	out, err := json.Marshal(changes)
	if err != nil {
		return ""
	}
	return string(out)
}

func diffValue(path string, before, after any) []FieldChange {
	if reflect.DeepEqual(before, after) {
		return nil
	}
	bMap, bOK := asMap(before)
	aMap, aOK := asMap(after)
	if bOK && aOK {
		return diffMaps(path, bMap, aMap)
	}
	return []FieldChange{{Path: displayPath(path), Before: before, After: after}}
}

func diffMaps(prefix string, before, after map[string]any) []FieldChange {
	keys := make(map[string]struct{}, len(before)+len(after))
	for k := range before {
		keys[k] = struct{}{}
	}
	for k := range after {
		keys[k] = struct{}{}
	}
	ordered := make([]string, 0, len(keys))
	for k := range keys {
		ordered = append(ordered, k)
	}
	sort.Strings(ordered)
	var out []FieldChange
	for _, k := range ordered {
		child := k
		if prefix != "" {
			child = prefix + "." + k
		}
		out = append(out, diffValue(child, before[k], after[k])...)
	}
	return out
}

func asMap(v any) (map[string]any, bool) {
	if v == nil {
		return nil, false
	}
	m, ok := v.(map[string]any)
	return m, ok
}

func displayPath(path string) string {
	if path == "" {
		return "$"
	}
	return path
}

// ParseFieldChanges 解析已存储的 diff_json；非法输入返回空切片。
func ParseFieldChanges(raw string) []FieldChange {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "[]" {
		return nil
	}
	var changes []FieldChange
	if err := json.Unmarshal([]byte(raw), &changes); err != nil {
		return nil
	}
	return changes
}

// CompactJSON 将任意值序列化为脱敏后的 JSON 字符串。
func CompactJSON(v any, maxLen int) string {
	if v == nil {
		return ""
	}
	raw, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	s := DesensitizeJSON(string(raw))
	if maxLen > 0 && len(s) > maxLen {
		return s[:maxLen] + "...[truncated]"
	}
	return s
}
