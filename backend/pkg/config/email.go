package config

import (
	"encoding/json"
	"os"
	"strings"
)

const (
	// SMTPSettingsKey is the runtime configs row for SMTP settings and an encrypted password.
	SMTPSettingsKey = "smtp_settings"

	SSLModeImplicit = "implicit"
	SSLModeStartTLS = "starttls"
	SSLModeNone     = "none"

	defaultSMTPHost     = "smtp.exmail.qq.com"
	defaultSMTPPort     = 465
	defaultSMTPFrom     = "computerassociation@smbu.edu.cn"
	defaultSMTPFromName = "StarByte-SMTP"
	defaultSMTPSSLMode  = SSLModeImplicit
)

// SMTPRuntime is the persisted SMTP overlay (passwords are encrypted) stored in configs.
type SMTPRuntime struct {
	PasswordCiphertext string `json:"password_ciphertext,omitempty"`
	Host               string `json:"host"`
	Port               int    `json:"port"`
	SSLMode            string `json:"ssl_mode"`
	From               string `json:"from"`
	FromName           string `json:"from_name"`
	Username           string `json:"username"`
}

// SMTPPasswordFromEnv returns the SMTP password from the preferred secret name,
// then the legacy SMTP_PASSWORD alias. Empty if neither is set.
func SMTPPasswordFromEnv() string {
	if v := strings.TrimSpace(os.Getenv("STARBYTE_SMTP_PASSWORD")); v != "" {
		return v
	}
	return strings.TrimSpace(os.Getenv("SMTP_PASSWORD"))
}

// SMTPPasswordSource reports which env var currently supplies the password.
func SMTPPasswordSource() string {
	if strings.TrimSpace(os.Getenv("STARBYTE_SMTP_PASSWORD")) != "" {
		return "STARBYTE_SMTP_PASSWORD"
	}
	if strings.TrimSpace(os.Getenv("SMTP_PASSWORD")) != "" {
		return "SMTP_PASSWORD"
	}
	return ""
}

// NormalizeSSLMode maps user/env input to a known SSL mode. Empty stays empty.
func NormalizeSSLMode(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case SSLModeImplicit, "ssl", "tls", "smtps":
		return SSLModeImplicit
	case SSLModeStartTLS, "start_tls":
		return SSLModeStartTLS
	case SSLModeNone, "off", "disable", "plain":
		return SSLModeNone
	default:
		return strings.TrimSpace(raw)
	}
}

func ValidSSLMode(mode string) bool {
	switch NormalizeSSLMode(mode) {
	case SSLModeImplicit, SSLModeStartTLS, SSLModeNone:
		return true
	default:
		return false
	}
}

func DefaultSMTPRuntime() SMTPRuntime {
	return SMTPRuntime{
		Host:     defaultSMTPHost,
		Port:     defaultSMTPPort,
		SSLMode:  defaultSMTPSSLMode,
		From:     defaultSMTPFrom,
		FromName: defaultSMTPFromName,
		Username: defaultSMTPFrom,
	}
}

func ParseSMTPRuntime(raw string) (SMTPRuntime, error) {
	out := SMTPRuntime{}
	if strings.TrimSpace(raw) == "" {
		return out, nil
	}
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return SMTPRuntime{}, err
	}
	out.SSLMode = NormalizeSSLMode(out.SSLMode)
	return out, nil
}

func (s SMTPRuntime) Overlay(base EmailConfig) EmailConfig {
	if v := strings.TrimSpace(s.Host); v != "" {
		base.SMTPHost = v
	}
	if s.Port > 0 {
		base.SMTPPort = s.Port
	}
	if v := NormalizeSSLMode(s.SSLMode); ValidSSLMode(v) {
		base.SSLMode = v
	}
	if v := strings.TrimSpace(s.From); v != "" {
		base.From = v
	}
	if v := strings.TrimSpace(s.FromName); v != "" {
		base.FromName = v
	}
	if v := strings.TrimSpace(s.Username); v != "" {
		base.Username = v
	}
	return base
}

func (e EmailConfig) EffectiveUsername() string {
	if v := strings.TrimSpace(e.Username); v != "" {
		return v
	}
	return strings.TrimSpace(e.From)
}

func (e EmailConfig) EffectiveSSLMode() string {
	if v := NormalizeSSLMode(e.SSLMode); ValidSSLMode(v) {
		return v
	}
	if e.SMTPPort == 465 {
		return SSLModeImplicit
	}
	if e.SMTPPort == 587 {
		return SSLModeStartTLS
	}
	return defaultSMTPSSLMode
}

// ApplyEnvPassword refreshes the legacy password fallback from the environment.
func (e EmailConfig) ApplyEnvPassword() EmailConfig {
	if v := SMTPPasswordFromEnv(); v != "" {
		e.Password = v
		e.PasswordSource = SMTPPasswordSource()
	}
	return e
}

func firstEnv(keys ...string) string {
	for _, key := range keys {
		if v := strings.TrimSpace(os.Getenv(key)); v != "" {
			return v
		}
	}
	return ""
}
