package dto

import (
	"encoding/json"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/model"
	"github.com/google/uuid"
)

// ========== Flow Definition DTOs ==========

// CreateDefinitionRequest is the body for POST /api/v1/workflow/definitions.
type CreateDefinitionRequest struct {
	Key         string `json:"key" binding:"required,max=100"`
	Name        string `json:"name" binding:"required,max=200"`
	Description string `json:"description" binding:"max=2000"`
	Category    string `json:"category" binding:"max=50"`
}

// UpdateDefinitionRequest is the body for PUT /api/v1/workflow/definitions/:id.
type UpdateDefinitionRequest struct {
	Name        string `json:"name" binding:"required,max=200"`
	Description string `json:"description" binding:"max=2000"`
}

// PublishDefinitionRequest is the body for POST /api/v1/workflow/definitions/:id/publish.
type PublishDefinitionRequest struct {
	GraphData *GraphData `json:"graph_data" binding:"required"`
}

// SaveDraftRequest is the body for PUT /api/v1/workflow/definitions/:id/draft.
type SaveDraftRequest struct {
	GraphData *GraphData `json:"graph_data" binding:"required"`
}

// GraphData represents the React Flow graph structure stored in bpmn_data.
type GraphData struct {
	Nodes []GraphNode `json:"nodes"`
	Edges []GraphEdge `json:"edges"`
}

// GraphNode represents a single node in the flow graph.
type GraphNode struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Position GraphPosition          `json:"position"`
	Data     map[string]interface{} `json:"data"`
}

// GraphPosition is the x/y coordinate of a node on the canvas.
type GraphPosition struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// GraphEdge represents a connection between two nodes.
type GraphEdge struct {
	ID           string                 `json:"id"`
	Source       string                 `json:"source"`
	Target       string                 `json:"target"`
	SourceHandle string                 `json:"sourceHandle,omitempty"`
	Label        string                 `json:"label,omitempty"`
	Data         map[string]interface{} `json:"data,omitempty"`
}

// DefinitionResponse is the response for a single flow definition.
type DefinitionResponse struct {
	ID          uuid.UUID  `json:"id"`
	Key         string     `json:"key"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Category    string     `json:"category"`
	Status      int        `json:"status"`
	DraftGraph  *GraphData `json:"draft_graph"`
	CreatedBy   *uuid.UUID `json:"created_by"`
	UpdatedBy   *uuid.UUID `json:"updated_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// VersionResponse is the response for a flow definition version.
type VersionResponse struct {
	ID           uuid.UUID  `json:"id"`
	DefinitionID uuid.UUID  `json:"definition_id"`
	Version      int        `json:"version"`
	BpmnData     GraphData  `json:"bpmn_data"`
	Status       int        `json:"status"`
	PublishedBy  *uuid.UUID `json:"published_by"`
	PublishedAt  *time.Time `json:"published_at"`
	CreatedAt    time.Time  `json:"created_at"`
}

// ========== Flow Instance DTOs ==========

// StartInstanceRequest is the body for POST /api/v1/workflow/instances.
type StartInstanceRequest struct {
	DefinitionID uuid.UUID              `json:"definition_id" binding:"required"`
	BusinessKey  string                 `json:"business_key" binding:"max=100"`
	BusinessType string                 `json:"business_type" binding:"max=50"`
	Variables    map[string]interface{} `json:"variables"`
}

// InstanceResponse is the response for a single flow instance.
type InstanceResponse struct {
	ID                  uuid.UUID  `json:"id"`
	DefinitionID        uuid.UUID  `json:"definition_id"`
	DefinitionVersionID uuid.UUID  `json:"definition_version_id"`
	BusinessKey         string     `json:"business_key"`
	BusinessType        string     `json:"business_type"`
	InitiatorID         uuid.UUID  `json:"initiator_id"`
	Status              int        `json:"status"`
	CurrentNodeIDs      []string   `json:"current_node_ids"`
	StartedAt           time.Time  `json:"started_at"`
	EndedAt             *time.Time `json:"ended_at"`
	TerminateReason     string     `json:"terminate_reason"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// TerminateInstanceRequest is the body for POST /api/v1/workflow/instances/:id/terminate.
type TerminateInstanceRequest struct {
	Reason string `json:"reason" binding:"required,max=500"`
}

// SuspendInstanceRequest is the body for POST /api/v1/workflow/instances/:id/suspend.
type SuspendInstanceRequest struct {
	Reason string `json:"reason" binding:"required,max=500"`
}

// ========== Flow Task DTOs ==========

// CompleteTaskRequest is the body for POST /api/v1/workflow/tasks/:id/complete.
type CompleteTaskRequest struct {
	Action    string                 `json:"action" binding:"required,oneof=approve reject transfer"`
	Comment   string                 `json:"comment" binding:"max=1000"`
	Variables map[string]interface{} `json:"variables"`
}

// TransferTaskRequest is the body for POST /api/v1/workflow/tasks/:id/transfer.
type TransferTaskRequest struct {
	TargetUserID uuid.UUID `json:"target_user_id" binding:"required"`
	Comment      string    `json:"comment" binding:"max=1000"`
}

// RollbackTaskRequest is the body for POST /api/v1/workflow/tasks/:id/back.
type RollbackTaskRequest struct {
	TargetNodeID string `json:"target_node_id" binding:"required"`
	Comment      string `json:"comment" binding:"max=1000"`
}

// TaskResponse is the response for a single flow task.
type TaskResponse struct {
	ID          uuid.UUID       `json:"id"`
	InstanceID  uuid.UUID       `json:"instance_id"`
	NodeID      string          `json:"node_id"`
	NodeName    string          `json:"node_name"`
	TaskType    string          `json:"task_type"`
	AssigneeID  *uuid.UUID      `json:"assignee_id"`
	Status      int             `json:"status"`
	Action      string          `json:"action"`
	Comment     string          `json:"comment"`
	FormData    json.RawMessage `json:"form_data"`
	DueDate     *time.Time      `json:"due_date"`
	ClaimedAt   *time.Time      `json:"claimed_at"`
	CompletedAt *time.Time      `json:"completed_at"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// HistoryResponse is the response for a flow history entry.
type HistoryResponse struct {
	ID         uuid.UUID  `json:"id"`
	InstanceID uuid.UUID  `json:"instance_id"`
	TaskID     *uuid.UUID `json:"task_id"`
	NodeID     string     `json:"node_id"`
	NodeName   string     `json:"node_name"`
	NodeType   string     `json:"node_type"`
	OperatorID *uuid.UUID `json:"operator_id"`
	Action     string     `json:"action"`
	Comment    string     `json:"comment"`
	FromNodeID string     `json:"from_node_id"`
	ToNodeID   string     `json:"to_node_id"`
	CreatedAt  time.Time  `json:"created_at"`
}

// ========== DTO Conversion Functions ==========

// ToDefinitionResponse converts FlowDefinition model to DefinitionResponse.
func ToDefinitionResponse(def *model.FlowDefinition) DefinitionResponse {
	var draft *GraphData
	if len(def.DraftGraph) > 0 {
		var graph GraphData
		if err := json.Unmarshal(def.DraftGraph, &graph); err == nil {
			draft = &graph
		}
	}
	return DefinitionResponse{
		ID:          def.ID,
		Key:         def.Key,
		Name:        def.Name,
		Description: def.Description,
		Category:    def.Category,
		Status:      def.Status,
		DraftGraph:  draft,
		CreatedBy:   def.CreatedBy,
		UpdatedBy:   def.UpdatedBy,
		CreatedAt:   def.CreatedAt,
		UpdatedAt:   def.UpdatedAt,
	}
}

// ToVersionResponse converts FlowDefinitionVersion model to VersionResponse.
func ToVersionResponse(ver *model.FlowDefinitionVersion) VersionResponse {
	var graphData GraphData
	if len(ver.BpmnData) > 0 {
		_ = json.Unmarshal(ver.BpmnData, &graphData)
	}
	return VersionResponse{
		ID:           ver.ID,
		DefinitionID: ver.DefinitionID,
		Version:      ver.Version,
		BpmnData:     graphData,
		Status:       ver.Status,
		PublishedBy:  ver.PublishedBy,
		PublishedAt:  ver.PublishedAt,
		CreatedAt:    ver.CreatedAt,
	}
}

// ToInstanceResponse converts FlowInstance model to InstanceResponse.
func ToInstanceResponse(inst *model.FlowInstance) InstanceResponse {
	var currentNodeIDs []string
	if len(inst.CurrentNodeIDs) > 0 {
		_ = json.Unmarshal(inst.CurrentNodeIDs, &currentNodeIDs)
	}
	return InstanceResponse{
		ID:                  inst.ID,
		DefinitionID:        inst.DefinitionID,
		DefinitionVersionID: inst.DefinitionVersionID,
		BusinessKey:         inst.BusinessKey,
		BusinessType:        inst.BusinessType,
		InitiatorID:         inst.InitiatorID,
		Status:              inst.Status,
		CurrentNodeIDs:      currentNodeIDs,
		StartedAt:           inst.StartedAt,
		EndedAt:             inst.EndedAt,
		TerminateReason:     inst.TerminateReason,
		CreatedAt:           inst.CreatedAt,
		UpdatedAt:           inst.UpdatedAt,
	}
}

// ToTaskResponse converts FlowTask model to TaskResponse.
func ToTaskResponse(task *model.FlowTask) TaskResponse {
	var formData json.RawMessage
	if len(task.FormData) > 0 {
		formData = json.RawMessage(task.FormData)
	}
	return TaskResponse{
		ID:          task.ID,
		InstanceID:  task.InstanceID,
		NodeID:      task.NodeID,
		NodeName:    task.NodeName,
		TaskType:    task.TaskType,
		AssigneeID:  task.AssigneeID,
		Status:      task.Status,
		Action:      task.Action,
		Comment:     task.Comment,
		FormData:    formData,
		DueDate:     task.DueDate,
		ClaimedAt:   task.ClaimedAt,
		CompletedAt: task.CompletedAt,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}
}

// ToHistoryResponse converts FlowHistory model to HistoryResponse.
func ToHistoryResponse(h *model.FlowHistory) HistoryResponse {
	return HistoryResponse{
		ID:         h.ID,
		InstanceID: h.InstanceID,
		TaskID:     h.TaskID,
		NodeID:     h.NodeID,
		NodeName:   h.NodeName,
		NodeType:   h.NodeType,
		OperatorID: h.OperatorID,
		Action:     h.Action,
		Comment:    h.Comment,
		FromNodeID: h.FromNodeID,
		ToNodeID:   h.ToNodeID,
		CreatedAt:  h.CreatedAt,
	}
}
