package repo

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/model"
)

// InstanceRepo manages flow_instances.
type InstanceRepo interface {
	// Create creates a new flow instance.
	Create(ctx context.Context, tx *gorm.DB, inst *model.FlowInstance) error
	// GetByID retrieves an instance by its primary key.
	GetByID(ctx context.Context, id uuid.UUID) (*model.FlowInstance, error)
	// Update saves changes to an existing instance.
	Update(ctx context.Context, tx *gorm.DB, inst *model.FlowInstance) error
	// List returns a paginated list of instances with filters.
	List(ctx context.Context, page, pageSize int, status *int, definitionID, initiatorID *uuid.UUID) ([]model.FlowInstance, int64, error)
}

type instanceRepo struct {
	db *gorm.DB
}

// NewInstanceRepo creates an InstanceRepo backed by the given GORM DB.
func NewInstanceRepo(db *gorm.DB) InstanceRepo {
	return &instanceRepo{db: db}
}

func (r *instanceRepo) getDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.db
}

func (r *instanceRepo) Create(ctx context.Context, tx *gorm.DB, inst *model.FlowInstance) error {
	return r.getDB(tx).WithContext(ctx).Create(inst).Error
}

func (r *instanceRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.FlowInstance, error) {
	var inst model.FlowInstance
	err := r.instanceQuery(ctx).Where("flow_instances.id = ?", id).First(&inst).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &inst, err
}

func (r *instanceRepo) Update(ctx context.Context, tx *gorm.DB, inst *model.FlowInstance) error {
	return r.getDB(tx).WithContext(ctx).Save(inst).Error
}

func (r *instanceRepo) List(ctx context.Context, page, pageSize int, status *int, definitionID, initiatorID *uuid.UUID) ([]model.FlowInstance, int64, error) {
	var instances []model.FlowInstance
	var total int64

	query := visibleInstances(r.instanceQuery(ctx), model.ViewerFromContext(ctx), true)

	if status != nil {
		query = query.Where("flow_instances.status = ?", *status)
	}
	if definitionID != nil && *definitionID != uuid.Nil {
		query = query.Where("flow_instances.definition_id = ?", *definitionID)
	}
	if initiatorID != nil && *initiatorID != uuid.Nil {
		query = query.Where("flow_instances.initiator_id = ?", *initiatorID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("flow_instances.created_at DESC").Offset(offset).Limit(pageSize).Find(&instances).Error
	return instances, total, err
}

func (r *instanceRepo) instanceQuery(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Model(&model.FlowInstance{}).Select("flow_instances.*, definition.name AS definition_name, initiator.real_name AS initiator_name").Joins("LEFT JOIN flow_definitions definition ON definition.id=flow_instances.definition_id").Joins("LEFT JOIN users initiator ON initiator.id=flow_instances.initiator_id")
}
