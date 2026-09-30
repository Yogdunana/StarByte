package utils

import (
	"golang.org/x/crypto/bcrypt"
)

// bcryptCost is the bcrypt cost factor. Issue #17 requires cost=12.
const bcryptCost = 12

// HashPassword returns a bcrypt hash of the given plain-text password.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPassword reports whether the plain-text password matches the bcrypt hash.
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// passwordMinLen 是口令最小长度。测试组要求「8 位以上」。
const passwordMinLen = 8

// passwordMinClasses 是口令必须命中的字符种类数：数字、小写字母、大写字母、
// 特殊字符共四类，至少占三类。只要求「字母+数字」时 Passw0rd 这类弱口令能过，
// 达不到等保口令复杂度要求。
const passwordMinClasses = 3

// PasswordPolicyHint 是口令规则的统一文案。每个设置密码的入口都引用它，
// 免得改了规则只改一处、剩下几处还在说旧要求。
const PasswordPolicyHint = "密码强度不足：至少 8 位，且需包含数字、小写字母、大写字母、特殊字符中的至少三种"

// ValidatePasswordStrength reports whether password 满足口令策略：
// 长度不低于 passwordMinLen，且四类字符中至少命中 passwordMinClasses 类。
//
// 只有可见 ASCII（0x21~0x7e）才算特殊字符；中文等非 ASCII 字符不计入任何一类，
// 否则一段中文就能顶掉一整类，策略形同虚设。
func ValidatePasswordStrength(password string) bool {
	if len(password) < passwordMinLen {
		return false
	}
	var hasLower, hasUpper, hasDigit, hasSpecial bool
	for _, ch := range password {
		switch {
		case ch >= 'a' && ch <= 'z':
			hasLower = true
		case ch >= 'A' && ch <= 'Z':
			hasUpper = true
		case ch >= '0' && ch <= '9':
			hasDigit = true
		case ch >= 0x21 && ch <= 0x7e:
			hasSpecial = true
		}
	}
	classes := 0
	for _, ok := range []bool{hasLower, hasUpper, hasDigit, hasSpecial} {
		if ok {
			classes++
		}
	}
	return classes >= passwordMinClasses
}
