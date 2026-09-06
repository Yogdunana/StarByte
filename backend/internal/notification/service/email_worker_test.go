package service

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/notification/dto"
	"github.com/Yogdunana/StarByte/backend/internal/notification/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubMIME struct {
	mu    sync.Mutex
	fail  int
	calls int
	last  MailJob
}

func (s *stubMIME) SendMIME(_ context.Context, job MailJob, _ []MailAttachment) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls++
	s.last = job
	if s.calls <= s.fail {
		return fmt.Errorf("smtp down")
	}
	return nil
}

type memLogs struct {
	mu   sync.Mutex
	rows map[uuid.UUID]*model.EmailLog
}

func newMemLogs() *memLogs { return &memLogs{rows: map[uuid.UUID]*model.EmailLog{}} }

func (m *memLogs) Create(_ context.Context, row *model.EmailLog) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *row
	m.rows[row.ID] = &cp
	return nil
}
func (m *memLogs) Update(_ context.Context, row *model.EmailLog) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *row
	m.rows[row.ID] = &cp
	return nil
}
func (m *memLogs) GetByID(_ context.Context, id uuid.UUID) (*model.EmailLog, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	row, ok := m.rows[id]
	if !ok {
		return nil, fmt.Errorf("missing")
	}
	cp := *row
	return &cp, nil
}
func (m *memLogs) List(_ context.Context, status int16, _, _ *time.Time, _, _ int) ([]*model.EmailLog, int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := []*model.EmailLog{}
	for _, r := range m.rows {
		if status < 0 || r.Status == status {
			cp := *r
			out = append(out, &cp)
		}
	}
	return out, int64(len(out)), nil
}

func TestMinuteLimiterCapsAt50(t *testing.T) {
	now := time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)
	l := newMinuteLimiter(50, func() time.Time { return now })
	for i := 0; i < 50; i++ {
		require.True(t, l.Allow())
	}
	assert.False(t, l.Allow())
	now = now.Add(time.Minute + time.Second)
	assert.True(t, l.Allow())
}

func processQueued(t *testing.T, w *EmailWorker, ctx context.Context) {
	t.Helper()
	select {
	case job := <-w.jobs:
		w.process(ctx, job)
	default:
		t.Fatal("email queue empty")
	}
}

func TestEmailWorkerRetriesThenFails(t *testing.T) {
	sender := &stubMIME{fail: 5}
	logs := newMemLogs()
	w := NewEmailWorker(sender, logs, nil, newMinuteLimiter(50, time.Now))
	w.backoff = []time.Duration{0, 0, 0}
	ctx := context.Background()
	id, err := w.Enqueue(ctx, MailJob{To: []string{"a@b.c"}, Subject: "s", Body: "b"})
	require.NoError(t, err)
	for i := 0; i < emailMaxAttempts; i++ {
		processQueued(t, w, ctx)
	}
	row, err := logs.GetByID(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, model.EmailFailed, row.Status)
	assert.Equal(t, int16(3), row.RetryCount)
	assert.Equal(t, 3, sender.calls)
}

func TestEmailWorkerSucceedsAfterRetry(t *testing.T) {
	sender := &stubMIME{fail: 1}
	logs := newMemLogs()
	w := NewEmailWorker(sender, logs, nil, newMinuteLimiter(50, time.Now))
	w.backoff = []time.Duration{0, 0, 0}
	ctx := context.Background()
	id, err := w.Enqueue(ctx, MailJob{To: []string{"a@b.c"}, Subject: "s", Body: "b"})
	require.NoError(t, err)
	processQueued(t, w, ctx)
	processQueued(t, w, ctx)
	row, err := logs.GetByID(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, model.EmailSent, row.Status)
	assert.Equal(t, 2, sender.calls)
}

type failToMIME struct {
	mu    sync.Mutex
	fail  map[string]bool
	order []string
}

func (s *failToMIME) SendMIME(_ context.Context, job MailJob, _ []MailAttachment) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	to := ""
	if len(job.To) > 0 {
		to = job.To[0]
	}
	s.order = append(s.order, to)
	if s.fail[to] {
		return fmt.Errorf("smtp down")
	}
	return nil
}

func TestEmailWorkerDoesNotBlockOnRetryBackoff(t *testing.T) {
	sender := &failToMIME{fail: map[string]bool{"fail@x.test": true}}
	logs := newMemLogs()
	w := NewEmailWorker(sender, logs, nil, newMinuteLimiter(50, time.Now))
	w.backoff = []time.Duration{time.Hour, time.Hour, time.Hour}
	ctx := context.Background()
	failID, err := w.Enqueue(ctx, MailJob{To: []string{"fail@x.test"}, Subject: "s", Body: "b"})
	require.NoError(t, err)
	okID, err := w.Enqueue(ctx, MailJob{To: []string{"ok@x.test"}, Subject: "s", Body: "b"})
	require.NoError(t, err)
	processQueued(t, w, ctx)
	processQueued(t, w, ctx)
	assert.Equal(t, []string{"fail@x.test", "ok@x.test"}, sender.order)
	failed, err := logs.GetByID(ctx, failID)
	require.NoError(t, err)
	ok, err := logs.GetByID(ctx, okID)
	require.NoError(t, err)
	assert.Equal(t, model.EmailRetrying, failed.Status)
	assert.Equal(t, model.EmailSent, ok.Status)
	assert.Len(t, w.delayed, 1)
}

func TestEmailWorkerQueueBoundIncludesDelayed(t *testing.T) {
	now := time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)
	l := newMinuteLimiter(50, func() time.Time { return now })
	for i := 0; i < 50; i++ {
		require.True(t, l.Allow())
	}
	logs := newMemLogs()
	w := NewEmailWorker(&stubMIME{}, logs, nil, l)
	ctx := context.Background()
	for i := 0; i < emailQueueSize; i++ {
		_, err := w.Enqueue(ctx, MailJob{To: []string{"a@b.c"}, Subject: "s", Body: "b"})
		require.NoError(t, err)
	}
	for i := 0; i < emailQueueSize; i++ {
		processQueued(t, w, ctx)
	}
	_, err := w.Enqueue(ctx, MailJob{To: []string{"overflow@x.test"}, Subject: "s", Body: "b"})
	require.ErrorIs(t, err, errEmailQueueFull)
	assert.Equal(t, emailQueueSize, w.queued)
	assert.Len(t, w.delayed, emailQueueSize)
}

func TestFlushDueKeepsRetryWhenQueueFull(t *testing.T) {
	logs := newMemLogs()
	w := NewEmailWorker(&stubMIME{}, logs, nil, newMinuteLimiter(50, time.Now))
	ctx := context.Background()
	for i := 0; i < emailQueueSize; i++ {
		_, err := w.Enqueue(ctx, MailJob{To: []string{"a@b.c"}, Subject: "s", Body: "b"})
		require.NoError(t, err)
	}
	extra := MailJob{To: []string{"retry@x.test"}, Subject: "s", Body: "b", Attempts: 1}
	row := w.newLog(extra)
	row.Status = model.EmailRetrying
	require.NoError(t, logs.Create(ctx, row))
	extra.LogID = row.ID
	w.delayed = []delayedJob{{job: extra, at: w.now()}}
	w.flushDue()
	got, err := logs.GetByID(ctx, row.ID)
	require.NoError(t, err)
	assert.Equal(t, model.EmailRetrying, got.Status)
	assert.NotEqual(t, "queue is full", got.ErrorMessage)
	assert.Len(t, w.delayed, 1)
}

func TestEmailWorkerStoresLongRecipientList(t *testing.T) {
	logs := newMemLogs()
	w := NewEmailWorker(&stubMIME{}, logs, nil, newMinuteLimiter(50, time.Now))
	to := make([]string, 20)
	for i := range to {
		to[i] = fmt.Sprintf("user%02d@example.test", i)
	}
	ctx := context.Background()
	id, err := w.Enqueue(ctx, MailJob{To: to, Subject: "s", Body: "b"})
	require.NoError(t, err)
	row, err := logs.GetByID(ctx, id)
	require.NoError(t, err)
	assert.Greater(t, len(row.ToAddress), 200)
	assert.Contains(t, row.ToAddress, "user19@example.test")
}

type stubEngine struct{}

func (stubEngine) Render(context.Context, string, map[string]interface{}) (*dto.TestTemplateResponse, error) {
	return &dto.TestTemplateResponse{Title: "T", Content: "C"}, nil
}
func (stubEngine) RenderTemplate(*model.NotificationTemplate, map[string]interface{}) (*dto.TestTemplateResponse, error) {
	return &dto.TestTemplateResponse{Title: "T", Content: "C"}, nil
}
func (stubEngine) Validate(context.Context, string, map[string]interface{}) error { return nil }

func TestEmailServiceBatchLimit(t *testing.T) {
	logs := newMemLogs()
	w := NewEmailWorker(&stubMIME{}, logs, nil, newMinuteLimiter(50, time.Now))
	svc := NewEmailService(w, stubEngine{}, logs)
	recs := make([]dto.BatchRecipient, 51)
	for i := range recs {
		recs[i] = dto.BatchRecipient{Email: "a@b.c"}
	}
	_, err := svc.SendBatch(context.Background(), &dto.BatchSendEmailRequest{Recipients: recs, Subject: "s", Body: "b"}, uuid.New())
	require.Error(t, err)
}

func TestEmailServiceSendQueues(t *testing.T) {
	logs := newMemLogs()
	w := NewEmailWorker(&stubMIME{}, logs, nil, newMinuteLimiter(50, time.Now))
	svc := NewEmailService(w, stubEngine{}, logs)
	res, err := svc.Send(context.Background(), &dto.SendEmailRequest{
		To: []string{"a@b.c"}, Subject: "hello", Body: "world",
	}, uuid.New())
	require.NoError(t, err)
	assert.Equal(t, "queued", res.Status)
	assert.NotEmpty(t, res.MessageID)
}
