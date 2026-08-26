package engine

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/model"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/events"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// FlowEngine is the core workflow engine that drives process execution.
type FlowEngine struct {
	defRepo    repo.DefinitionRepo
	instRepo   repo.InstanceRepo
	taskRepo   repo.TaskRepo
	varRepo    repo.VariableRepo
	db         *gorm.DB
	registry   NodeRegistry
	exprEngine *ExpressionEngine
	eventBus   *events.EventBus
	logger     *zap.Logger
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
func (e *FlowEngine) Start(ctx context.Context, definitionKey string, businessKey string, businessType string, initiatorID uuid.UUID, variables map[string]interface{}) (*model.FlowInstance, error) {
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

	// 5. Publish FlowStartedEvent.
	e.eventBus.Publish(ctx, events.FlowStartedEvent{
		InstanceID:   inst.ID,
		DefinitionID: def.ID,
		InitiatorID:  initiatorID,
		BusinessKey:  businessKey,
		BusinessType: businessType,
		StartedAt:    inst.StartedAt,
	})

	// 6. Execute from the start node.
	currentNodeIDs := []string{startNode.ID}
	if err := e.executeFromNodes(ctx, inst, graph, currentNodeIDs, variables); err != nil {
		return nil, err
	}

	return inst, nil
}

// CompleteTask processes a task completion and continues the flow.
func (e *FlowEngine) CompleteTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID, action TaskAction, comment string, formData map[string]interface{}) error {
	// 1. Get the task.
	task, err := e.taskRepo.GetTaskByID(ctx, taskID)
	if err != nil {
		return response.NewAppErrorf(response.CodeWorkflowTaskNotFnd,
			"failed to query task: %v", err)
	}
	if task == nil {
		return response.NewAppError(response.CodeWorkflowTaskNotFnd,
			"流程任务不存在")
	}
	if task.Status != 0 {
		return response.NewAppError(response.CodeWorkflowTaskStatus,
			"流程任务状态不允许操作")
	}

	// 2. Check permission (assignee or initiator for withdraw).
	if action != ActionWithdraw {
		if task.AssigneeID == nil || *task.AssigneeID != userID {
			return response.NewAppError(response.CodeWorkflowTaskNoAccess,
				"无权操作流程任务")
		}
	}

	// 3. Get the instance.
	inst, err := e.instRepo.GetByID(ctx, task.InstanceID)
	if err != nil {
		return response.NewAppErrorf(response.CodeWorkflowInstNotFound,
			"failed to query instance: %v", err)
	}
	if inst == nil {
		return response.NewAppError(response.CodeWorkflowInstNotFound,
			"流程实例不存在")
	}
	if inst.Status != 0 {
		return response.NewAppError(response.CodeWorkflowInstStatus,
			"流程实例状态不允许操作")
	}

	// 4. Get the definition version to parse the graph.
	version, err := e.defRepo.GetVersionByID(ctx, inst.DefinitionVersionID)
	if err != nil || version == nil {
		return response.NewAppError(response.CodeWorkflowVerNotFound,
			"流程版本不存在")
	}

	graph, err := ParseGraph(version.BpmnData)
	if err != nil {
		return response.NewAppErrorf(response.CodeWorkflowInvalidNode,
			"failed to parse graph: %v", err)
	}

	node := graph.GetNode(task.NodeID)
	if node == nil {
		return response.NewAppError(response.CodeWorkflowNodeNotFound,
			"流程节点不存在: "+task.NodeID)
	}

	// 5. Update task status within a transaction.
	now := time.Now()
	switch action {
	case ActionApprove:
		task.Status = 1
	case ActionReject:
		task.Status = 2
	case ActionTransfer:
		task.Status = 3
	case ActionRollback:
		task.Status = 4
	case ActionWithdraw:
		task.Status = 4
	}
	task.Action = string(action)
	task.Comment = comment
	task.CompletedAt = &now
	task.UpdatedAt = now

	// Merge formData into task.
	if len(formData) > 0 {
		formBytes, _ := json.Marshal(formData)
		task.FormData = formBytes
	}

	txErr := e.withTransaction(ctx, func(tx *gorm.DB) error {
		if err := e.taskRepo.UpdateTask(ctx, tx, task); err != nil {
			return err
		}

		// Record history.
		hist := &model.FlowHistory{
			ID:         uuid.New(),
			InstanceID: inst.ID,
			TaskID:     &task.ID,
			NodeID:     node.ID,
			NodeName:   node.Label,
			NodeType:   node.Type,
			OperatorID: &userID,
			Action:     string(action),
			Comment:    comment,
			CreatedAt:  now,
		}
		if err := e.taskRepo.CreateHistory(ctx, tx, hist); err != nil {
			return err
		}

		// Merge form data into variables.
		if len(formData) > 0 {
			if err := e.varRepo.SetMap(ctx, tx, inst.ID, formData); err != nil {
				return err
			}
		}

		return nil
	})
	if txErr != nil {
		return txErr
	}

	// 6. Publish TaskCompletedEvent.
	e.eventBus.Publish(ctx, events.TaskCompletedEvent{
		InstanceID: inst.ID,
		TaskID:     task.ID,
		OperatorID: userID,
		Action:     string(action),
		Comment:    comment,
	})

	// 7. If rejected or withdrawn, stop the flow.
	if action == ActionReject {
		return e.terminateInstance(ctx, inst, userID, "审批驳回")
	}
	if action == ActionWithdraw {
		return e.terminateInstance(ctx, inst, userID, "发起人撤回")
	}

	// 8. Continue execution from the next node.
	if action == ActionApprove {
		vars, _ := e.varRepo.GetMap(ctx, inst.ID)
		nextNodes, err := e.executeNode(ctx, inst, graph, node, vars)
		if err != nil {
			return err
		}
		if len(nextNodes) > 0 {
			return e.executeFromNodes(ctx, inst, graph, nextNodes, vars)
		}
	}

	return nil
}

// Terminate terminates a running flow instance.
func (e *FlowEngine) Terminate(ctx context.Context, instanceID uuid.UUID, operatorID uuid.UUID, reason string) error {
	inst, err := e.instRepo.GetByID(ctx, instanceID)
	if err != nil || inst == nil {
		return response.NewAppError(response.CodeWorkflowInstNotFound,
			"流程实例不存在")
	}
	if inst.Status != 0 {
		return response.NewAppError(response.CodeWorkflowInstStatus,
			"流程实例状态不允许操作")
	}

	return e.terminateInstance(ctx, inst, operatorID, reason)
}

// Suspend suspends a running flow instance.
func (e *FlowEngine) Suspend(ctx context.Context, instanceID uuid.UUID, operatorID uuid.UUID, reason string) error {
	inst, err := e.instRepo.GetByID(ctx, instanceID)
	if err != nil || inst == nil {
		return response.NewAppError(response.CodeWorkflowInstNotFound,
			"流程实例不存在")
	}
	if inst.Status != 0 {
		return response.NewAppError(response.CodeWorkflowInstStatus,
			"流程实例状态不允许操作")
	}

	inst.Status = 3 // suspended
	inst.TerminateReason = reason
	inst.UpdatedAt = time.Now()

	return e.instRepo.Update(ctx, nil, inst)
}

// Resume resumes a suspended flow instance.
func (e *FlowEngine) Resume(ctx context.Context, instanceID uuid.UUID, operatorID uuid.UUID) error {
	inst, err := e.instRepo.GetByID(ctx, instanceID)
	if err != nil || inst == nil {
		return response.NewAppError(response.CodeWorkflowInstNotFound,
			"流程实例不存在")
	}
	if inst.Status != 3 {
		return response.NewAppError(response.CodeWorkflowInstStatus,
			"流程实例状态不允许操作")
	}

	inst.Status = 0 // running
	inst.TerminateReason = ""
	inst.UpdatedAt = time.Now()

	return e.instRepo.Update(ctx, nil, inst)
}

// executeFromNodes processes the flow starting from the given node IDs.
// It follows the flow graph, executing each node until it reaches a wait
// node (approval) or an end node.
func (e *FlowEngine) executeFromNodes(ctx context.Context, inst *model.FlowInstance, graph *FlowGraph, nodeIDs []string, vars map[string]interface{}) error {
	// Update instance's current node IDs.
	e.updateCurrentNodes(ctx, inst, nodeIDs)

	for _, nodeID := range nodeIDs {
		node := graph.GetNode(nodeID)
		if node == nil {
			return response.NewAppErrorf(response.CodeWorkflowNodeNotFound,
				"流程节点不存在: %s", nodeID)
		}

		handler, err := e.registry.Get(node.Type)
		if err != nil {
			return err
		}

		// Publish NodeEnteredEvent.
		e.eventBus.Publish(ctx, events.NodeEnteredEvent{
			InstanceID: inst.ID,
			NodeID:     node.ID,
			NodeType:   node.Type,
		})

		// OnEnter.
		if err := handler.OnEnter(ctx, inst, node, vars); err != nil {
			return err
		}

		// Execute.
		nextNodeIDs, err := handler.Execute(ctx, inst, node, graph, vars)
		if err != nil {
			return err
		}

		// OnLeave.
		if err := handler.OnLeave(ctx, inst, node, vars); err != nil {
			return err
		}

		// Publish NodeLeftEvent.
		e.eventBus.Publish(ctx, events.NodeLeftEvent{
			InstanceID: inst.ID,
			NodeID:     node.ID,
			NodeType:   node.Type,
		})

		// If no next nodes, check if it's an end node.
		if len(nextNodeIDs) == 0 {
			if node.Type == "end" {
				// Instance completed.
				return e.completeInstance(ctx, inst)
			}
			// Wait node (approval) — flow pauses here.
			continue
		}

		// Recursively execute next nodes.
		if err := e.executeFromNodes(ctx, inst, graph, nextNodeIDs, vars); err != nil {
			return err
		}
	}

	return nil
}

// executeNode runs a single node and returns the next node IDs.
// Used by CompleteTask to continue after task approval.
func (e *FlowEngine) executeNode(ctx context.Context, inst *model.FlowInstance, graph *FlowGraph, node *FlowNode, vars map[string]interface{}) ([]string, error) {
	handler, err := e.registry.Get(node.Type)
	if err != nil {
		return nil, err
	}

	// Publish NodeLeftEvent for the completed node.
	e.eventBus.Publish(ctx, events.NodeLeftEvent{
		InstanceID: inst.ID,
		NodeID:     node.ID,
		NodeType:   node.Type,
	})

	// Execute to get next nodes.
	nextNodeIDs, err := handler.Execute(ctx, inst, node, graph, vars)
	if err != nil {
		return nil, err
	}

	// OnLeave.
	if err := handler.OnLeave(ctx, inst, node, vars); err != nil {
		return nil, err
	}

	return nextNodeIDs, nil
}

// completeInstance marks an instance as completed.
func (e *FlowEngine) completeInstance(ctx context.Context, inst *model.FlowInstance) error {
	now := time.Now()
	inst.Status = 1 // completed
	inst.EndedAt = &now
	inst.UpdatedAt = now

	if err := e.instRepo.Update(ctx, nil, inst); err != nil {
		return err
	}

	// Publish FlowCompletedEvent.
	e.eventBus.Publish(ctx, events.FlowCompletedEvent{
		InstanceID: inst.ID,
		EndedAt:    now,
	})

	return nil
}

// terminateInstance marks an instance as terminated.
func (e *FlowEngine) terminateInstance(ctx context.Context, inst *model.FlowInstance, operatorID uuid.UUID, reason string) error {
	now := time.Now()
	inst.Status = 2 // terminated
	inst.EndedAt = &now
	inst.TerminateReason = reason
	inst.UpdatedAt = now

	if err := e.instRepo.Update(ctx, nil, inst); err != nil {
		return err
	}

	// Publish FlowTerminatedEvent.
	e.eventBus.Publish(ctx, events.FlowTerminatedEvent{
		InstanceID: inst.ID,
		Reason:     reason,
		OperatorID: operatorID,
		EndedAt:    now,
	})

	return nil
}

// updateCurrentNodes persists the current node IDs to the instance.
func (e *FlowEngine) updateCurrentNodes(ctx context.Context, inst *model.FlowInstance, nodeIDs []string) {
	nodeBytes, _ := json.Marshal(nodeIDs)
	inst.CurrentNodeIDs = nodeBytes
	_ = e.instRepo.Update(ctx, nil, inst)
}

// withTransaction runs a function within a database transaction.
func (e *FlowEngine) withTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return e.db.WithContext(ctx).Transaction(fn)
}
