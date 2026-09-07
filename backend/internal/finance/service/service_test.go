package service

import (
	"context"
	"testing"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/finance/dto"
	"github.com/Yogdunana/StarByte/backend/internal/finance/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFinanceCRUDAndSummary(t *testing.T) {
	mem := newMem()
	cat := &model.Category{ID: uuid.New(), Name: "会费", Code: "dues", Direction: model.DirectionIncome}
	mem.cats[cat.ID] = cat
	svc := New(mem)
	op := uuid.New()
	ctx := context.Background()

	_, err := svc.Create(ctx, op, &dto.CreateRecordRequest{
		CategoryID: cat.ID.String(), Amount: 0, Direction: 2, OccurredAt: time.Now(), Title: "x",
	})
	require.Error(t, err)
	assert.Equal(t, response.CodeFinanceInvalidAmount, err.(*response.AppError).Code)

	created, err := svc.Create(ctx, op, &dto.CreateRecordRequest{
		CategoryID: cat.ID.String(), Amount: 100, Direction: model.DirectionIncome,
		OccurredAt: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC), Title: "会费",
	})
	require.NoError(t, err)
	assert.Equal(t, "会费", created.Title)

	title := "秋季会费"
	updated, err := svc.Update(ctx, uuid.MustParse(created.ID), &dto.UpdateRecordRequest{Title: &title})
	require.NoError(t, err)
	assert.Equal(t, title, updated.Title)

	list, total, _, _, err := svc.List(ctx, &dto.ListRecordRequest{})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, list, 1)

	sum, err := svc.Summary(ctx, dto.SummaryQuery{})
	require.NoError(t, err)
	assert.Equal(t, 100.0, sum.IncomeTotal)
	assert.Equal(t, 100.0, sum.Balance)

	cats, err := svc.Categories(ctx)
	require.NoError(t, err)
	assert.Len(t, cats, 1)

	require.NoError(t, svc.Delete(ctx, uuid.MustParse(created.ID)))
	_, err = svc.Get(ctx, uuid.MustParse(created.ID))
	require.Error(t, err)
	assert.Equal(t, response.CodeFinanceNotFound, err.(*response.AppError).Code)
}
