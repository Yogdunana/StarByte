package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/configstore/model"
	"github.com/google/uuid"
)

// BackendAdapter 让 pkg/configstore 读写已有 configs 表。
type BackendAdapter struct {
	Rows ConfigRepo
}

func (a *BackendAdapter) Load(ctx context.Context, key string) (string, bool, error) {
	row, err := a.Rows.GetByKey(ctx, key)
	if err != nil {
		return "", false, err
	}
	if row == nil {
		return "", false, nil
	}
	return row.ConfigValue, true, nil
}

func (a *BackendAdapter) Save(ctx context.Context, key, value string) error {
	row, err := a.Rows.GetByKey(ctx, key)
	if err != nil {
		return err
	}
	now := time.Now()
	if row == nil {
		row = &model.Config{
			ID:          uuid.New(),
			ConfigKey:   key,
			ConfigType:  model.TypeString,
			Category:    "system",
			CreatedAt:   now,
			UpdatedAt:   now,
			ConfigValue: value,
		}
		if err := a.Rows.Create(ctx, row); err != nil {
			return fmt.Errorf("create config: %w", err)
		}
		return nil
	}
	row.ConfigValue = value
	row.UpdatedAt = now
	return a.Rows.Update(ctx, row)
}
