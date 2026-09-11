package service

import (
	"context"
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
	cfg := ch.resolve(context.Background())
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
	ch := NewEmailChannel("", 0, "", "", "")
	assert.False(t, ch.IsAvailable())
	ch = NewEmailChannelFromConfig(config.DefaultSMTPRuntime().Overlay(config.EmailConfig{}))
	assert.True(t, ch.IsAvailable())
}
