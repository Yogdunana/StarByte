package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

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

func (c *EmailChannel) SendMIME(_ context.Context, job MailJob, files []MailAttachment) error {
	if !c.IsAvailable() {
		return fmt.Errorf("smtp is not configured")
	}
	m := mail.NewMessage()
	m.SetHeader("From", c.from)
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
	d := mail.NewDialer(c.smtpHost, c.smtpPort, c.username, c.password)
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("send email: %w", err)
	}
	return nil
}

type attachmentLoader struct {
	db    *gorm.DB
	store storage.ObjectStorage
}

func NewAttachmentLoader(db *gorm.DB, store storage.ObjectStorage) *attachmentLoader {
	return &attachmentLoader{db: db, store: store}
}

func (l *attachmentLoader) Load(ctx context.Context, ids []uuid.UUID) ([]MailAttachment, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	if len(ids) > maxEmailAttachments {
		return nil, fmt.Errorf("too many attachments")
	}
	if l.store == nil || l.db == nil {
		return nil, fmt.Errorf("attachment storage unavailable")
	}
	var rows []struct {
		ID           uuid.UUID
		OriginalName string
		Name         string
		Path         string
		MimeType     string
		Size         int64
	}
	if err := l.db.WithContext(ctx).Table("files").Select("id, original_name, name, path, mime_type, size").
		Where("id IN ?", ids).Find(&rows).Error; err != nil {
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
