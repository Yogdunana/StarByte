package service

import (
	"context"
	"testing"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/finance/dto"
	"github.com/Yogdunana/StarByte/backend/internal/finance/model"
	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
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
	}, nil)
	require.Error(t, err)
	assert.Equal(t, response.CodeFinanceInvalidAmount, err.(*response.AppError).Code)

	_, err = svc.Create(ctx, op, &dto.CreateRecordRequest{
		CategoryID: cat.ID.String(), Amount: 10, Direction: model.DirectionExpense,
		OccurredAt: time.Now(), Title: "错向",
	}, nil)
	require.Error(t, err)
	assert.Equal(t, response.CodeFinanceDirectionMismatch, err.(*response.AppError).Code)

	created, err := svc.Create(ctx, op, &dto.CreateRecordRequest{
		CategoryID: cat.ID.String(), Amount: 100, Direction: model.DirectionIncome,
		OccurredAt: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC), Title: "会费",
	}, nil)
	require.NoError(t, err)
	assert.Equal(t, "会费", created.Title)

	title := "秋季会费"
	updated, err := svc.Update(ctx, op, uuid.MustParse(created.ID), &dto.UpdateRecordRequest{Title: &title}, nil)
	require.NoError(t, err)
	assert.Equal(t, title, updated.Title)

	list, total, _, _, err := svc.List(ctx, op, &dto.ListRecordRequest{}, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, list, 1)

	sum, err := svc.Summary(ctx, op, dto.SummaryQuery{}, nil)
	require.NoError(t, err)
	assert.Equal(t, 100.0, sum.IncomeTotal)
	assert.Equal(t, 100.0, sum.Balance)

	cats, err := svc.Categories(ctx)
	require.NoError(t, err)
	assert.Len(t, cats, 1)

	require.NoError(t, svc.Delete(ctx, op, uuid.MustParse(created.ID), nil))
	_, err = svc.Get(ctx, op, uuid.MustParse(created.ID), nil)
	require.Error(t, err)
	assert.Equal(t, response.CodeFinanceNotFound, err.(*response.AppError).Code)
}

func TestDateOnlyUsesShanghaiCalendar(t *testing.T) {
	// 2026-09-07 00:00 GMT+8 serializes to 2026-09-06T16:00:00Z
	got := dateOnly(time.Date(2026, 9, 6, 16, 0, 0, 0, time.UTC))
	assert.Equal(t, 2026, got.Year())
	assert.Equal(t, time.September, got.Month())
	assert.Equal(t, 7, got.Day())
}

func TestFinanceSelfCannotTagForeignDept(t *testing.T) {
	mem := newMem()
	cat := &model.Category{ID: uuid.New(), Name: "会费", Code: "dues", Direction: model.DirectionIncome}
	mem.cats[cat.ID] = cat
	svc := New(mem)
	op := uuid.New()
	self := &rbacModel.DataScopeCondition{Query: "1 = 0", IsSelf: true}
	other := uuid.New()
	_, err := svc.Create(context.Background(), op, &dto.CreateRecordRequest{
		CategoryID: cat.ID.String(), Amount: 10, Direction: model.DirectionIncome,
		OccurredAt: time.Now(), Title: "会费", DepartmentID: other.String(),
	}, self)
	require.Error(t, err)
	assert.Equal(t, response.CodeFinanceNoAccess, err.(*response.AppError).Code)
}
