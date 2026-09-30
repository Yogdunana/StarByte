package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidatePasswordStrength_Valid(t *testing.T) {
	valid := []string{
		"Passw0rd!", // 四类齐全
		"Aa1b2c3d",  // 小写 + 大写 + 数字
		"abc123!@#", // 小写 + 数字 + 符号
		"ABC123!@#", // 大写 + 数字 + 符号
		"AbcDef!@#", // 小写 + 大写 + 符号
		"aA1!aA1!",  // 刚好 8 位
		"Password12",
	}
	for _, pwd := range valid {
		assert.True(t, ValidatePasswordStrength(pwd), "password should be valid: %s", pwd)
	}
}

func TestValidatePasswordStrength_TooShort(t *testing.T) {
	short := []string{"", "a", "ab1", "1234567", "Aa1!aA1"}
	for _, pwd := range short {
		assert.False(t, ValidatePasswordStrength(pwd), "password should be rejected (too short): %s", pwd)
	}
}

// TestValidatePasswordStrength_OnlyTwoClasses 守住"至少三种"这条线：
// 长度够但只占两类的口令必须拒，这是本次策略调整的核心。
func TestValidatePasswordStrength_OnlyTwoClasses(t *testing.T) {
	weak := []string{
		"password12", // 小写 + 数字
		"PASSWORD12", // 大写 + 数字
		"password!!", // 小写 + 符号
		"PASSWORD!!", // 大写 + 符号
		"1234!@#$",   // 数字 + 符号
		"abcdefgh",   // 纯小写
		"ABCDEFGH",   // 纯大写
		"12345678",   // 纯数字
		"!@#$%^&*",   // 纯符号
	}
	for _, pwd := range weak {
		assert.False(t, ValidatePasswordStrength(pwd), "password should be rejected (only two classes): %s", pwd)
	}
}

// TestValidatePasswordStrength_NonASCII 保证中文等非 ASCII 字符不能顶替一整类，
// 否则「密码ab12」这种两段中文 + 两类字符就能凑满三类。
func TestValidatePasswordStrength_NonASCII(t *testing.T) {
	assert.False(t, ValidatePasswordStrength("密码abcd"))
	assert.False(t, ValidatePasswordStrength("密码1234"))
	assert.False(t, ValidatePasswordStrength("密码ab12"))
}

func TestPasswordPolicyHint(t *testing.T) {
	// 改策略时最容易漏的就是只改代码不改文案，这里把文案里的两个要点钉住。
	assert.Contains(t, PasswordPolicyHint, "8")
	assert.Contains(t, PasswordPolicyHint, "三种")
}

func TestHashPassword_AndCheckPassword(t *testing.T) {
	password := "testpass123"
	hash, err := HashPassword(password)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)

	// Correct password
	assert.True(t, CheckPassword(password, hash))
	// Wrong password
	assert.False(t, CheckPassword("wrongpass", hash))
}

func TestHashPassword_DifferentHashes(t *testing.T) {
	password := "samepassword123"
	hash1, _ := HashPassword(password)
	hash2, _ := HashPassword(password)
	// bcrypt generates different hashes due to random salt
	assert.NotEqual(t, hash1, hash2)
}
