package audit

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDesensitizeJSON_Empty(t *testing.T) {
	assert.Equal(t, "", DesensitizeJSON(""))
}

func TestDesensitizeJSON_PasswordAndToken(t *testing.T) {
	input := `{"username":"admin","password":"secret123","token":"abc"}`
	result := DesensitizeJSON(input)
	assert.Contains(t, result, `"password":"***"`)
	assert.Contains(t, result, `"token":"***"`)
	assert.NotContains(t, result, "secret123")
	assert.Contains(t, result, "admin")
}

func TestDesensitizeJSON_OldNewPasswordSecret(t *testing.T) {
	input := `{"old_password":"old123","new_password":"new456","secret":"s"}`
	result := DesensitizeJSON(input)
	assert.NotContains(t, result, "old123")
	assert.NotContains(t, result, "new456")
	assert.NotContains(t, result, `"secret":"s"`)
	assert.Contains(t, result, `"***"`)
}

func TestDesensitizeJSON_PhoneEmailIDCard(t *testing.T) {
	input := `{"phone":"13812345678","email":"zhang@example.com","id_card":"110101199001011234"}`
	result := DesensitizeJSON(input)
	assert.Contains(t, result, "138****5678")
	assert.Contains(t, result, "z***@example.com")
	assert.Contains(t, result, "110***********1234")
	assert.NotContains(t, result, "13812345678")
	assert.NotContains(t, result, "zhang@example.com")
	assert.NotContains(t, result, "110101199001011234")
}

func TestMaskHelpers(t *testing.T) {
	assert.Equal(t, "138****5678", MaskPhone("13812345678"))
	assert.Equal(t, "z***@example.com", MaskEmail("zhang@example.com"))
	assert.Equal(t, "110***********1234", MaskIDCard("110101199001011234"))
	assert.Equal(t, "***", MaskPhone("123"))
	assert.Equal(t, "***", MaskEmail("bad"))
}

func TestDesensitizeJSON_MalformedStillMasks(t *testing.T) {
	input := `{"name":"test","password":"secret"`
	result := DesensitizeJSON(input)
	assert.Contains(t, result, "***")
	assert.NotContains(t, result, "secret")
}

func TestDesensitizeJSON_Nested(t *testing.T) {
	input := `{"user":{"email":"ab@x.com","phone":"13900001111"}}`
	result := DesensitizeJSON(input)
	assert.Contains(t, result, "a***@x.com")
	assert.Contains(t, result, "139****1111")
}

func TestDesensitizeJSON_RestoreDrillCredentials(t *testing.T) {
	input := `{"confirmation":"DRILL","target_dbname":"starbyte_drill","target_password":"cross-host-secret","target_dsn":"postgres://u:p@db:5432/starbyte_drill"}`
	result := DesensitizeJSON(input)
	assert.Contains(t, result, `"target_password":"***"`)
	assert.Contains(t, result, `"target_dsn":"***"`)
	assert.Contains(t, result, "starbyte_drill")
	assert.Contains(t, result, "DRILL")
	assert.NotContains(t, result, "cross-host-secret")
	assert.NotContains(t, result, "postgres://u:p@")
}

func TestDesensitizeJSON_MalformedDrillStillMasks(t *testing.T) {
	input := `{"target_password":"cross-host-secret","target_dsn":"postgres://u:p@db/x"`
	result := DesensitizeJSON(input)
	assert.Contains(t, result, "***")
	assert.NotContains(t, result, "cross-host-secret")
	assert.NotContains(t, result, "postgres://u:p@")
}
