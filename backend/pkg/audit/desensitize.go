package audit

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"
)

const redacted = "***"

var hideCompletely = map[string]struct{}{
	"password":      {},
	"old_password":  {},
	"new_password":  {},
	"token":         {},
	"secret":        {},
	"access_token":  {},
	"refresh_token": {},
}

var compiledPatterns struct {
	once   sync.Once
	hide   *regexp.Regexp
	phone  *regexp.Regexp
	email  *regexp.Regexp
	idCard *regexp.Regexp
}

func initPatterns() {
	hideKeys := []string{
		"password", "old_password", "new_password",
		"token", "secret", "access_token", "refresh_token",
	}
	compiledPatterns.hide = regexp.MustCompile(
		`(?i)"(` + strings.Join(hideKeys, "|") + `)"\s*:\s*"[^"]*"`,
	)
	compiledPatterns.phone = regexp.MustCompile(`(?i)"phone"\s*:\s*"([^"]*)"`)
	compiledPatterns.email = regexp.MustCompile(`(?i)"email"\s*:\s*"([^"]*)"`)
	compiledPatterns.idCard = regexp.MustCompile(`(?i)"id_card"\s*:\s*"([^"]*)"`)
}

func patterns() {
	compiledPatterns.once.Do(initPatterns)
}

// DesensitizeJSON 按 Issue #5 规则脱敏 JSON 字符串中的敏感字段。
func DesensitizeJSON(body string) string {
	if body == "" {
		return body
	}
	var v any
	if err := json.Unmarshal([]byte(body), &v); err != nil {
		return desensitizeByRegex(body)
	}
	walk(v)
	out, err := json.Marshal(v)
	if err != nil {
		return desensitizeByRegex(body)
	}
	return string(out)
}

func walk(v any) {
	switch node := v.(type) {
	case map[string]any:
		for k, child := range node {
			key := strings.ToLower(k)
			if _, ok := hideCompletely[key]; ok {
				node[k] = redacted
				continue
			}
			if s, ok := child.(string); ok {
				switch key {
				case "phone":
					node[k] = MaskPhone(s)
					continue
				case "email":
					node[k] = MaskEmail(s)
					continue
				case "id_card":
					node[k] = MaskIDCard(s)
					continue
				}
			}
			walk(child)
		}
	case []any:
		for _, child := range node {
			walk(child)
		}
	}
}

func desensitizeByRegex(body string) string {
	patterns()
	result := compiledPatterns.hide.ReplaceAllString(body, `"$1":"`+redacted+`"`)
	result = compiledPatterns.phone.ReplaceAllStringFunc(result, func(m string) string {
		sub := compiledPatterns.phone.FindStringSubmatch(m)
		if len(sub) < 2 {
			return m
		}
		return fmt.Sprintf(`"phone":"%s"`, MaskPhone(sub[1]))
	})
	result = compiledPatterns.email.ReplaceAllStringFunc(result, func(m string) string {
		sub := compiledPatterns.email.FindStringSubmatch(m)
		if len(sub) < 2 {
			return m
		}
		return fmt.Sprintf(`"email":"%s"`, MaskEmail(sub[1]))
	})
	result = compiledPatterns.idCard.ReplaceAllStringFunc(result, func(m string) string {
		sub := compiledPatterns.idCard.FindStringSubmatch(m)
		if len(sub) < 2 {
			return m
		}
		return fmt.Sprintf(`"id_card":"%s"`, MaskIDCard(sub[1]))
	})
	return result
}

// MaskPhone 中间 4 位隐藏，例如 138****5678。
func MaskPhone(phone string) string {
	runes := []rune(phone)
	if len(runes) < 7 {
		return redacted
	}
	return string(runes[:3]) + "****" + string(runes[len(runes)-4:])
}

// MaskEmail 本地部分除首字符外隐藏，例如 z***@example.com。
func MaskEmail(email string) string {
	at := strings.LastIndex(email, "@")
	if at <= 0 {
		return redacted
	}
	local := email[:at]
	domain := email[at:]
	first := string([]rune(local)[0])
	return first + "***" + domain
}

// MaskIDCard 保留前 3 后 4，中间以 * 填充（18 位时为 110**********1234）。
func MaskIDCard(id string) string {
	runes := []rune(id)
	if len(runes) < 7 {
		return redacted
	}
	stars := len(runes) - 7
	if stars < 8 {
		stars = 8
	}
	return string(runes[:3]) + strings.Repeat("*", stars) + string(runes[len(runes)-4:])
}
