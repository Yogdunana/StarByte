package service

import (
	"context"
	"testing"

	"github.com/Yogdunana/StarByte/backend/internal/configstore/dto"
	"github.com/Yogdunana/StarByte/backend/internal/configstore/model"
	"github.com/Yogdunana/StarByte/backend/pkg/configstore"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestSvc() ConfigService {
	return NewConfigService(newMemRepo(), configstore.New(nil, configstore.NewMemoryBackend()))
}

func TestConfigService_CRUD(t *testing.T) {
	ctx := context.Background()
	svc := newTestSvc()
	op := uuid.New()

	created, err := svc.Create(ctx, op, &dto.CreateConfigRequest{
		ConfigKey:   "site.name",
		ConfigValue: "计协",
		ConfigType:  model.TypeString,
		Category:    "system",
		Description: "协会名称",
	})
	require.NoError(t, err)
	assert.Equal(t, "site.name", created.ConfigKey)

	got, err := svc.GetByKey(ctx, "site.name")
	require.NoError(t, err)
	assert.Equal(t, "计协", got.ConfigValue)

	list, err := svc.List(ctx, dto.ListQuery{Category: "system"})
	require.NoError(t, err)
	assert.Len(t, list, 1)

	val := "计算机协会"
	updated, err := svc.Update(ctx, op, uuid.MustParse(created.ID), &dto.UpdateConfigRequest{ConfigValue: &val})
	require.NoError(t, err)
	assert.Equal(t, "计算机协会", updated.ConfigValue)

	require.NoError(t, svc.Delete(ctx, uuid.MustParse(created.ID)))
	_, err = svc.GetByKey(ctx, "site.name")
	assert.Error(t, err)
}

func TestConfigService_RejectBadKeyAndProtectedDelete(t *testing.T) {
	ctx := context.Background()
	rows := newMemRepo()
	svc := NewConfigService(rows, configstore.New(nil, configstore.NewMemoryBackend()))
	op := uuid.New()

	_, err := svc.Create(ctx, op, &dto.CreateConfigRequest{
		ConfigKey: "Bad Key", ConfigValue: "x", ConfigType: model.TypeString, Category: "system",
	})
	assert.Error(t, err)

	_, err = svc.Create(ctx, op, &dto.CreateConfigRequest{
		ConfigKey: "dup.key", ConfigValue: "1", ConfigType: model.TypeNumber, Category: "business",
	})
	require.NoError(t, err)
	_, err = svc.Create(ctx, op, &dto.CreateConfigRequest{
		ConfigKey: "dup.key", ConfigValue: "2", ConfigType: model.TypeNumber, Category: "business",
	})
	assert.Error(t, err)

	id := uuid.New()
	require.NoError(t, rows.Create(ctx, &model.Config{
		ID: id, ConfigKey: "internship_config", ConfigType: model.TypeJSON,
		ConfigValue: `{"a":1}`, Category: "internship",
	}))
	err = svc.Delete(ctx, id)
	assert.Error(t, err)
}

func TestValidateValue(t *testing.T) {
	assert.NoError(t, validateValue(model.TypeNumber, "3.14"))
	assert.Error(t, validateValue(model.TypeNumber, "x"))
	assert.NoError(t, validateValue(model.TypeBoolean, "true"))
	assert.Error(t, validateValue(model.TypeBoolean, "yes"))
	assert.NoError(t, validateValue(model.TypeJSON, `{"a":1}`))
	assert.Error(t, validateValue(model.TypeJSON, "{"))
}
