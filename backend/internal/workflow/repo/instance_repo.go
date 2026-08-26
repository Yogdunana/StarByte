package repo

import (
	"context"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
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
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&inst).Error
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

	query := r.db.WithContext(ctx).Model(&model.FlowInstance{})

	if status != nil {
		query = query.Where("status = ?", *status)
	}
	if definitionID != nil && *definitionID != uuid.Nil {
		query = query.Where("definition_id = ?", *definitionID)
	}
	if initiatorID != nil && *initiatorID != uuid.Nil {
		query = query.Where("initiator_id = ?", *initiatorID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&instances).Error
	return instances, total, err
}
