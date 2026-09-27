package engine

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

// OverdueCollaborationTodo is a pending collaboration-task todo past its node due date.
type OverdueCollaborationTodo struct {
	InstanceID  uuid.UUID
	FlowTaskID  uuid.UUID
	BusinessKey string
	Stage       string
	AssigneeID  *uuid.UUID
	DueDate     time.Time
}

type overdueScanRow struct {
	ID                  uuid.UUID  `gorm:"column:id"`
	InstanceID          uuid.UUID  `gorm:"column:instance_id"`
	AssigneeID          *uuid.UUID `gorm:"column:assignee_id"`
	DueDate             time.Time  `gorm:"column:due_date"`
	NodeID              string     `gorm:"column:node_id"`
	BusinessKey         string     `gorm:"column:business_key"`
	DefinitionVersionID uuid.UUID  `gorm:"column:definition_version_id"`
}

// OverdueApprovalTodo is a pending approval past its node due date on any
// business type. The collaboration scan above hard-codes task flows, so
// admissions and leaves configured with dueDays never reached a timeout handler.
type OverdueApprovalTodo struct {
	InstanceID   uuid.UUID
	FlowTaskID   uuid.UUID
	BusinessType string
	BusinessKey  string
	NodeID       string
	NodeName     string
	AssigneeID   *uuid.UUID
	DueDate      time.Time
}

type overdueApprovalRow struct {
	ID           uuid.UUID  `gorm:"column:id"`
	InstanceID   uuid.UUID  `gorm:"column:instance_id"`
	AssigneeID   *uuid.UUID `gorm:"column:assignee_id"`
	DueDate      time.Time  `gorm:"column:due_date"`
	NodeID       string     `gorm:"column:node_id"`
	NodeName     string     `gorm:"column:node_name"`
	BusinessType string     `gorm:"column:business_type"`
	BusinessKey  string     `gorm:"column:business_key"`
}

// ListOverdueApprovalTodos scans every running instance for pending approvals
// past their due date. Pass businessTypes to narrow it down; an empty list
// covers all flows.
func (e *FlowEngine) ListOverdueApprovalTodos(ctx context.Context, now time.Time, businessTypes ...string) ([]OverdueApprovalTodo, error) {
	if e.db == nil {
		return nil, nil
	}
	query := e.db.WithContext(ctx).Table("flow_tasks AS ft").
		Select("ft.id, ft.instance_id, ft.assignee_id, ft.due_date, ft.node_id, ft.node_name, fi.business_type, fi.business_key").
		Joins("JOIN flow_instances fi ON fi.id = ft.instance_id").
		Where("fi.status = 0 AND ft.status = 0 AND ft.due_date IS NOT NULL AND ft.due_date < ?", now)
	if len(businessTypes) > 0 {
		query = query.Where("fi.business_type IN ?", businessTypes)
	}
	rows := []overdueApprovalRow{}
	if err := query.Order("ft.due_date, ft.id").Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]OverdueApprovalTodo, 0, len(rows))
	for _, row := range rows {
		out = append(out, OverdueApprovalTodo{
			InstanceID:   row.InstanceID,
			FlowTaskID:   row.ID,
			BusinessType: row.BusinessType,
			BusinessKey:  row.BusinessKey,
			NodeID:       row.NodeID,
			NodeName:     row.NodeName,
			AssigneeID:   row.AssigneeID,
			DueDate:      row.DueDate,
		})
	}
	return out, nil
}

func (e *FlowEngine) ListOverdueCollaborationTodos(ctx context.Context, now time.Time) ([]OverdueCollaborationTodo, error) {
	if e.db == nil {
		return nil, nil
	}
	rows := []overdueScanRow{}
	err := e.db.WithContext(ctx).Table("flow_tasks AS ft").
		Select("ft.id, ft.instance_id, ft.assignee_id, ft.due_date, ft.node_id, fi.business_key, fi.definition_version_id").
		Joins("JOIN flow_instances fi ON fi.id = ft.instance_id").
		Where("fi.business_type = ? AND fi.status = 0 AND ft.status = 0 AND ft.due_date IS NOT NULL AND ft.due_date < ?", TaskBusinessType, now).
		Order("ft.due_date, ft.id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	graphs := map[uuid.UUID]*FlowGraph{}
	out := make([]OverdueCollaborationTodo, 0, len(rows))
	for _, row := range rows {
		graph, ok := graphs[row.DefinitionVersionID]
		if !ok {
			version, err := e.defRepo.GetVersionByID(ctx, row.DefinitionVersionID)
			if err != nil {
				return nil, err
			}
			if version == nil {
				return nil, response.NewError(response.CodeWorkflowVerNotFound, "任务流程版本不存在")
			}
			graph, err = ParseGraph(version.BpmnData)
			if err != nil {
				return nil, err
			}
			graphs[row.DefinitionVersionID] = graph
		}
		stage := row.NodeID
		if node := graph.GetNode(row.NodeID); node != nil {
			if value, _ := node.Config["taskStage"].(string); value != "" {
				stage = value
			}
		}
		out = append(out, OverdueCollaborationTodo{
			InstanceID:  row.InstanceID,
			FlowTaskID:  row.ID,
			BusinessKey: row.BusinessKey,
			Stage:       stage,
			AssigneeID:  row.AssigneeID,
			DueDate:     row.DueDate,
		})
	}
	return out, nil
}
