package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/dto"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/engine"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/engine/nodes"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/model"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
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

// Update updates a draft while holding the same lock used by publication.
func (s *definitionServiceImpl) Update(ctx context.Context, id uuid.UUID, req *dto.UpdateDefinitionRequest, userID uuid.UUID) (*model.FlowDefinition, error) {
	return s.changeDefinition(ctx, id, func(repository repo.DefinitionRepo, def *model.FlowDefinition) error {
		if def.Status == 1 {
			return response.NewAppError(response.CodeWorkflowDefPublished, "流程定义已发布，不可修改")
		}
		def.Name, def.Description, def.UpdatedBy, def.UpdatedAt = req.Name, req.Description, &userID, time.Now()
		return repository.Update(ctx, nil, def)
	})
}

func (s *definitionServiceImpl) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := s.changeDefinition(ctx, id, func(repository repo.DefinitionRepo, def *model.FlowDefinition) error {
		if def.Status == 1 {
			return response.NewAppError(response.CodeWorkflowDefPublished, "已发布的流程定义不可删除")
		}
		return repository.Delete(ctx, id)
	})
	return err
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

	if err := s.requireDefinitionAccess(ctx, def); err != nil {
		return nil, err
	}

	// Defensive check: binding:"required" should already reject nil,
	// but guard against direct service calls without handler validation.
	if req.GraphData == nil {
		return nil, response.NewAppError(response.CodeBadRequest, "graph_data 不能为空")
	}

	// Serialize the graph data for storage.
	graphData, err := json.Marshal(req.GraphData)
	if err != nil {
		return nil, response.NewAppErrorf(response.CodeWorkflowInvalidNode,
			"failed to marshal graph data: %v", err)
	}

	graph, err := engine.ParseGraph(graphData)
	if err != nil {
		return nil, response.NewAppError(response.CodeWorkflowInvalidNode, err.Error())
	}
	if err := engine.ValidateGraph(graph, nodes.NewDefaultRegistry(engine.NewExpressionEngine(), nil, nil)); err != nil {
		return nil, response.NewAppError(response.CodeWorkflowInvalidNode, err.Error())
	}
	if err := engine.ValidateBusinessDefinition(def.Key, graph); err != nil {
		return nil, response.NewAppError(response.CodeWorkflowInvalidNode, err.Error())
	}
	var ver *model.FlowDefinitionVersion
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		bound := repo.NewDefinitionRepo(tx)
		locked, err := repo.NewRuntimeRepo(tx).LockDefinition(ctx, id)
		if err != nil {
			return err
		}
		if err := s.requireDefinitionAccess(ctx, locked); err != nil {
			return err
		}
		versions, err := bound.ListVersions(ctx, id)
		if err != nil {
			return err
		}
		next := 1
		for _, previous := range versions {
			if previous.Version >= next {
				next = previous.Version + 1
			}
		}
		now := time.Now()
		ver = &model.FlowDefinitionVersion{ID: uuid.New(), DefinitionID: id, Version: next, BpmnData: graphData, Status: 1, PublishedBy: &userID, PublishedAt: &now, CreatedAt: now}
		if err := bound.MarkVersionHistorical(ctx, nil, id); err != nil {
			return err
		}
		if err := bound.CreateVersion(ctx, nil, ver); err != nil {
			return err
		}
		locked.Status, locked.UpdatedBy, locked.UpdatedAt = 1, &userID, now
		return bound.Update(ctx, nil, locked)
	})
	if err != nil {
		return nil, err
	}

	return ver, nil
}

// SaveDraft stores a working-copy graph on the definition.
// Allowed for both draft and published definitions (next-version WIP).
func (s *definitionServiceImpl) SaveDraft(ctx context.Context, id uuid.UUID, req *dto.SaveDraftRequest, userID uuid.UUID) (*model.FlowDefinition, error) {
	return s.changeDefinition(ctx, id, func(repository repo.DefinitionRepo, def *model.FlowDefinition) error {
		if req == nil || req.GraphData == nil {
			return response.NewAppError(response.CodeBadRequest, "graph_data 不能为空")
		}
		data, err := json.Marshal(req.GraphData)
		if err != nil {
			return err
		}
		def.DraftGraph, def.UpdatedBy, def.UpdatedAt = data, &userID, time.Now()
		return repository.Update(ctx, nil, def)
	})
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

func (s *definitionServiceImpl) changeDefinition(ctx context.Context, id uuid.UUID, change func(repo.DefinitionRepo, *model.FlowDefinition) error) (*model.FlowDefinition, error) {
	var result *model.FlowDefinition
	operation := func(tx *gorm.DB) error {
		repository := s.defRepo
		var err error
		if tx != nil {
			repository = repo.NewDefinitionRepo(tx)
			result, err = repo.NewRuntimeRepo(tx).LockDefinition(ctx, id)
		} else {
			result, err = repository.GetByID(ctx, id)
		}
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}
		if result == nil || err == gorm.ErrRecordNotFound {
			return response.NewAppError(response.CodeWorkflowNotFound, "流程定义不存在")
		}
		if err := s.requireDefinitionAccess(ctx, result); err != nil {
			return err
		}
		return change(repository, result)
	}
	var err error
	if s.db == nil {
		err = operation(nil)
	} else {
		err = s.db.WithContext(ctx).Transaction(operation)
	}
	return result, err
}
