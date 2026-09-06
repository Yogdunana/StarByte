package service

import (
	"context"
	"errors"
	"strings"
	"sync"
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
const rateRetryDelay = 200 * time.Millisecond

var defaultEmailBackoff = []time.Duration{time.Minute, 5 * time.Minute, 15 * time.Minute}

type delayedJob struct {
	job MailJob
	at  time.Time
}

type EmailWorker struct {
	sender  MIMESender
	logs    repo.EmailLogRepo
	attach  *attachmentLoader
	limiter *minuteLimiter
	backoff []time.Duration
	jobs    chan MailJob
	now     func() time.Time
	delayMu sync.Mutex
	delayed []delayedJob
}

func NewEmailWorker(sender MIMESender, logs repo.EmailLogRepo, attach *attachmentLoader, limiter *minuteLimiter) *EmailWorker {
	return &EmailWorker{
		sender:  sender,
		logs:    logs,
		attach:  attach,
		limiter: limiter,
		backoff: defaultEmailBackoff,
		jobs:    make(chan MailJob, emailQueueSize),
		now:     time.Now,
	}
}

func (w *EmailWorker) Start(ctx context.Context) {
	go w.loop(ctx)
	go w.delayLoop(ctx)
}

func (w *EmailWorker) loop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case job := <-w.jobs:
			w.process(ctx, job)
		}
	}
}

func (w *EmailWorker) delayLoop(ctx context.Context) {
	tick := time.NewTicker(rateRetryDelay)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			w.flushDue()
		}
	}
}

func (w *EmailWorker) Enqueue(ctx context.Context, job MailJob) (uuid.UUID, error) {
	if job.LogID == uuid.Nil {
		row := w.newLog(job)
		if err := w.logs.Create(ctx, row); err != nil {
			return uuid.Nil, err
		}
		job.LogID = row.ID
	}
	if err := w.push(ctx, job); err != nil {
		return job.LogID, err
	}
	return job.LogID, nil
}

func (w *EmailWorker) push(ctx context.Context, job MailJob) error {
	select {
	case w.jobs <- job:
		return nil
	default:
		_ = w.mark(ctx, job.LogID, model.EmailFailed, "queue is full", int16(job.Attempts))
		return errEmailQueueFull
	}
}

func (w *EmailWorker) schedule(job MailJob, d time.Duration) {
	if d <= 0 {
		_ = w.push(context.Background(), job)
		return
	}
	w.delayMu.Lock()
	w.delayed = append(w.delayed, delayedJob{job: job, at: w.now().Add(d)})
	w.delayMu.Unlock()
}

func (w *EmailWorker) flushDue() {
	now := w.now()
	w.delayMu.Lock()
	keep := w.delayed[:0]
	due := make([]MailJob, 0)
	for _, item := range w.delayed {
		if !item.at.After(now) {
			due = append(due, item.job)
		} else {
			keep = append(keep, item)
		}
	}
	w.delayed = keep
	w.delayMu.Unlock()
	for _, job := range due {
		_ = w.push(context.Background(), job)
	}
}

func (w *EmailWorker) process(ctx context.Context, job MailJob) {
	if w.limiter != nil && !w.limiter.Allow() {
		w.schedule(job, rateRetryDelay)
		return
	}
	var files []MailAttachment
	var err error
	if len(job.AttachmentIDs) > 0 && w.attach != nil {
		files, err = w.attach.Load(ctx, job.AttachmentIDs)
		if err != nil {
			_ = w.mark(ctx, job.LogID, model.EmailFailed, err.Error(), int16(job.Attempts))
			return
		}
	}
	err = w.sender.SendMIME(ctx, job, files)
	if err == nil {
		_ = w.mark(ctx, job.LogID, model.EmailSent, "", int16(job.Attempts))
		return
	}
	job.Attempts++
	if job.Attempts >= emailMaxAttempts {
		_ = w.mark(ctx, job.LogID, model.EmailFailed, err.Error(), int16(job.Attempts))
		return
	}
	_ = w.mark(ctx, job.LogID, model.EmailRetrying, err.Error(), int16(job.Attempts))
	wait := rateRetryDelay
	if i := job.Attempts - 1; i >= 0 && i < len(w.backoff) {
		wait = w.backoff[i]
	}
	w.schedule(job, wait)
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
