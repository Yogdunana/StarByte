package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/contract/dto"
	"github.com/Yogdunana/StarByte/backend/internal/contract/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContractCRUDAndExpiry(t *testing.T) {
	mem := newMem()
	tpl := &model.Template{ID: uuid.New(), Name: "赞助模板", Code: "sponsor", Content: "条款"}
	mem.tpls[tpl.ID] = tpl
	svc := New(mem, nil)
	ctx := context.Background()
	op := uuid.New()
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	st := model.StatusActive

	created, err := svc.Create(ctx, op, &dto.CreateContractRequest{
		Title: "赞助协议", ContractType: 1, PartyName: "某公司",
		TemplateID: tpl.ID.String(), StartAt: &start, ExpiredAt: &end, Status: &st,
	}, nil)
	require.NoError(t, err)
	assert.Equal(t, "赞助协议", created.Title)
	assert.Equal(t, model.StatusDraft, created.Status)

	created, err = svc.Update(ctx, op, uuid.MustParse(created.ID), &dto.UpdateContractRequest{Status: &st}, nil)
	require.NoError(t, err)
	assert.Equal(t, model.StatusActive, created.Status)

	badEnd := start.Add(-24 * time.Hour)
	_, err = svc.Update(ctx, op, uuid.MustParse(created.ID), &dto.UpdateContractRequest{ExpiredAt: &badEnd}, nil)
	require.Error(t, err)
	assert.Equal(t, response.CodeContractInvalidPeriod, err.(*response.AppError).Code)

	past := time.Now().Add(-24 * time.Hour)
	_, err = svc.Update(ctx, op, uuid.MustParse(created.ID), &dto.UpdateContractRequest{ExpiredAt: &past}, nil)
	require.NoError(t, err)
	got, err := svc.Get(ctx, op, uuid.MustParse(created.ID), nil)
	require.NoError(t, err)
	assert.Equal(t, model.StatusExpired, got.Status)

	require.NoError(t, svc.Delete(ctx, op, uuid.MustParse(created.ID), nil))
	require.NoError(t, svc.ExpiryJob(ctx, "", func(string) {}))

	tpls, err := svc.Templates(ctx)
	require.NoError(t, err)
	assert.Len(t, tpls, 1)
}

type captureNotify struct{ n int }

func (c *captureNotify) Send(_ context.Context, _ []uuid.UUID, _ string, _ map[string]interface{}) error {
	c.n++
	return nil
}

func TestExpiryJobNotifiesOnce(t *testing.T) {
	mem := newMem()
	n := &captureNotify{}
	svc := New(mem, n)
	ctx := context.Background()
	op := uuid.New()
	st := model.StatusActive
	exp := time.Now().Add(3 * 24 * time.Hour)
	created, err := svc.Create(ctx, op, &dto.CreateContractRequest{
		Title: "临期合同", ContractType: 1, PartyName: "某公司",
		ExpiredAt: &exp, Status: &st,
	}, nil)
	require.NoError(t, err)
	_, err = svc.Update(ctx, op, uuid.MustParse(created.ID), &dto.UpdateContractRequest{Status: &st}, nil)
	require.NoError(t, err)
	require.NoError(t, svc.ExpiryJob(ctx, "", func(string) {}))
	assert.Equal(t, 1, n.n)
	require.NoError(t, svc.ExpiryJob(ctx, "", func(string) {}))
	assert.Equal(t, 1, n.n)
	got, err := svc.Get(ctx, op, uuid.MustParse(created.ID), nil)
	require.NoError(t, err)
	assert.Equal(t, model.StatusActive, got.Status)
}

type failNotify struct{ n int }

func (f *failNotify) Send(_ context.Context, _ []uuid.UUID, _ string, _ map[string]interface{}) error {
	f.n++
	return errors.New("notify down")
}

func TestExpiryJobRetriesWhenNotifyFails(t *testing.T) {
	mem := newMem()
	n := &failNotify{}
	svc := New(mem, n)
	ctx := context.Background()
	op := uuid.New()
	st := model.StatusActive
	exp := time.Now().Add(3 * 24 * time.Hour)
	created, err := svc.Create(ctx, op, &dto.CreateContractRequest{
		Title: "临期合同", ContractType: 1, PartyName: "某公司",
		ExpiredAt: &exp, Status: &st,
	}, nil)
	require.NoError(t, err)
	_, err = svc.Update(ctx, op, uuid.MustParse(created.ID), &dto.UpdateContractRequest{Status: &st}, nil)
	require.NoError(t, err)
	require.NoError(t, svc.ExpiryJob(ctx, "", func(string) {}))
	assert.Equal(t, 1, n.n)
	row, err := mem.GetByID(ctx, uuid.MustParse(created.ID))
	require.NoError(t, err)
	require.NotNil(t, row)
	assert.Nil(t, row.ExpiryNotifiedAt)
	require.NoError(t, svc.ExpiryJob(ctx, "", func(string) {}))
	assert.Equal(t, 2, n.n)
}

func TestTodayExpiryRemainsActive(t *testing.T) {
	mem := newMem()
	svc := New(mem, nil)
	ctx := context.Background()
	op := uuid.New()
	st := model.StatusActive
	today := dateOnly(time.Now())
	created, err := svc.Create(ctx, op, &dto.CreateContractRequest{
		Title: "今日到期", ContractType: 1, PartyName: "某公司",
		ExpiredAt: &today, Status: &st,
	}, nil)
	require.NoError(t, err)
	_, err = svc.Update(ctx, op, uuid.MustParse(created.ID), &dto.UpdateContractRequest{Status: &st}, nil)
	require.NoError(t, err)
	got, err := svc.Get(ctx, op, uuid.MustParse(created.ID), nil)
	require.NoError(t, err)
	assert.Equal(t, model.StatusActive, got.Status)
	require.NoError(t, svc.ExpiryJob(ctx, "", func(string) {}))
	got, err = svc.Get(ctx, op, uuid.MustParse(created.ID), nil)
	require.NoError(t, err)
	assert.Equal(t, model.StatusActive, got.Status)
}
