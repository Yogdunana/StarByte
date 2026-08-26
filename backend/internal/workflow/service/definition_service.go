package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/dto"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/model"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// definitionServiceImpl handles flow definition business logic.
type definitionServiceImpl struct {
	defRepo repo.DefinitionRepo
	db      *gorm.DB
}

// NewDefinitionService creates a DefinitionService.
func NewDefinitionService(defRepo repo.DefinitionRepo, db *gorm.DB) DefinitionService {
	return &definitionServiceImpl{defRepo: defRepo, db: db}
}

// Create creates a new flow definition (draft status).
func (s *definitionServiceImpl) Create(ctx context.Context, req *dto.CreateDefinitionRequest, userID uuid.UUID) (*model.FlowDefinition, error) {
	// Check for duplicate key.
	existing, err := s.defRepo.GetByKey(ctx, req.Key)
	if err != nil {
		return nil, response.NewAppErrorf(response.CodeInternalError,
			"failed to check key: %v", err)
	}
	if existing != nil {
		return nil, response.NewAppError(response.CodeWorkflowKeyExists,
			"流程定义 key 已存在")
	}

	// Apply defaults.
	category := req.Category
	if category == "" {
		category = "custom"
	}

	def := &model.FlowDefinition{
		ID:          uuid.New(),
		Key:         req.Key,
		Name:        req.Name,
		Description: req.Description,
		Category:    category,
		Status:      0, // draft
		CreatedBy:   &userID,
		UpdatedBy:   &userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.defRepo.Create(ctx, nil, def); err != nil {
		return nil, response.NewAppErrorf(response.CodeInternalError,
			"failed to create definition: %v", err)
	}

	return def, nil
}

// GetByID retrieves a flow definition by ID.
func (s *definitionServiceImpl) GetByID(ctx context.Context, id uuid.UUID) (*model.FlowDefinition, error) {
	def, err := s.defRepo.GetByID(ctx, id)
	if err != nil {
		return nil, response.NewAppErrorf(response.CodeInternalError,
			"failed to query definition: %v", err)
	}
	if def == nil {
		return nil, response.NewAppError(response.CodeWorkflowNotFound,
			"流程定义不存在")
	}
	return def, nil
}

// Update updates a flow definition (only in draft status).
func (s *definitionServiceImpl) Update(ctx context.Context, id uuid.UUID, req *dto.UpdateDefinitionRequest, userID uuid.UUID) (*model.FlowDefinition, error) {
	def, err := s.defRepo.GetByID(ctx, id)
	if err != nil || def == nil {
		return nil, response.NewAppError(response.CodeWorkflowNotFound,
			"流程定义不存在")
	}
	if def.Status == 1 {
		return nil, response.NewAppError(response.CodeWorkflowDefPublished,
			"流程定义已发布，不可修改")
	}

	def.Name = req.Name
	def.Description = req.Description
	def.UpdatedBy = &userID
	def.UpdatedAt = time.Now()

	if err := s.defRepo.Update(ctx, nil, def); err != nil {
		return nil, response.NewAppErrorf(response.CodeInternalError,
			"failed to update definition: %v", err)
	}

	return def, nil
}

// Delete deletes a flow definition (only in draft status).
func (s *definitionServiceImpl) Delete(ctx context.Context, id uuid.UUID) error {
	def, err := s.defRepo.GetByID(ctx, id)
	if err != nil || def == nil {
		return response.NewAppError(response.CodeWorkflowNotFound,
			"流程定义不存在")
	}
	if def.Status == 1 {
		return response.NewAppError(response.CodeWorkflowDefPublished,
			"已发布的流程定义不可删除")
	}

	if err := s.defRepo.Delete(ctx, id); err != nil {
		return response.NewAppErrorf(response.CodeInternalError,
			"failed to delete definition: %v", err)
	}

	return nil
}

// List returns a paginated list of flow definitions.
func (s *definitionServiceImpl) List(ctx context.Context, page, pageSize int, keyword, category string, status *int) ([]model.FlowDefinition, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	defs, total, err := s.defRepo.List(ctx, page, pageSize, keyword, category, status)
	if err != nil {
		return nil, 0, response.NewAppErrorf(response.CodeInternalError,
			"failed to list definitions: %v", err)
	}

	return defs, total, nil
}

// Publish publishes a new version of a flow definition.
func (s *definitionServiceImpl) Publish(ctx context.Context, id uuid.UUID, req *dto.PublishDefinitionRequest, userID uuid.UUID) (*model.FlowDefinitionVersion, error) {
	def, err := s.defRepo.GetByID(ctx, id)
	if err != nil || def == nil {
		return nil, response.NewAppError(response.CodeWorkflowNotFound,
			"流程定义不存在")
	}

	// Validate the graph has a start node.
	graphData, err := json.Marshal(req.GraphData)
	if err != nil {
		return nil, response.NewAppErrorf(response.CodeWorkflowInvalidNode,
			"failed to marshal graph data: %v", err)
	}

	// Calculate next version number.
	versions, err := s.defRepo.ListVersions(ctx, id)
	if err != nil {
		return nil, response.NewAppErrorf(response.CodeInternalError,
			"failed to list versions: %v", err)
	}
	nextVersion := 1
	if len(versions) > 0 {
		nextVersion = versions[0].Version + 1
	}

	now := time.Now()
	ver := &model.FlowDefinitionVersion{
		ID:           uuid.New(),
		DefinitionID: id,
		Version:      nextVersion,
		BpmnData:     graphData,
		Status:       1, // current
		PublishedBy:  &userID,
		PublishedAt:  &now,
		CreatedAt:    now,
	}

	// Use a transaction: mark old version as historical + create new version.
	txErr := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Mark previous current version as historical.
		if err := s.defRepo.MarkVersionHistorical(ctx, tx, id); err != nil {
			return err
		}
		// Create new version.
		if err := s.defRepo.CreateVersion(ctx, tx, ver); err != nil {
			return err
		}
		return nil
	})

	if txErr != nil {
		return nil, response.NewAppErrorf(response.CodeInternalError,
			"failed to publish version: %v", txErr)
	}

	// Update definition status to published.
	def.Status = 1
	def.UpdatedBy = &userID
	def.UpdatedAt = now
	if err := s.defRepo.Update(ctx, nil, def); err != nil {
		return nil, response.NewAppErrorf(response.CodeInternalError,
			"failed to update definition status: %v", err)
	}

	return ver, nil
}

// ListVersions returns all versions of a definition.
func (s *definitionServiceImpl) ListVersions(ctx context.Context, definitionID uuid.UUID) ([]model.FlowDefinitionVersion, error) {
	versions, err := s.defRepo.ListVersions(ctx, definitionID)
	if err != nil {
		return nil, response.NewAppErrorf(response.CodeInternalError,
			"failed to list versions: %v", err)
	}
	return versions, nil
}

// GetVersionByID retrieves a specific version.
func (s *definitionServiceImpl) GetVersionByID(ctx context.Context, id uuid.UUID) (*model.FlowDefinitionVersion, error) {
	ver, err := s.defRepo.GetVersionByID(ctx, id)
	if err != nil {
		return nil, response.NewAppErrorf(response.CodeInternalError,
			"failed to query version: %v", err)
	}
	if ver == nil {
		return nil, response.NewAppError(response.CodeWorkflowVerNotFound,
			"流程版本不存在")
	}
	return ver, nil
}
