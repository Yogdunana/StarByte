package activation

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/user/model"
	"github.com/Yogdunana/StarByte/backend/pkg/locale"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const tokenTTL = 24 * time.Hour
const resendInterval = time.Minute

type Mailer interface {
	Send(ctx context.Context, to, subject, htmlBody string) error
}

type MailFunc func(ctx context.Context, to, subject, htmlBody string) error

func (f MailFunc) Send(ctx context.Context, to, subject, htmlBody string) error {
	if f == nil {
		return fmt.Errorf("mailer is not configured")
	}
	return f(ctx, to, subject, htmlBody)
}

// Service issues and consumes email verification tokens.
type Service struct {
	db         *gorm.DB
	mailer     Mailer
	publicBase string
}

type tokenRow struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null"`
	TokenHash string     `gorm:"type:varchar(64);uniqueIndex;not null"`
	ExpiresAt time.Time  `gorm:"not null"`
	UsedAt    *time.Time `gorm:"column:used_at"`
	CreatedAt time.Time
}

func (tokenRow) TableName() string { return "email_verification_tokens" }

// New creates an email activation service. publicBase is the site origin used in links.
func New(db *gorm.DB, mailer Mailer, publicBase string) *Service {
	return &Service{db: db, mailer: mailer, publicBase: strings.TrimRight(strings.TrimSpace(publicBase), "/")}
}

// Ready reports whether verification mail can be sent.
func (s *Service) Ready() bool {
	return s != nil && s.db != nil && s.mailer != nil
}

// Verified reports whether the account may log in.
func Verified(user *model.User) bool {
	return user != nil && user.EmailVerifiedAt != nil && !user.EmailVerifiedAt.IsZero()
}

// Start creates a token and sends the activation mail. publicOrigin overrides the configured base when set.
// A second call within resendInterval is a no-op so login/CAS retries cannot flood SMTP.
func (s *Service) Start(ctx context.Context, user *model.User, publicOrigin string) error {
	if user == nil {
		return response.NewError(response.CodeUserNotFound, "用户不存在")
	}
	if !s.Ready() {
		return response.NewError(response.CodeNotificationEmailFail, "邮件服务未配置，无法完成注册")
	}
	email := strings.TrimSpace(user.Email)
	if email == "" {
		return response.NewError(response.CodeBadRequest, "注册需要有效邮箱")
	}
	if Verified(user) {
		return nil
	}
	latest, err := s.latestUnusedToken(ctx, user.ID)
	if err != nil {
		return err
	}
	if latest != nil && time.Since(latest.CreatedAt) < resendInterval {
		return nil
	}
	raw, hash, err := newToken()
	if err != nil {
		return fmt.Errorf("verification token: %w", err)
	}
	row := tokenRow{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: hash,
		ExpiresAt: time.Now().Add(tokenTTL),
	}
	if err := s.db.WithContext(ctx).Create(&row).Error; err != nil {
		return fmt.Errorf("store verification token: %w", err)
	}
	link := s.verifyURL(publicOrigin, raw)
	lang := locale.FromContext(ctx)
	subject := locale.Text(lang, "激活你的 StarByte 账号")
	body := fmt.Sprintf(
		"<p>%s</p><p><a href=\"%s\">%s</a></p><p>%s</p>",
		locale.Text(lang, "请点击以下链接激活账号后登录："),
		link,
		link,
		locale.Text(lang, "激活链接 24 小时内有效。若非本人操作请忽略此邮件。"),
	)
	if err := s.mailer.Send(ctx, email, subject, body); err != nil {
		return response.NewError(response.CodeNotificationEmailFail, "验证邮件发送失败，请稍后重试")
	}
	return nil
}

// Confirm activates the account for a raw token from the email link.
func (s *Service) Confirm(ctx context.Context, rawToken string) error {
	if !s.Ready() {
		return response.NewError(response.CodeNotificationEmailFail, "邮件服务未配置")
	}
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		return response.NewError(response.CodeEmailVerifyInvalid, "验证链接无效或已过期")
	}
	hash := hashToken(rawToken)
	var row tokenRow
	err := s.db.WithContext(ctx).Where("token_hash = ?", hash).First(&row).Error
	if err == gorm.ErrRecordNotFound {
		return response.NewError(response.CodeEmailVerifyInvalid, "验证链接无效或已过期")
	}
	if err != nil {
		return err
	}
	if row.UsedAt != nil || time.Now().After(row.ExpiresAt) {
		return response.NewError(response.CodeEmailVerifyInvalid, "验证链接无效或已过期")
	}
	now := time.Now()
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&tokenRow{}).Where("id = ? AND used_at IS NULL", row.ID).Update("used_at", now).Error; err != nil {
			return err
		}
		return tx.Model(&model.User{}).Where("id = ?", row.UserID).Update("email_verified_at", now).Error
	})
}

// Resend sends a new link for an unverified account identified by username or email.
func (s *Service) Resend(ctx context.Context, identifier, publicOrigin string) error {
	if !s.Ready() {
		return response.NewError(response.CodeNotificationEmailFail, "邮件服务未配置")
	}
	identifier = strings.TrimSpace(identifier)
	if identifier == "" {
		return response.NewError(response.CodeBadRequest, "请输入用户名或邮箱")
	}
	var user model.User
	err := s.db.WithContext(ctx).Where("username = ? OR email = ?", identifier, identifier).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		return nil
	}
	if err != nil {
		return err
	}
	if Verified(&user) {
		return nil
	}
	latest, err := s.latestUnusedToken(ctx, user.ID)
	if err != nil {
		return err
	}
	if latest != nil && time.Since(latest.CreatedAt) < resendInterval {
		return response.NewError(response.CodeTooManyReq, "验证邮件发送过于频繁，请稍后再试")
	}
	return s.Start(ctx, &user, publicOrigin)
}

func (s *Service) latestUnusedToken(ctx context.Context, userID uuid.UUID) (*tokenRow, error) {
	var latest tokenRow
	err := s.db.WithContext(ctx).Where("user_id = ? AND used_at IS NULL", userID).Order("created_at DESC").First(&latest).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &latest, nil
}

func (s *Service) verifyURL(publicOrigin, raw string) string {
	base := strings.TrimRight(strings.TrimSpace(publicOrigin), "/")
	if base == "" {
		base = s.publicBase
	}
	if base == "" {
		base = "http://127.0.0.1"
	}
	return base + "/verify-email?token=" + raw
}

func newToken() (raw, hash string, err error) {
	buf := make([]byte, 32)
	if _, err = rand.Read(buf); err != nil {
		return "", "", err
	}
	raw = hex.EncodeToString(buf)
	return raw, hashToken(raw), nil
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
