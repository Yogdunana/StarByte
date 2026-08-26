package model

import (
	"time"

	"github.com/google/uuid"
)

// FlowDefinition represents a workflow process definition (the "template").
// It supports multi-version management via FlowDefinitionVersion.
//
// Status values: 0=draft, 1=published, 2=disabled.
type FlowDefinition struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	Key         string     `gorm:"type:varchar(100);uniqueIndex;not null" json:"key"`
	Name        string     `gorm:"type:varchar(200);not null" json:"name"`
	Description string     `gorm:"type:text" json:"description"`
	Category    string     `gorm:"type:varchar(50);default:custom" json:"category"`
	Status      int        `gorm:"type:smallint;default:0;index" json:"status"`
	CreatedBy   *uuid.UUID `gorm:"type:uuid" json:"created_by"`
	UpdatedBy   *uuid.UUID `gorm:"type:uuid" json:"updated_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// TableName overrides the default GORM table name.
func (FlowDefinition) TableName() string {
	return "flow_definitions"
}

// FlowDefinitionVersion represents a specific version of a flow definition.
// The bpmn_data field stores the complete graph structure (nodes + edges)
// compatible with React Flow.
//
// Status values: 0=historical, 1=current version.
type FlowDefinitionVersion struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	DefinitionID uuid.UUID  `gorm:"type:uuid;index;not null" json:"definition_id"`
	Version      int        `gorm:"not null" json:"version"`
	BpmnData     []byte     `gorm:"type:jsonb;not null" json:"bpmn_data"`
	Status       int        `gorm:"type:smallint;default:0;index" json:"status"`
	PublishedBy  *uuid.UUID `gorm:"type:uuid" json:"published_by"`
	PublishedAt  *time.Time `json:"published_at"`
	CreatedAt    time.Time  `json:"created_at"`
}

// TableName overrides the default GORM table name.
func (FlowDefinitionVersion) TableName() string {
	return "flow_definition_versions"
}

// FlowInstance represents a running (or completed) process instance.
//
// Status values: 0=running, 1=completed, 2=terminated, 3=suspended.
type FlowInstance struct {
	ID                  uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	DefinitionID        uuid.UUID  `gorm:"type:uuid;index;not null" json:"definition_id"`
	DefinitionVersionID uuid.UUID  `gorm:"type:uuid;not null" json:"definition_version_id"`
	BusinessKey         string     `gorm:"type:varchar(100)" json:"business_key"`
	BusinessType        string     `gorm:"type:varchar(50)" json:"business_type"`
	InitiatorID         uuid.UUID  `gorm:"type:uuid;index;not null" json:"initiator_id"`
	Status              int        `gorm:"type:smallint;default:0;index" json:"status"`
	CurrentNodeIDs      []byte     `gorm:"type:jsonb" json:"current_node_ids"`
	StartedAt           time.Time  `json:"started_at"`
	EndedAt             *time.Time `json:"ended_at"`
	TerminateReason     string     `gorm:"type:text" json:"terminate_reason"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// TableName overrides the default GORM table name.
func (FlowInstance) TableName() string {
	return "flow_instances"
}

// FlowTask represents a todo item within a flow instance, typically
// created when the process reaches an approval node.
//
// Status values: 0=pending, 1=approved, 2=rejected, 3=transferred,
// 4=withdrawn, 5=cancelled.
type FlowTask struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	InstanceID  uuid.UUID  `gorm:"type:uuid;index;not null" json:"instance_id"`
	NodeID      string     `gorm:"type:varchar(100);not null" json:"node_id"`
	NodeName    string     `gorm:"type:varchar(200)" json:"node_name"`
	TaskType    string     `gorm:"type:varchar(50);default:approval" json:"task_type"`
	AssigneeID  *uuid.UUID `gorm:"type:uuid;index" json:"assignee_id"`
	Status      int        `gorm:"type:smallint;default:0;index" json:"status"`
	Action      string     `gorm:"type:varchar(50)" json:"action"`
	Comment     string     `gorm:"type:text" json:"comment"`
	FormData    []byte     `gorm:"type:jsonb" json:"form_data"`
	DueDate     *time.Time `json:"due_date"`
	ClaimedAt   *time.Time `json:"claimed_at"`
	CompletedAt *time.Time `json:"completed_at"`
	CreatedAt   time.Time  `gorm:"index" json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// TableName overrides the default GORM table name.
func (FlowTask) TableName() string {
	return "flow_tasks"
}

// FlowHistory records every significant operation in a flow instance,
// providing a full audit trail of process execution.
type FlowHistory struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	InstanceID uuid.UUID  `gorm:"type:uuid;index;not null" json:"instance_id"`
	TaskID     *uuid.UUID `gorm:"type:uuid" json:"task_id"`
	NodeID     string     `gorm:"type:varchar(100)" json:"node_id"`
	NodeName   string     `gorm:"type:varchar(200)" json:"node_name"`
	NodeType   string     `gorm:"type:varchar(50)" json:"node_type"`
	OperatorID *uuid.UUID `gorm:"type:uuid;index" json:"operator_id"`
	Action     string     `gorm:"type:varchar(50);not null" json:"action"`
	Comment    string     `gorm:"type:text" json:"comment"`
	FromNodeID string     `gorm:"type:varchar(100)" json:"from_node_id"`
	ToNodeID   string     `gorm:"type:varchar(100)" json:"to_node_id"`
	CreatedAt  time.Time  `gorm:"index" json:"created_at"`
}

// TableName overrides the default GORM table name.
func (FlowHistory) TableName() string {
	return "flow_histories"
}

// FlowVariable stores runtime data for a flow instance, such as
// applicant info, approval comments, form data, etc.
// Condition branches use these variables to determine flow direction.
//
// Scope values: global, local.
type FlowVariable struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	InstanceID uuid.UUID `gorm:"type:uuid;index;not null" json:"instance_id"`
	Key        string    `gorm:"type:varchar(100);not null" json:"key"`
	Value      []byte    `gorm:"type:jsonb" json:"value"`
	Scope      string    `gorm:"type:varchar(20);default:global" json:"scope"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// TableName overrides the default GORM table name.
func (FlowVariable) TableName() string {
	return "flow_variables"
}
