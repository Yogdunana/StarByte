package service

import (
	"context"
	"testing"

	"github.com/Yogdunana/StarByte/backend/internal/dict/dto"
	"github.com/Yogdunana/StarByte/backend/internal/dict/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func newTestSvc() (DictService, *memRepo, *memCache) {
	repo := newMemRepo()
	cache := newMemCache()
	return NewDictService(repo, cache), repo, cache
}

func TestTypeAndItemCRUD(t *testing.T) {
	svc, _, _ := newTestSvc()
	ctx := context.Background()

	created, err := svc.CreateType(ctx, dto.CreateTypeRequest{Code: "color", Name: "颜色", SortOrder: 1})
	require.NoError(t, err)
	require.Equal(t, "color", created.Code)
	require.False(t, created.IsSystem)

	types, err := svc.ListTypes(ctx)
	require.NoError(t, err)
	require.Len(t, types, 1)

	name := "色彩"
	updated, err := svc.UpdateType(ctx, uuid.MustParse(created.ID), dto.UpdateTypeRequest{Name: &name})
	require.NoError(t, err)
	require.Equal(t, "色彩", updated.Name)

	item, err := svc.CreateItem(ctx, dto.CreateItemRequest{TypeCode: "color", ItemValue: "red", ItemLabel: "红"})
	require.NoError(t, err)
	require.Equal(t, "red", item.ItemValue)

	items, err := svc.ListItems(ctx, "color", true)
	require.NoError(t, err)
	require.Len(t, items, 1)

	label := "红色"
	item, err = svc.UpdateItem(ctx, uuid.MustParse(item.ID), dto.UpdateItemRequest{ItemLabel: &label})
	require.NoError(t, err)
	require.Equal(t, "红色", item.ItemLabel)

	require.NoError(t, svc.DeleteItem(ctx, uuid.MustParse(item.ID)))
	items, err = svc.ListItems(ctx, "color", true)
	require.NoError(t, err)
	require.Empty(t, items)

	require.NoError(t, svc.DeleteType(ctx, uuid.MustParse(created.ID)))
	types, err = svc.ListTypes(ctx)
	require.NoError(t, err)
	require.Empty(t, types)
}

func TestDuplicateAndNotFound(t *testing.T) {
	svc, _, _ := newTestSvc()
	ctx := context.Background()
	_, err := svc.CreateType(ctx, dto.CreateTypeRequest{Code: "a", Name: "A"})
	require.NoError(t, err)
	_, err = svc.CreateType(ctx, dto.CreateTypeRequest{Code: "a", Name: "B"})
	require.Error(t, err)
	require.Equal(t, response.CodeDictTypeExists, err.(*response.AppError).Code)

	_, err = svc.CreateItem(ctx, dto.CreateItemRequest{TypeCode: "missing", ItemValue: "1", ItemLabel: "一"})
	require.Error(t, err)
	require.Equal(t, response.CodeDictTypeNotFound, err.(*response.AppError).Code)

	_, err = svc.CreateItem(ctx, dto.CreateItemRequest{TypeCode: "a", ItemValue: "1", ItemLabel: "一"})
	require.NoError(t, err)
	_, err = svc.CreateItem(ctx, dto.CreateItemRequest{TypeCode: "a", ItemValue: "1", ItemLabel: "壹"})
	require.Error(t, err)
	require.Equal(t, response.CodeDictItemExists, err.(*response.AppError).Code)

	_, err = svc.UpdateItem(ctx, uuid.New(), dto.UpdateItemRequest{})
	require.Error(t, err)
	require.Equal(t, response.CodeDictItemNotFound, err.(*response.AppError).Code)
}

func TestSystemTypeCannotDelete(t *testing.T) {
	svc, repo, _ := newTestSvc()
	ctx := context.Background()
	id := uuid.New()
	require.NoError(t, repo.CreateType(ctx, &model.DictType{ID: id, Code: "task_status", Name: "任务状态", IsSystem: true}))
	err := svc.DeleteType(ctx, id)
	require.Error(t, err)
	require.Equal(t, response.CodeDictSystemLocked, err.(*response.AppError).Code)
}

func TestEnabledOnlyAndCacheInvalidate(t *testing.T) {
	svc, _, cache := newTestSvc()
	ctx := context.Background()
	_, err := svc.CreateType(ctx, dto.CreateTypeRequest{Code: "prio", Name: "优先级"})
	require.NoError(t, err)
	_, err = svc.CreateItem(ctx, dto.CreateItemRequest{TypeCode: "prio", ItemValue: "1", ItemLabel: "高"})
	require.NoError(t, err)
	disabled := int16(model.StatusDisabled)
	hidden, err := svc.CreateItem(ctx, dto.CreateItemRequest{TypeCode: "prio", ItemValue: "0", ItemLabel: "隐藏"})
	require.NoError(t, err)
	_, err = svc.UpdateItem(ctx, uuid.MustParse(hidden.ID), dto.UpdateItemRequest{Status: &disabled})
	require.NoError(t, err)

	enabled, err := svc.ListItems(ctx, "prio", true)
	require.NoError(t, err)
	require.Len(t, enabled, 1)
	require.Equal(t, "1", enabled[0].ItemValue)

	all, err := svc.ListItems(ctx, "prio", false)
	require.NoError(t, err)
	require.Len(t, all, 2)

	cached, err := cache.Get(ctx, cacheKey("prio"))
	require.NoError(t, err)
	require.Contains(t, cached, `"item_value":"1"`)

	_, err = svc.CreateItem(ctx, dto.CreateItemRequest{TypeCode: "prio", ItemValue: "2", ItemLabel: "紧急"})
	require.NoError(t, err)
	_, err = cache.Get(ctx, cacheKey("prio"))
	require.Error(t, err)
}
