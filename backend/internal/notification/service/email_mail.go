package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/Yogdunana/StarByte/backend/pkg/config"
	"github.com/Yogdunana/StarByte/backend/pkg/storage"
	"github.com/go-mail/mail"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const maxEmailAttachmentBytes = 8 << 20
const maxEmailAttachments = 5

type MailAttachment struct {
	Name        string
	ContentType string
	Data        []byte
}

type MailJob struct {
	To             []string
	CC             []string
	Subject        string
	Body           string
	IsHTML         bool
	AttachmentIDs  []uuid.UUID
	NotificationID *uuid.UUID
	TemplateCode   string
	UserID         *uuid.UUID
	LogID          uuid.UUID
	Attempts       int
}

type MailDispatcher interface {
	Enqueue(ctx context.Context, job MailJob) (uuid.UUID, error)
}

type MIMESender interface {
	SendMIME(ctx context.Context, job MailJob, files []MailAttachment) error
}

func (c *EmailChannel) WithDispatcher(d MailDispatcher) *EmailChannel {
	c.dispatch = d
	return c
}

func smtpReady(cfg config.EmailConfig) error {
	if cfg.SMTPHost == "" || cfg.SMTPPort <= 0 || strings.TrimSpace(cfg.From) == "" {
		return fmt.Errorf("smtp is not configured")
	}
	// A password must come from the encrypted runtime setting or an environment secret.
	if cfg.Password == "" || cfg.PasswordSource == "" {
		return fmt.Errorf("smtp password is not configured")
	}
	return nil
}

func (c *EmailChannel) SendMIME(ctx context.Context, job MailJob, files []MailAttachment) error {
	cfg, resolveErr := c.resolve(ctx)
	if resolveErr != nil {
		return resolveErr
	}
	if err := smtpReady(cfg); err != nil {
		return err
	}
	m := mail.NewMessage()
	if name := strings.TrimSpace(cfg.FromName); name != "" {
		m.SetAddressHeader("From", cfg.From, name)
	} else {
		m.SetHeader("From", cfg.From)
	}
	m.SetHeader("To", job.To...)
	if len(job.CC) > 0 {
		m.SetHeader("Cc", job.CC...)
	}
	m.SetHeader("Subject", job.Subject)
	if job.IsHTML {
		m.SetBody("text/html", job.Body)
	} else {
		m.SetBody("text/plain", job.Body)
	}
	for _, f := range files {
		name := f.Name
		data := f.Data
		m.AttachReader(name, bytes.NewReader(data))
	}
	d := mail.NewDialer(cfg.SMTPHost, cfg.SMTPPort, cfg.EffectiveUsername(), cfg.Password)
	applySMTPSecurity(d, cfg.EffectiveSSLMode())
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("send email: %w", err)
	}
	return nil
}

func applySMTPSecurity(d *mail.Dialer, mode string) {
	switch mode {
	case config.SSLModeImplicit:
		d.SSL = true
	case config.SSLModeStartTLS:
		d.SSL = false
		d.StartTLSPolicy = mail.MandatoryStartTLS
	case config.SSLModeNone:
		d.SSL = false
		d.StartTLSPolicy = mail.NoStartTLS
	}
}

type attachmentFileRow struct {
	ID           uuid.UUID
	OriginalName string
	Name         string
	Path         string
	MimeType     string
	Size         int64
	// UploadedBy / IsPublic are not selected by the production query; they are
	// kept here so test fakes can model the ownership scope. The production
	// gorm scan simply leaves them as zero values.
	UploadedBy uuid.UUID
	IsPublic   bool
}

// fileRepository abstracts the read of files the operator is allowed to access
// by id. The gorm-backed implementation enforces the ownership scope
// (uploaded_by = operator OR is_public); tests substitute a fake.
type fileRepository interface {
	accessibleFiles(ctx context.Context, operatorID uuid.UUID, ids []uuid.UUID) ([]attachmentFileRow, error)
}

type gormFileRepository struct {
	db *gorm.DB
}

func (g gormFileRepository) accessibleFiles(ctx context.Context, operatorID uuid.UUID, ids []uuid.UUID) ([]attachmentFileRow, error) {
	var rows []attachmentFileRow
	if err := g.db.WithContext(ctx).Table("files").
		Select("id, original_name, name, path, mime_type, size").
		Where("id IN ?", ids).
		Where("(uploaded_by = ? OR is_public = TRUE)", operatorID).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

type attachmentLoader struct {
	repo  fileRepository
	store storage.ObjectStorage
}

func NewAttachmentLoader(db *gorm.DB, store storage.ObjectStorage) *attachmentLoader {
	return &attachmentLoader{repo: gormFileRepository{db: db}, store: store}
}

// Load fetches the given file attachments on behalf of operatorID. Only files
// owned by the operator or marked public are returned. Any id the operator
// cannot access causes the whole request to be rejected (fail-closed), which
// prevents probing whether an arbitrary file id exists.
func (l *attachmentLoader) Load(ctx context.Context, operatorID uuid.UUID, ids []uuid.UUID) ([]MailAttachment, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	if len(ids) > maxEmailAttachments {
		return nil, fmt.Errorf("too many attachments")
	}
	if l.store == nil || l.repo == nil {
		return nil, fmt.Errorf("attachment storage unavailable")
	}
	rows, err := l.repo.accessibleFiles(ctx, operatorID, ids)
	if err != nil {
		return nil, err
	}
	if len(rows) != len(ids) {
		return nil, fmt.Errorf("attachment not found")
	}
	out := make([]MailAttachment, 0, len(rows))
	for _, row := range rows {
		if row.Size > maxEmailAttachmentBytes {
			return nil, fmt.Errorf("attachment too large")
		}
		rc, ctype, err := l.store.Download(ctx, row.Path)
		if err != nil {
			return nil, err
		}
		data, err := io.ReadAll(io.LimitReader(rc, maxEmailAttachmentBytes+1))
		_ = rc.Close()
		if err != nil {
			return nil, err
		}
		if int64(len(data)) > maxEmailAttachmentBytes {
			return nil, fmt.Errorf("attachment too large")
		}
		name := row.OriginalName
		if strings.TrimSpace(name) == "" {
			name = row.Name
		}
		if ctype == "" {
			ctype = row.MimeType
		}
		out = append(out, MailAttachment{Name: name, ContentType: ctype, Data: data})
	}
	return out, nil
}
