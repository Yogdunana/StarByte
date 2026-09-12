package config

import (
	"testing"
)

func TestSMTPPasswordFromEnvPrefersStarbyte(t *testing.T) {
	t.Setenv("STARBYTE_SMTP_PASSWORD", "preferred")
	t.Setenv("SMTP_PASSWORD", "legacy")
	if got := SMTPPasswordFromEnv(); got != "preferred" {
		t.Fatalf("SMTPPasswordFromEnv = %q, want preferred", got)
	}
	if got := SMTPPasswordSource(); got != "STARBYTE_SMTP_PASSWORD" {
		t.Fatalf("SMTPPasswordSource = %q", got)
	}
}

func TestSMTPPasswordFromEnvLegacyFallback(t *testing.T) {
	t.Setenv("STARBYTE_SMTP_PASSWORD", "")
	t.Setenv("SMTP_PASSWORD", "legacy")
	if got := SMTPPasswordFromEnv(); got != "legacy" {
		t.Fatalf("SMTPPasswordFromEnv = %q, want legacy", got)
	}
}

func TestOverlayAndNormalize(t *testing.T) {
	base := EmailConfig{SMTPHost: "old", SMTPPort: 25, From: "a@b.c", SSLMode: SSLModeNone}
	over, err := ParseSMTPRuntime(`{"host":"smtp.exmail.qq.com","port":465,"ssl_mode":"ssl","from":"me@x.test","from_name":"StarByte-SMTP"}`)
	if err != nil {
		t.Fatal(err)
	}
	got := over.Overlay(base)
	if got.SMTPHost != "smtp.exmail.qq.com" || got.SMTPPort != 465 || got.SSLMode != SSLModeImplicit {
		t.Fatalf("overlay = %+v", got)
	}
	if got.FromName != "StarByte-SMTP" {
		t.Fatalf("from_name = %s", got.FromName)
	}
}

func TestEffectiveSSLModeFromPort(t *testing.T) {
	if (EmailConfig{SMTPPort: 465}).EffectiveSSLMode() != SSLModeImplicit {
		t.Fatal("465 should default to implicit")
	}
	if (EmailConfig{SMTPPort: 587}).EffectiveSSLMode() != SSLModeStartTLS {
		t.Fatal("587 should default to starttls")
	}
}

func TestApplyEnvPasswordIgnoresStructValue(t *testing.T) {
	t.Setenv("STARBYTE_SMTP_PASSWORD", "env-secret")
	got := EmailConfig{Password: "should-not-win"}.ApplyEnvPassword()
	if got.Password != "env-secret" {
		t.Fatalf("password = %q", got.Password)
	}
}
