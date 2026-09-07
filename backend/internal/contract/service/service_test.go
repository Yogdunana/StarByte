package service

import (
	"context"
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
	})
	require.NoError(t, err)
	assert.Equal(t, "赞助协议", created.Title)
	assert.Equal(t, model.StatusActive, created.Status)

	badEnd := start.Add(-24 * time.Hour)
	_, err = svc.Update(ctx, uuid.MustParse(created.ID), &dto.UpdateContractRequest{ExpiredAt: &badEnd})
	require.Error(t, err)
	assert.Equal(t, response.CodeContractInvalidPeriod, err.(*response.AppError).Code)

	past := time.Now().Add(-24 * time.Hour)
	_, err = svc.Update(ctx, uuid.MustParse(created.ID), &dto.UpdateContractRequest{ExpiredAt: &past})
	require.NoError(t, err)
	got, err := svc.Get(ctx, uuid.MustParse(created.ID))
	require.NoError(t, err)
	assert.Equal(t, model.StatusExpired, got.Status)

	require.NoError(t, svc.Delete(ctx, uuid.MustParse(created.ID)))
	require.NoError(t, svc.ExpiryJob(ctx, "", func(string) {}))

	tpls, err := svc.Templates(ctx)
	require.NoError(t, err)
	assert.Len(t, tpls, 1)
}
