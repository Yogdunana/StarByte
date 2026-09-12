package service

import (
	"context"
	"testing"

	notifdto "github.com/Yogdunana/StarByte/backend/internal/notification/dto"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubSender struct {
	n       int
	ctxErr  error
	lastReq *notifdto.SendNotificationRequest
}

func (s *stubSender) Send(ctx context.Context, req *notifdto.SendNotificationRequest) error {
	s.n++
	s.ctxErr = ctx.Err()
	s.lastReq = req
	return nil
}

type stubLookup struct {
	ids []uuid.UUID
}

func (s stubLookup) ListOpsUserIDs(ctx context.Context) ([]uuid.UUID, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return s.ids, nil
}

func TestFailedAlertUsesFreshContextAfterTimeout(t *testing.T) {
	ops := uuid.New()
	sender := &stubSender{}
	a := &notifAlerter{inner: sender, lookup: stubLookup{ids: []uuid.UUID{ops}}}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	require.Error(t, ctx.Err())

	a.Failed(ctx, nil, "db.dump.gz", "context deadline exceeded")
	assert.Equal(t, 1, sender.n)
	assert.NoError(t, sender.ctxErr)
	require.NotNil(t, sender.lastReq)
	assert.Equal(t, tplBackupFailed, sender.lastReq.TemplateCode)
	assert.Equal(t, []uuid.UUID{ops}, sender.lastReq.UserIDs)
}
