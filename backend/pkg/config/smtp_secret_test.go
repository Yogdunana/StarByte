package config

import (
	"encoding/base64"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestSMTPSecretRoundTripAndFailClosed(t *testing.T) {
	t.Setenv("STARBYTE_CONFIG_ENCRYPTION_KEY", base64.StdEncoding.EncodeToString([]byte(strings.Repeat("k", 32))))
	t.Setenv("STARBYTE_SMTP_PASSWORD", "environment-fallback")
	encrypted, err := EncryptSMTPPassword(" web-secret ")
	require.NoError(t, err)
	another, err := EncryptSMTPPassword(" web-secret ")
	require.NoError(t, err)
	require.NotEqual(t, encrypted, another)
	runtime := SMTPRuntime{PasswordCiphertext: encrypted}
	cfg, err := runtime.Resolve(EmailConfig{})
	require.NoError(t, err)
	require.Equal(t, " web-secret ", cfg.Password)
	require.Equal(t, "web", cfg.PasswordSource)
	t.Setenv("STARBYTE_CONFIG_ENCRYPTION_KEY", base64.StdEncoding.EncodeToString([]byte(strings.Repeat("x", 32))))
	cfg, err = runtime.Resolve(EmailConfig{})
	require.Error(t, err)
	require.Empty(t, cfg.Password)
	t.Setenv("STARBYTE_CONFIG_ENCRYPTION_KEY", "")
	_, err = EncryptSMTPPassword("new")
	require.Error(t, err)
	cfg, err = (SMTPRuntime{}).Resolve(EmailConfig{})
	require.NoError(t, err)
	require.Equal(t, "environment-fallback", cfg.Password)
}
