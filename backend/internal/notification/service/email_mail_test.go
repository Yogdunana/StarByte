package service

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/Yogdunana/StarByte/backend/pkg/config"
	"github.com/Yogdunana/StarByte/backend/pkg/configstore"
	"github.com/go-mail/mail"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmailChannelResolvesRuntimeOverlay(t *testing.T) {
	t.Setenv("STARBYTE_SMTP_PASSWORD", "env-pass")
	backend := configstore.NewMemoryBackend()
	require.NoError(t, backend.Save(context.Background(), config.SMTPSettingsKey,
		`{"host":"smtp.runtime.test","port":465,"ssl_mode":"implicit","from":"from@runtime.test","from_name":"Runtime","username":"user@runtime.test"}`))
	ch := NewEmailChannel("smtp.fallback.test", 25, "old", "yaml-pass", "old@x.test").
		WithStore(configstore.New(nil, backend))
	cfg, err := ch.resolve(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "smtp.runtime.test", cfg.SMTPHost)
	assert.Equal(t, 465, cfg.SMTPPort)
	assert.Equal(t, "from@runtime.test", cfg.From)
	assert.Equal(t, "Runtime", cfg.FromName)
	assert.Equal(t, "env-pass", cfg.Password)
	assert.Equal(t, "user@runtime.test", cfg.EffectiveUsername())
}

func TestApplySMTPSecurity(t *testing.T) {
	d := mail.NewDialer("h", 465, "u", "p")
	applySMTPSecurity(d, config.SSLModeImplicit)
	assert.True(t, d.SSL)

	d = mail.NewDialer("h", 587, "u", "p")
	applySMTPSecurity(d, config.SSLModeStartTLS)
	assert.False(t, d.SSL)
	if d.StartTLSPolicy != mail.MandatoryStartTLS {
		t.Fatalf("starttls policy = %v, want MandatoryStartTLS", d.StartTLSPolicy)
	}

	d = mail.NewDialer("h", 25, "u", "p")
	applySMTPSecurity(d, config.SSLModeNone)
	assert.False(t, d.SSL)
	if d.StartTLSPolicy != mail.NoStartTLS {
		t.Fatalf("none policy = %v, want NoStartTLS", d.StartTLSPolicy)
	}
}

func TestEmailChannelIsAvailableUsesResolvedFrom(t *testing.T) {
	t.Setenv("STARBYTE_SMTP_PASSWORD", "")
	t.Setenv("SMTP_PASSWORD", "")
	ch := NewEmailChannel("", 0, "", "", "")
	assert.False(t, ch.IsAvailable())

	ch = NewEmailChannelFromConfig(config.DefaultSMTPRuntime().Overlay(config.EmailConfig{}))
	assert.False(t, ch.IsAvailable(), "campus defaults without env password must stay unavailable")

	// YAML leftover password must not count — same bar as TestSMTP (env only).
	ch = NewEmailChannel("smtp.exmail.qq.com", 465, "u", "yaml-only", "a@b.c")
	assert.False(t, ch.IsAvailable())

	t.Setenv("STARBYTE_SMTP_PASSWORD", "env-pass")
	assert.True(t, ch.IsAvailable())
}

func TestSendMIMERefusesMissingPassword(t *testing.T) {
	t.Setenv("STARBYTE_SMTP_PASSWORD", "")
	t.Setenv("SMTP_PASSWORD", "")
	ch := NewEmailChannelFromConfig(config.DefaultSMTPRuntime().Overlay(config.EmailConfig{}))
	err := ch.SendMIME(context.Background(), MailJob{To: []string{"a@b.c"}, Subject: "s", Body: "b"}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "password")

	// Would otherwise DialAndSend smtp.exmail.qq.com; yaml password is not enough.
	ch = NewEmailChannel("smtp.exmail.qq.com", 465, "u", "yaml-only", "a@b.c")
	err = ch.SendMIME(context.Background(), MailJob{To: []string{"a@b.c"}, Subject: "s", Body: "b"}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "password")
}

func TestEmailChannelUsesWebSecretAndRejectsWrongKey(t *testing.T) {
	t.Setenv("STARBYTE_CONFIG_ENCRYPTION_KEY", "a2tra2tra2tra2tra2tra2tra2tra2tra2tra2tra2s=")
	t.Setenv("STARBYTE_SMTP_PASSWORD", "")
	t.Setenv("SMTP_PASSWORD", "")
	encrypted, err := config.EncryptSMTPPassword("web-secret")
	require.NoError(t, err)
	runtime := config.SMTPRuntime{PasswordCiphertext: encrypted}
	raw, err := json.Marshal(runtime)
	require.NoError(t, err)
	backend := configstore.NewMemoryBackend()
	require.NoError(t, backend.Save(context.Background(), config.SMTPSettingsKey, string(raw)))
	ch := NewEmailChannel("smtp.example.test", 465, "user", "old", "a@example.test").WithStore(configstore.New(nil, backend))
	cfg, err := ch.resolve(context.Background())
	require.NoError(t, err)
	require.Equal(t, "web-secret", cfg.Password)
	require.True(t, ch.IsAvailable(), "encrypted web password works without an environment SMTP password")
	t.Setenv("STARBYTE_CONFIG_ENCRYPTION_KEY", "")
	require.False(t, ch.IsAvailable())
	err = ch.SendTest(context.Background(), "to@example.test")
	require.Error(t, err)
}
