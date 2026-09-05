package service

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

var mentionRe = regexp.MustCompile(`@([A-Za-z0-9_\-\.]{2,64})`)

func parseMentionNames(content string) []string {
	matches := mentionRe.FindAllStringSubmatch(content, -1)
	seen := map[string]struct{}{}
	out := make([]string, 0, len(matches))
	for _, m := range matches {
		name := strings.TrimSpace(m[1])
		if name == "" {
			continue
		}
		key := strings.ToLower(name)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, name)
	}
	return out
}

func encodeJSONList(ids []string) string {
	if len(ids) == 0 {
		return "[]"
	}
	b, err := json.Marshal(ids)
	if err != nil {
		return "[]"
	}
	return string(b)
}

func decodeJSONList(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "[]" {
		return []string{}
	}
	var out []string
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		parts := strings.Split(raw, ",")
		out = make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				out = append(out, p)
			}
		}
	}
	if out == nil {
		return []string{}
	}
	return out
}

func encodeTags(tags []string) string {
	clean := make([]string, 0, len(tags))
	for _, t := range tags {
		t = strings.TrimSpace(t)
		if t != "" {
			clean = append(clean, t)
		}
	}
	s := encodeJSONList(clean)
	if len(s) > 500 {
		s = s[:500]
	}
	return s
}

func parseUUIDPtr(raw string) *uuid.UUID {
	id, err := uuid.Parse(strings.TrimSpace(raw))
	if err != nil {
		return nil
	}
	return &id
}

func uuidPtrString(id *uuid.UUID) string {
	if id == nil {
		return ""
	}
	return id.String()
}
