package repo

import (
	"context"
	"encoding/json"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// VariableRepo manages flow_variables for runtime process data.
type VariableRepo interface {
	// Set upserts a variable (create or update by instance_id + key + scope).
	Set(ctx context.Context, tx *gorm.DB, v *model.FlowVariable) error
	// Get retrieves a variable by instance_id and key.
	Get(ctx context.Context, instanceID uuid.UUID, key string) (*model.FlowVariable, error)
	// ListByInstance returns all variables for an instance.
	ListByInstance(ctx context.Context, instanceID uuid.UUID) ([]model.FlowVariable, error)
	// GetMap returns all variables for an instance as a map[string]interface{}.
	GetMap(ctx context.Context, instanceID uuid.UUID) (map[string]interface{}, error)
	// SetMap batch-upserts multiple variables for an instance.
	SetMap(ctx context.Context, tx *gorm.DB, instanceID uuid.UUID, vars map[string]interface{}) error
	// DeleteByInstance removes all variables for an instance.
	DeleteByInstance(ctx context.Context, instanceID uuid.UUID) error
}

type variableRepo struct {
	db *gorm.DB
}

// NewVariableRepo creates a VariableRepo backed by the given GORM DB.
func NewVariableRepo(db *gorm.DB) VariableRepo {
	return &variableRepo{db: db}
}

func (r *variableRepo) getDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.db
}

func (r *variableRepo) Set(ctx context.Context, tx *gorm.DB, v *model.FlowVariable) error {
	// Upsert: try to find existing, update or create.
	var existing model.FlowVariable
	err := r.getDB(tx).WithContext(ctx).
		Where("instance_id = ? AND key = ? AND scope = ?", v.InstanceID, v.Key, v.Scope).
		First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		return r.getDB(tx).WithContext(ctx).Create(v).Error
	}
	if err != nil {
		return err
	}

	existing.Value = v.Value
	return r.getDB(tx).WithContext(ctx).Save(&existing).Error
}

func (r *variableRepo) Get(ctx context.Context, instanceID uuid.UUID, key string) (*model.FlowVariable, error) {
	var v model.FlowVariable
	err := r.db.WithContext(ctx).
		Where("instance_id = ? AND key = ?", instanceID, key).
		First(&v).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &v, err
}

func (r *variableRepo) ListByInstance(ctx context.Context, instanceID uuid.UUID) ([]model.FlowVariable, error) {
	var vars []model.FlowVariable
	err := r.db.WithContext(ctx).
		Where("instance_id = ?", instanceID).
		Find(&vars).Error
	return vars, err
}

func (r *variableRepo) GetMap(ctx context.Context, instanceID uuid.UUID) (map[string]interface{}, error) {
	vars, err := r.ListByInstance(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{}, len(vars))
	for _, v := range vars {
		var val interface{}
		if err := json.Unmarshal(v.Value, &val); err != nil {
			result[v.Key] = string(v.Value)
			continue
		}
		result[v.Key] = val
	}
	return result, nil
}

func (r *variableRepo) SetMap(ctx context.Context, tx *gorm.DB, instanceID uuid.UUID, vars map[string]interface{}) error {
	for key, val := range vars {
		jsonBytes, err := json.Marshal(val)
		if err != nil {
			return err
		}
		v := &model.FlowVariable{
			InstanceID: instanceID,
			Key:        key,
			Value:      jsonBytes,
			Scope:      "global",
		}
		if err := r.Set(ctx, tx, v); err != nil {
			return err
		}
	}
	return nil
}

func (r *variableRepo) DeleteByInstance(ctx context.Context, instanceID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("instance_id = ?", instanceID).
		Delete(&model.FlowVariable{}).Error
}
