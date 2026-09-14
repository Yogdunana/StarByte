package activation

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/user/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/Yogdunana/StarByte/backend/pkg/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type captureMailer struct {
	to, subject, body string
	err               error
}

func (c *captureMailer) Send(_ context.Context, to, subject, htmlBody string) error {
	c.to, c.subject, c.body = to, subject, htmlBody
	return c.err
}

func TestStartConfirmAndLoginGate(t *testing.T) {
	tx := testutil.OpenPostgres(t).Begin()
	defer tx.Rollback()
	user := &model.User{
		ID: uuid.New(), Username: "act-" + uuid.NewString()[:8],
		PasswordHash: "x", Email: "act@example.test", Status: 0,
	}
	require.NoError(t, tx.Create(user).Error)
	mail := &captureMailer{}
	svc := New(tx, mail, "http://10.100.13.17")
	require.False(t, Verified(user))
	require.NoError(t, svc.Start(context.Background(), user, ""))
	require.Equal(t, "act@example.test", mail.to)
	require.Contains(t, mail.body, "http://10.100.13.17/verify-email?token=")
	raw := tokenFromBody(t, mail.body)
	require.NoError(t, svc.Confirm(context.Background(), raw))
	var got model.User
	require.NoError(t, tx.First(&got, "id = ?", user.ID).Error)
	require.True(t, Verified(&got))
	require.Error(t, svc.Confirm(context.Background(), raw))
}

func TestStartRequiresMailerAndEmail(t *testing.T) {
	svc := New(nil, nil, "")
	err := svc.Start(context.Background(), &model.User{Email: "a@b.c"}, "")
	require.Error(t, err)
	require.Equal(t, response.CodeNotificationEmailFail, err.(*response.AppError).Code)
	tx := testutil.OpenPostgres(t).Begin()
	defer tx.Rollback()
	mail := &captureMailer{}
	ready := New(tx, mail, "http://example.test")
	err = ready.Start(context.Background(), &model.User{ID: uuid.New(), Email: ""}, "")
	require.Error(t, err)
	require.Equal(t, response.CodeBadRequest, err.(*response.AppError).Code)
}

func TestResendRateLimit(t *testing.T) {
	tx := testutil.OpenPostgres(t).Begin()
	defer tx.Rollback()
	user := &model.User{
		ID: uuid.New(), Username: "rsd-" + uuid.NewString()[:8],
		PasswordHash: "x", Email: "rsd@example.test", Status: 0,
	}
	require.NoError(t, tx.Create(user).Error)
	mail := &captureMailer{}
	svc := New(tx, mail, "http://example.test")
	require.NoError(t, svc.Start(context.Background(), user, ""))
	firstBody := mail.body
	require.NoError(t, svc.Resend(context.Background(), user.Email, ""))
	require.Equal(t, firstBody, mail.body)
	require.NoError(t, tx.Model(&tokenRow{}).Where("user_id = ?", user.ID).Update("created_at", time.Now().Add(-2*time.Minute)).Error)
	require.NoError(t, svc.Resend(context.Background(), user.Username, "http://override.test"))
	require.Contains(t, mail.body, "http://override.test/verify-email?token=")
}

func TestStartIsIdempotentWithinInterval(t *testing.T) {
	tx := testutil.OpenPostgres(t).Begin()
	defer tx.Rollback()
	user := &model.User{
		ID: uuid.New(), Username: "idp-" + uuid.NewString()[:8],
		PasswordHash: "x", Email: "idp@example.test", Status: 0,
	}
	require.NoError(t, tx.Create(user).Error)
	mail := &countingMailer{}
	svc := New(tx, mail, "http://example.test")
	require.NoError(t, svc.Start(context.Background(), user, ""))
	require.NoError(t, svc.Start(context.Background(), user, ""))
	require.Equal(t, 1, mail.n)
	var n int64
	require.NoError(t, tx.Model(&tokenRow{}).Where("user_id = ?", user.ID).Count(&n).Error)
	require.Equal(t, int64(1), n)
}

type countingMailer struct{ n int }

func (c *countingMailer) Send(_ context.Context, _, _, _ string) error {
	c.n++
	return nil
}

func TestResendUnknownIdentifierIsSilent(t *testing.T) {
	tx := testutil.OpenPostgres(t).Begin()
	defer tx.Rollback()
	mail := &captureMailer{}
	svc := New(tx, mail, "http://example.test")
	require.NoError(t, svc.Resend(context.Background(), "nobody@example.test", ""))
	require.Empty(t, mail.to)
}

func TestResendByStudentNo(t *testing.T) {
	tx := testutil.OpenPostgres(t).Begin()
	defer tx.Rollback()
	user := &model.User{
		ID: uuid.New(), Username: "stu-" + uuid.NewString()[:8],
		PasswordHash: "x", Email: "stu@example.test", Status: 0,
	}
	require.NoError(t, tx.Create(user).Error)
	studentNo := "1120" + uuid.NewString()[:6]
	require.NoError(t, tx.Exec(
		`INSERT INTO member_profiles (id, user_id, real_name, student_no) VALUES (?, ?, ?, ?)`,
		uuid.New(), user.ID, "段茗尧", studentNo,
	).Error)
	mail := &captureMailer{}
	svc := New(tx, mail, "http://example.test")
	require.NoError(t, svc.Resend(context.Background(), studentNo, ""))
	require.Equal(t, "stu@example.test", mail.to)
	require.NoError(t, svc.Resend(context.Background(), studentNo, ""))
	require.Equal(t, "stu@example.test", mail.to)
}

func tokenFromBody(t *testing.T, body string) string {
	t.Helper()
	const marker = "token="
	i := strings.Index(body, marker)
	require.GreaterOrEqual(t, i, 0)
	raw := body[i+len(marker):]
	if j := strings.IndexAny(raw, `"<> \n`); j >= 0 {
		raw = raw[:j]
	}
	raw = strings.TrimSpace(raw)
	require.NotEmpty(t, raw)
	return raw
}

func TestVerifyURLSanitizesOrigin(t *testing.T) {
	svc := New(nil, nil, "http://10.100.13.17/app")
	require.Equal(t, "http://10.100.13.17/verify-email?token=abc", svc.verifyURL("javascript:alert(1)", "abc"))
	require.Equal(t, "https://safe.example/verify-email?token=abc", svc.verifyURL("https://safe.example/phish", "abc"))
	require.Equal(t, "http://10.100.13.17/verify-email?token=abc", svc.verifyURL("//evil.example", "abc"))
	require.Equal(t, "http://10.100.13.17/verify-email?token=abc", svc.verifyURL("https://user:pass@evil.example", "abc"))
}
