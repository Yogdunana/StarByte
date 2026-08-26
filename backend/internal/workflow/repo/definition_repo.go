package repo

import (
	"context"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DefinitionRepo manages flow_definitions and flow_definition_versions.
type DefinitionRepo interface {
	// Create creates a new flow definition.
	Create(ctx context.Context, tx *gorm.DB, def *model.FlowDefinition) error
	// GetByID retrieves a definition by its primary key.
	GetByID(ctx context.Context, id uuid.UUID) (*model.FlowDefinition, error)
	// GetByKey retrieves a definition by its unique key.
	GetByKey(ctx context.Context, key string) (*model.FlowDefinition, error)
	// Update saves changes to an existing definition.
	Update(ctx context.Context, tx *gorm.DB, def *model.FlowDefinition) error
	// Delete removes a definition by its primary key.
	Delete(ctx context.Context, id uuid.UUID) error
	// List returns a paginated list of definitions with filters.
	List(ctx context.Context, page, pageSize int, keyword, category string, status *int) ([]model.FlowDefinition, int64, error)

	// CreateVersion creates a new version of a flow definition.
	CreateVersion(ctx context.Context, tx *gorm.DB, ver *model.FlowDefinitionVersion) error
	// GetVersionByID retrieves a specific version by its primary key.
	GetVersionByID(ctx context.Context, id uuid.UUID) (*model.FlowDefinitionVersion, error)
	// GetCurrentVersion retrieves the current (status=1) version of a definition.
	GetCurrentVersion(ctx context.Context, definitionID uuid.UUID) (*model.FlowDefinitionVersion, error)
	// ListVersions returns all versions of a definition, newest first.
	ListVersions(ctx context.Context, definitionID uuid.UUID) ([]model.FlowDefinitionVersion, error)
	// MarkVersionHistorical sets status=0 for old current version.
	MarkVersionHistorical(ctx context.Context, tx *gorm.DB, definitionID uuid.UUID) error
}

type definitionRepo struct {
	db *gorm.DB
}

// NewDefinitionRepo creates a DefinitionRepo backed by the given GORM DB.
func NewDefinitionRepo(db *gorm.DB) DefinitionRepo {
	return &definitionRepo{db: db}
}

func (r *definitionRepo) getDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.db
}

func (r *definitionRepo) Create(ctx context.Context, tx *gorm.DB, def *model.FlowDefinition) error {
	return r.getDB(tx).WithContext(ctx).Create(def).Error
}

func (r *definitionRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.FlowDefinition, error) {
	var def model.FlowDefinition
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&def).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &def, err
}

func (r *definitionRepo) GetByKey(ctx context.Context, key string) (*model.FlowDefinition, error) {
	var def model.FlowDefinition
	err := r.db.WithContext(ctx).Where("key = ?", key).First(&def).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &def, err
}

func (r *definitionRepo) Update(ctx context.Context, tx *gorm.DB, def *model.FlowDefinition) error {
	return r.getDB(tx).WithContext(ctx).Save(def).Error
}

func (r *definitionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.FlowDefinition{}, id).Error
}

func (r *definitionRepo) List(ctx context.Context, page, pageSize int, keyword, category string, status *int) ([]model.FlowDefinition, int64, error) {
	var defs []model.FlowDefinition
	var total int64

	query := r.db.WithContext(ctx).Model(&model.FlowDefinition{})

	if keyword != "" {
		query = query.Where("name LIKE ? OR key LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("updated_at DESC").Offset(offset).Limit(pageSize).Find(&defs).Error
	return defs, total, err
}

func (r *definitionRepo) CreateVersion(ctx context.Context, tx *gorm.DB, ver *model.FlowDefinitionVersion) error {
	return r.getDB(tx).WithContext(ctx).Create(ver).Error
}

func (r *definitionRepo) GetVersionByID(ctx context.Context, id uuid.UUID) (*model.FlowDefinitionVersion, error) {
	var ver model.FlowDefinitionVersion
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&ver).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &ver, err
}

func (r *definitionRepo) GetCurrentVersion(ctx context.Context, definitionID uuid.UUID) (*model.FlowDefinitionVersion, error) {
	var ver model.FlowDefinitionVersion
	err := r.db.WithContext(ctx).
		Where("definition_id = ? AND status = 1", definitionID).
		First(&ver).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &ver, err
}

func (r *definitionRepo) ListVersions(ctx context.Context, definitionID uuid.UUID) ([]model.FlowDefinitionVersion, error) {
	var versions []model.FlowDefinitionVersion
	err := r.db.WithContext(ctx).
		Where("definition_id = ?", definitionID).
		Order("version DESC").
		Find(&versions).Error
	return versions, err
}

func (r *definitionRepo) MarkVersionHistorical(ctx context.Context, tx *gorm.DB, definitionID uuid.UUID) error {
	return r.getDB(tx).WithContext(ctx).
		Model(&model.FlowDefinitionVersion{}).
		Where("definition_id = ? AND status = 1", definitionID).
		Update("status", 0).Error
}
