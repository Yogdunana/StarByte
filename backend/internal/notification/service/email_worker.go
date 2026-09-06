package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/notification/model"
	"github.com/Yogdunana/StarByte/backend/internal/notification/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var errEmailQueueFull = errors.New("email queue is full")

const emailQueueSize = 64
const emailMaxAttempts = 3

var defaultEmailBackoff = []time.Duration{time.Minute, 5 * time.Minute, 15 * time.Minute}

type EmailWorker struct {
	sender  MIMESender
	logs    repo.EmailLogRepo
	attach  *attachmentLoader
	limiter *minuteLimiter
	backoff []time.Duration
	sleep   func(time.Duration)
	jobs    chan MailJob
	now     func() time.Time
}

func NewEmailWorker(sender MIMESender, logs repo.EmailLogRepo, attach *attachmentLoader, limiter *minuteLimiter) *EmailWorker {
	return &EmailWorker{
		sender:  sender,
		logs:    logs,
		attach:  attach,
		limiter: limiter,
		backoff: defaultEmailBackoff,
		sleep:   time.Sleep,
		jobs:    make(chan MailJob, emailQueueSize),
		now:     time.Now,
	}
}

func (w *EmailWorker) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case job := <-w.jobs:
				w.process(ctx, job)
			}
		}
	}()
}

func (w *EmailWorker) Enqueue(ctx context.Context, job MailJob) (uuid.UUID, error) {
	if job.LogID == uuid.Nil {
		row := w.newLog(job)
		if err := w.logs.Create(ctx, row); err != nil {
			return uuid.Nil, err
		}
		job.LogID = row.ID
	}
	select {
	case w.jobs <- job:
		return job.LogID, nil
	default:
		_ = w.mark(ctx, job.LogID, model.EmailFailed, "queue is full", 0)
		return job.LogID, errEmailQueueFull
	}
}

func (w *EmailWorker) process(ctx context.Context, job MailJob) {
	var files []MailAttachment
	var err error
	if len(job.AttachmentIDs) > 0 && w.attach != nil {
		files, err = w.attach.Load(ctx, job.AttachmentIDs)
		if err != nil {
			_ = w.mark(ctx, job.LogID, model.EmailFailed, err.Error(), 0)
			return
		}
	}
	for attempt := 1; attempt <= emailMaxAttempts; attempt++ {
		for w.limiter != nil && !w.limiter.Allow() {
			w.sleep(200 * time.Millisecond)
			if ctx.Err() != nil {
				return
			}
		}
		err = w.sender.SendMIME(ctx, job, files)
		if err == nil {
			_ = w.mark(ctx, job.LogID, model.EmailSent, "", int16(attempt-1))
			return
		}
		status := model.EmailRetrying
		if attempt == emailMaxAttempts {
			status = model.EmailFailed
		}
		_ = w.mark(ctx, job.LogID, status, err.Error(), int16(attempt))
		if attempt == emailMaxAttempts {
			return
		}
		wait := w.backoff[attempt-1]
		if wait > 0 {
			w.sleep(wait)
		}
	}
}

func (w *EmailWorker) newLog(job MailJob) *model.EmailLog {
	return &model.EmailLog{
		ID:             uuid.New(),
		UserID:         job.UserID,
		ToAddress:      strings.Join(job.To, ","),
		Subject:        job.Subject,
		TemplateCode:   job.TemplateCode,
		Status:         model.EmailQueued,
		NotificationID: job.NotificationID,
		CC:             strings.Join(job.CC, ","),
		IsHTML:         job.IsHTML,
		CreatedAt:      w.now(),
	}
}

func (w *EmailWorker) mark(ctx context.Context, id uuid.UUID, status int16, errMsg string, retries int16) error {
	row, err := w.logs.GetByID(ctx, id)
	if err != nil {
		logger.Error("email log update skipped", zap.Error(err))
		return err
	}
	row.Status = status
	row.ErrorMessage = errMsg
	row.RetryCount = retries
	if status == model.EmailSent {
		now := w.now()
		row.SentAt = &now
	}
	return w.logs.Update(ctx, row)
}
