package dto

import (
	"time"

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
	GraphData GraphData `json:"graph_data" binding:"required"`
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
	ID     string `json:"id"`
	Source string `json:"source"`
	Target string `json:"target"`
}

// DefinitionResponse is the response for a single flow definition.
type DefinitionResponse struct {
	ID          uuid.UUID  `json:"id"`
	Key         string     `json:"key"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Category    string     `json:"category"`
	Status      int        `json:"status"`
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
	ID          uuid.UUID  `json:"id"`
	InstanceID  uuid.UUID  `json:"instance_id"`
	NodeID      string     `json:"node_id"`
	NodeName    string     `json:"node_name"`
	TaskType    string     `json:"task_type"`
	AssigneeID  *uuid.UUID `json:"assignee_id"`
	Status      int        `json:"status"`
	Action      string     `json:"action"`
	Comment     string     `json:"comment"`
	FormData    []byte     `json:"form_data"`
	DueDate     *time.Time `json:"due_date"`
	ClaimedAt   *time.Time `json:"claimed_at"`
	CompletedAt *time.Time `json:"completed_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
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
