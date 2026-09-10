package engine

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/model"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/events"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

// FlowEngine is the core workflow engine that drives process execution.
type FlowEngine struct {
	businessTransaction bool
	inTransaction       bool
	defRepo             repo.DefinitionRepo
	instRepo            repo.InstanceRepo
	taskRepo            repo.TaskRepo
	varRepo             repo.VariableRepo
	db                  *gorm.DB
	registry            NodeRegistry
	exprEngine          *ExpressionEngine
	eventBus            *events.EventBus
	logger              *zap.Logger
}

// NodeRegistry is the interface for the node handler registry.
type NodeRegistry interface {
	Get(nodeType string) (NodeHandler, error)
}

// NewFlowEngine creates a new FlowEngine with the given dependencies.
func NewFlowEngine(
	defRepo repo.DefinitionRepo,
	instRepo repo.InstanceRepo,
	taskRepo repo.TaskRepo,
	varRepo repo.VariableRepo,
	db *gorm.DB,
	registry NodeRegistry,
	exprEngine *ExpressionEngine,
	eventBus *events.EventBus,
	logger *zap.Logger,
) *FlowEngine {
	return &FlowEngine{
		defRepo:    defRepo,
		instRepo:   instRepo,
		taskRepo:   taskRepo,
		varRepo:    varRepo,
		db:         db,
		registry:   registry,
		exprEngine: exprEngine,
		eventBus:   eventBus,
		logger:     logger,
	}
}

// Start initiates a new flow instance and begins execution from the start node.
func (e *FlowEngine) start(ctx context.Context, definitionKey string, businessKey string, businessType string, initiatorID uuid.UUID, variables map[string]interface{}) (*model.FlowInstance, error) {
	if (IsProtectedBusiness(businessType) || definitionKey == TaskDefinitionKey || IsTaskTransferDefinition(definitionKey) || definitionKey == "officer_interview" || definitionKey == "member_admission") && !e.businessTransaction {
		return nil, response.NewAppError(response.CodeForbidden, "业务流程须从对应业务页面发起")
	}
	if err := validateInputVariables(variables); err != nil {
		return nil, response.NewAppError(response.CodeBadRequest, err.Error())
	}
	if variables == nil {
		variables = map[string]interface{}{}
	}
	// 1. Find the published definition by key.
	def, err := e.defRepo.GetByKey(ctx, definitionKey)
	if err != nil {
		return nil, response.NewAppErrorf(response.CodeWorkflowNotFound,
			"failed to query definition: %v", err)
	}
	if def == nil {
		return nil, response.NewAppError(response.CodeWorkflowNotFound,
			"流程定义不存在: "+definitionKey)
	}
	if def.Status != 1 {
		return nil, response.NewAppError(response.CodeWorkflowDefNotPub,
			"流程定义未发布")
	}

	// 2. Get the current published version.
	version, err := e.defRepo.GetCurrentVersion(ctx, def.ID)
	if err != nil {
		return nil, response.NewAppErrorf(response.CodeWorkflowVerNotFound,
			"failed to query current version: %v", err)
	}
	if version == nil {
		return nil, response.NewAppError(response.CodeWorkflowVerNotFound,
			"流程版本不存在")
	}

	// 3. Parse the graph.
	graph, err := ParseGraph(version.BpmnData)
	if err != nil {
		return nil, response.NewAppErrorf(response.CodeWorkflowInvalidNode,
			"failed to parse graph: %v", err)
	}

	startNode := graph.FindStartNode()
	if err := ValidateBusinessDefinition(def.Key, graph); err != nil {
		return nil, response.NewAppError(response.CodeWorkflowInvalidNode, err.Error())
	}
	if startNode == nil {
		return nil, response.NewAppError(response.CodeWorkflowInvalidNode,
			"流程图中没有 start 节点")
	}

	// 4. Create the instance within a transaction.
	inst := &model.FlowInstance{
		ID:                  uuid.New(),
		DefinitionID:        def.ID,
		DefinitionVersionID: version.ID,
		BusinessKey:         businessKey,
		BusinessType:        businessType,
		InitiatorID:         initiatorID,
		Status:              0,
		StartedAt:           time.Now(),
	}

	txErr := e.withTransaction(ctx, func(tx *gorm.DB) error {
		if err := e.instRepo.Create(ctx, tx, inst); err != nil {
			return err
		}
		// Initialize variables.
		if len(variables) > 0 {
			if err := e.varRepo.SetMap(ctx, tx, inst.ID, variables); err != nil {
				return err
			}
		}
		return nil
	})
	if txErr != nil {
		return nil, txErr
	}

	// 5. Execute from the start node.
	currentNodeIDs := []string{startNode.ID}
	if err := e.executeFromNodes(ctx, inst, graph, currentNodeIDs, variables); err != nil {
		// If execution fails, terminate the instance and publish a terminated event
		// to keep external systems in sync.
		_ = e.terminateInstance(ctx, inst, uuid.Nil, "启动失败: "+err.Error())
		return nil, err
	}

	// 6. Publish FlowStartedEvent after successful execution start.
	e.eventBus.Publish(ctx, events.FlowStartedEvent{
		InstanceID:   inst.ID,
		DefinitionID: def.ID,
		InitiatorID:  initiatorID,
		BusinessKey:  businessKey,
		BusinessType: businessType,
		StartedAt:    inst.StartedAt,
	})

	return inst, nil
}

// withTransaction runs a function within a database transaction.
func (e *FlowEngine) withTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return e.db.WithContext(ctx).Transaction(fn)
}
