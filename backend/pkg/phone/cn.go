package phone

import (
	"errors"
	"regexp"
	"strings"
)

var (
	// ErrInvalidCN 不是有效的中国大陆手机号。
	ErrInvalidCN = errors.New("invalid mainland mobile number")
	nonDigit     = regexp.MustCompile(`\D`)
	cnMobile     = regexp.MustCompile(`^1[3-9]\d{9}$`)
)

// NormalizeCN 把 +86 / 86 / 纯数字写成 11 位大陆手机号。
func NormalizeCN(raw string) (string, error) {
	digits := nonDigit.ReplaceAllString(strings.TrimSpace(raw), "")
	switch {
	case strings.HasPrefix(digits, "0086") && len(digits) >= 15:
		digits = digits[4:]
	case strings.HasPrefix(digits, "86") && len(digits) >= 13:
		digits = digits[2:]
	}
	if !cnMobile.MatchString(digits) {
		return "", ErrInvalidCN
	}
	return digits, nil
}

// NormalizeCNOptional 允许空值；有值则必须是大陆手机号。
func NormalizeCNOptional(raw string) (string, error) {
	if strings.TrimSpace(raw) == "" {
		return "", nil
	}
	return NormalizeCN(raw)
}
