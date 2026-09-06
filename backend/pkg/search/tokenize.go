package search

import (
	"strings"
	"unicode"
)

func isCJK(r rune) bool {
	return r >= 0x4E00 && r <= 0x9FFF
}

func isWordChar(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_'
}

// Tokenize splits Latin words and overlapping CJK unigrams/bigrams, matching
// the PostgreSQL function starbyte_cjk_tokens.
func Tokenize(s string) []string {
	var out []string
	var buf strings.Builder
	var lastCJK rune
	flushLatin := func() {
		if buf.Len() == 0 {
			return
		}
		out = append(out, strings.ToLower(buf.String()))
		buf.Reset()
	}
	for _, r := range s {
		switch {
		case isCJK(r):
			flushLatin()
			out = append(out, string(r))
			if lastCJK != 0 {
				out = append(out, string(lastCJK)+string(r))
			}
			lastCJK = r
		case isWordChar(r):
			lastCJK = 0
			buf.WriteRune(unicode.ToLower(r))
		default:
			flushLatin()
			lastCJK = 0
		}
	}
	flushLatin()
	return dedupe(out)
}

func dedupe(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, t := range in {
		if t == "" {
			continue
		}
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	return out
}

func tokenSafe(t string) bool {
	for _, r := range t {
		if isCJK(r) || isWordChar(r) {
			continue
		}
		return false
	}
	return t != ""
}

// ToTSQuery builds a simple AND tsquery from keyword tokens.
func ToTSQuery(keyword string) string {
	toks := Tokenize(keyword)
	parts := make([]string, 0, len(toks))
	for _, t := range toks {
		if !tokenSafe(t) {
			continue
		}
		parts = append(parts, t)
	}
	return strings.Join(parts, " & ")
}
