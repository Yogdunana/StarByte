package service

import (
	"context"
	"testing"

	"github.com/Yogdunana/StarByte/backend/internal/configstore/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/config"
	"github.com/Yogdunana/StarByte/backend/pkg/configstore"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubTester struct {
	to  string
	err error
}

func (s *stubTester) SendTest(_ context.Context, to string) error {
	s.to = to
	return s.err
}

func newSMTPSvc(t *testing.T) (ConfigService, *stubTester) {
	t.Helper()
	tester := &stubTester{}
	svc := NewConfigService(newMemRepo(), configstore.New(nil, configstore.NewMemoryBackend())).WithSMTP(
		config.DefaultSMTPRuntime().Overlay(config.EmailConfig{}), tester,
	)
	return svc, tester
}

func TestGetSMTPDefaultsAndNoPassword(t *testing.T) {
	t.Setenv("STARBYTE_SMTP_PASSWORD", "")
	t.Setenv("SMTP_PASSWORD", "")
	svc, _ := newSMTPSvc(t)
	got, err := svc.GetSMTP(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "smtp.exmail.qq.com", got.Host)
	assert.Equal(t, 465, got.Port)
	assert.Equal(t, "implicit", got.SSLMode)
	assert.Equal(t, "computerassociation@smbu.edu.cn", got.From)
	assert.Equal(t, "StarByte-SMTP", got.FromName)
	assert.False(t, got.PasswordConfigured)
	assert.Empty(t, got.PasswordSource)
}

func TestUpdateSMTPPersistsAndMasksSecret(t *testing.T) {
	t.Setenv("STARBYTE_SMTP_PASSWORD", "do-not-echo")
	svc, _ := newSMTPSvc(t)
	got, err := svc.UpdateSMTP(context.Background(), uuid.New(), &dto.UpdateSMTPRequest{
		Host: "smtp.example.test", Port: 587, SSLMode: "starttls",
		From: "noreply@example.test", FromName: "Demo",
	})
	require.NoError(t, err)
	assert.Equal(t, "smtp.example.test", got.Host)
	assert.Equal(t, 587, got.Port)
	assert.Equal(t, "starttls", got.SSLMode)
	assert.True(t, got.PasswordConfigured)
	assert.Equal(t, "STARBYTE_SMTP_PASSWORD", got.PasswordSource)

	row, err := svc.GetByKey(context.Background(), config.SMTPSettingsKey)
	require.NoError(t, err)
	assert.NotContains(t, row.ConfigValue, "do-not-echo")
	assert.NotContains(t, row.ConfigValue, "password")
}

func TestUpdateSMTPRejectsBadSSL(t *testing.T) {
	svc, _ := newSMTPSvc(t)
	_, err := svc.UpdateSMTP(context.Background(), uuid.New(), &dto.UpdateSMTPRequest{
		Host: "h", Port: 25, SSLMode: "weird", From: "a@b.c", FromName: "n",
	})
	assert.Error(t, err)
}

func TestTestSMTPRequiresPassword(t *testing.T) {
	t.Setenv("STARBYTE_SMTP_PASSWORD", "")
	t.Setenv("SMTP_PASSWORD", "")
	svc, tester := newSMTPSvc(t)
	_, err := svc.TestSMTP(context.Background(), &dto.TestSMTPRequest{To: "a@b.c"})
	assert.Error(t, err)
	assert.Empty(t, tester.to)
}

func TestTestSMTPSends(t *testing.T) {
	t.Setenv("STARBYTE_SMTP_PASSWORD", "secret")
	svc, tester := newSMTPSvc(t)
	out, err := svc.TestSMTP(context.Background(), &dto.TestSMTPRequest{To: "ops@example.test"})
	require.NoError(t, err)
	assert.True(t, out.Sent)
	assert.Equal(t, "ops@example.test", tester.to)
}

func TestProtectedSMTPKey(t *testing.T) {
	t.Setenv("STARBYTE_SMTP_PASSWORD", "")
	svc, _ := newSMTPSvc(t)
	_, err := svc.UpdateSMTP(context.Background(), uuid.New(), &dto.UpdateSMTPRequest{
		Host: "smtp.example.test", Port: 465, SSLMode: "implicit",
		From: "a@b.c", FromName: "N",
	})
	require.NoError(t, err)
	row, err := svc.GetByKey(context.Background(), config.SMTPSettingsKey)
	require.NoError(t, err)
	err = svc.Delete(context.Background(), uuid.MustParse(row.ID))
	assert.Error(t, err)
}
