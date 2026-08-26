package service

import (
	"context"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/engine"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/model"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// InstanceService handles flow instance business logic.
type InstanceService struct {
	instRepo   repo.InstanceRepo
	defRepo    repo.DefinitionRepo
	taskRepo   repo.TaskRepo
	flowEngine *engine.FlowEngine
	db         *gorm.DB
}

// NewInstanceService creates an InstanceService.
func NewInstanceService(instRepo repo.InstanceRepo, defRepo repo.DefinitionRepo, taskRepo repo.TaskRepo, flowEngine *engine.FlowEngine, db *gorm.DB) *InstanceService {
	return &InstanceService{
		instRepo:   instRepo,
		defRepo:    defRepo,
		taskRepo:   taskRepo,
		flowEngine: flowEngine,
		db:         db,
	}
}

// Start creates and starts a new flow instance.
func (s *InstanceService) Start(ctx context.Context, defID uuid.UUID, businessKey, businessType string, initiatorID uuid.UUID, variables map[string]interface{}) (*model.FlowInstance, error) {
	// Find the definition to get the key.
	def, err := s.defRepo.GetByID(ctx, defID)
	if err != nil || def == nil {
		return nil, response.NewAppError(response.CodeWorkflowNotFound,
			"流程定义不存在")
	}
	if def.Status != 1 {
		return nil, response.NewAppError(response.CodeWorkflowDefNotPub,
			"流程定义未发布")
	}

	// Start the instance via the flow engine.
	inst, err := s.flowEngine.Start(ctx, def.Key, businessKey, businessType, initiatorID, variables)
	if err != nil {
		return nil, err
	}

	return inst, nil
}

// GetByID retrieves a flow instance by ID.
func (s *InstanceService) GetByID(ctx context.Context, id uuid.UUID) (*model.FlowInstance, error) {
	inst, err := s.instRepo.GetByID(ctx, id)
	if err != nil {
		return nil, response.NewAppErrorf(response.CodeInternalError,
			"failed to query instance: %v", err)
	}
	if inst == nil {
		return nil, response.NewAppError(response.CodeWorkflowInstNotFound,
			"流程实例不存在")
	}
	return inst, nil
}

// List returns a paginated list of flow instances.
func (s *InstanceService) List(ctx context.Context, page, pageSize int, status *int, definitionID, initiatorID *uuid.UUID) ([]model.FlowInstance, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	instances, total, err := s.instRepo.List(ctx, page, pageSize, status, definitionID, initiatorID)
	if err != nil {
		return nil, 0, response.NewAppErrorf(response.CodeInternalError,
			"failed to list instances: %v", err)
	}

	return instances, total, nil
}

// Terminate terminates a flow instance.
func (s *InstanceService) Terminate(ctx context.Context, instanceID uuid.UUID, operatorID uuid.UUID, reason string) error {
	return s.flowEngine.Terminate(ctx, instanceID, operatorID, reason)
}

// Suspend suspends a running flow instance.
func (s *InstanceService) Suspend(ctx context.Context, instanceID uuid.UUID, operatorID uuid.UUID, reason string) error {
	return s.flowEngine.Suspend(ctx, instanceID, operatorID, reason)
}

// Resume resumes a suspended flow instance.
func (s *InstanceService) Resume(ctx context.Context, instanceID uuid.UUID, operatorID uuid.UUID) error {
	return s.flowEngine.Resume(ctx, instanceID, operatorID)
}

// ListHistory returns the history for a flow instance.
func (s *InstanceService) ListHistory(ctx context.Context, instanceID uuid.UUID) ([]model.FlowHistory, error) {
	histories, err := s.taskRepo.ListHistory(ctx, instanceID)
	if err != nil {
		return nil, response.NewAppErrorf(response.CodeInternalError,
			"failed to list history: %v", err)
	}
	return histories, nil
}
