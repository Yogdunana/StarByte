package dto

import (
	"encoding/json"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/model"
)

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
		DefinitionName: inst.DefinitionName, InitiatorName: inst.InitiatorName,
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
		DefinitionName: task.DefinitionName, AssigneeName: task.AssigneeName, InstanceStatus: task.InstanceStatus, BusinessType: task.BusinessType, BusinessKey: task.BusinessKey,
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
		OperatorName: h.OperatorName,
		ID:           h.ID,
		InstanceID:   h.InstanceID,
		TaskID:       h.TaskID,
		NodeID:       h.NodeID,
		NodeName:     h.NodeName,
		NodeType:     h.NodeType,
		OperatorID:   h.OperatorID,
		Action:       h.Action,
		Comment:      h.Comment,
		FromNodeID:   h.FromNodeID,
		ToNodeID:     h.ToNodeID,
		CreatedAt:    h.CreatedAt,
	}
}
