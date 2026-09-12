// Package locale contains translations for application-generated text only.
package locale

import (
	"context"
	"embed"
	"encoding/json"
	"regexp"
	"sort"
	"strings"

	"golang.org/x/text/language"
)

//go:embed catalog/*.json
var catalogFS embed.FS
var catalogs = loadCatalogs()

type contextKey struct{}

func Normalize(raw string) string {
	switch strings.ToLower(strings.Split(raw, "-")[0]) {
	case "en":
		return "en-US"
	case "ru":
		return "ru-RU"
	default:
		return "zh-CN"
	}
}
func FromHeader(header string) string {
	tags, _, err := language.ParseAcceptLanguage(header)
	if err != nil || len(tags) == 0 {
		return "zh-CN"
	}
	_, index, confidence := language.NewMatcher([]language.Tag{language.SimplifiedChinese, language.AmericanEnglish, language.Russian}).Match(tags...)
	if confidence == language.No {
		return "zh-CN"
	}
	return []string{"zh-CN", "en-US", "ru-RU"}[index]
}
func WithContext(ctx context.Context, lang string) context.Context {
	return context.WithValue(ctx, contextKey{}, Normalize(lang))
}
func FromContext(ctx context.Context) string {
	lang, _ := ctx.Value(contextKey{}).(string)
	return Normalize(lang)
}
func Text(lang, source string) string {
	if value, ok := catalogs[Normalize(lang)][source]; ok {
		return value
	}
	return source
}
func loadCatalogs() map[string]map[string]string {
	out := map[string]map[string]string{}
	for _, lang := range []string{"en-US", "ru-RU"} {
		raw, err := catalogFS.ReadFile("catalog/" + lang + ".json")
		if err != nil {
			panic(err)
		}
		values := map[string]string{}
		if err = json.Unmarshal(raw, &values); err != nil {
			panic(err)
		}
		out[lang] = values
	}
	return out
}

type messagePattern struct {
	source     string
	expression *regexp.Regexp
}

var formatToken = regexp.MustCompile(`%[sdv]`)
var messagePatterns = compileMessagePatterns()

func compileMessagePatterns() []messagePattern {
	out := []messagePattern{}
	for source := range catalogs["en-US"] {
		pattern := messagePattern{source: source}
		if formatToken.MatchString(source) {
			parts := formatToken.Split(source, -1)
			for i := range parts {
				parts[i] = regexp.QuoteMeta(parts[i])
			}
			pattern.expression = regexp.MustCompile("(?s)^" + strings.Join(parts, "(.*?)") + "$")
		}
		out = append(out, pattern)
	}
	sort.Slice(out, func(i, j int) bool { return len(out[i].source) > len(out[j].source) })
	return out
}

// Message translates system error text, preserving dynamic IDs and diagnostic details.
func Message(lang, source string) string {
	lang = Normalize(lang)
	if lang == "zh-CN" {
		return source
	}
	if value, ok := catalogs[lang][source]; ok {
		return value
	}
	for _, pattern := range messagePatterns {
		translated := Text(lang, pattern.source)
		if pattern.expression != nil {
			if match := pattern.expression.FindStringSubmatch(source); match != nil {
				index := 0
				return formatToken.ReplaceAllStringFunc(translated, func(_ string) string {
					index++
					if index < len(match) {
						return match[index]
					}
					return ""
				})
			}
		}
		// Prefix/suffix fragments are used by field validation and wrapped diagnostics.
		if strings.HasSuffix(pattern.source, ": ") && strings.HasPrefix(source, pattern.source) {
			return translated + strings.TrimPrefix(source, pattern.source)
		}
		if strings.HasPrefix(pattern.source, " ") && strings.HasSuffix(source, pattern.source) {
			return strings.TrimSuffix(source, pattern.source) + translated
		}
	}
	return source
}
