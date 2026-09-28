package engine

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/model"
	"github.com/Yogdunana/StarByte/backend/pkg/events"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

// EscalateOverdueApproval hands a stalled approval node to reviewers so a node
// that nobody picked up within its due date stops blocking the flow.
//
// Every still-pending task of the node is cancelled and one fresh task per
// reviewer opens under a new activation. Resetting the activation matters: the
// quorum is counted per activation, so countersign nodes recover unanimity
// instead of waiting for people who never answered.
//
// The new tasks carry no due date on purpose — a fresh due date would make the
// next scan escalate again and again for the same node.
func (e *FlowEngine) EscalateOverdueApproval(ctx context.Context, todo OverdueApprovalTodo, reviewers []uuid.UUID, operator uuid.UUID) (int, error) {
	inst, err := e.instRepo.GetByID(ctx, todo.InstanceID)
	if err != nil {
		return 0, err
	}
	if inst == nil {
		return 0, response.NewError(response.CodeWorkflowInstNotFound, "流程实例不存在")
	}
	if inst.Status != 0 {
		return 0, nil
	}
	tasks, err := e.taskRepo.ListTasksByInstance(ctx, inst.ID)
	if err != nil {
		return 0, err
	}
	pending := 0
	for _, item := range tasks {
		if item.NodeID != todo.NodeID {
			continue
		}
		switch item.Status {
		case 0:
			pending++
		case 1, 2:
			// Somebody already took a side; ordinary any / all arithmetic can
			// still resolve the node, so nobody needs rescuing.
			return 0, nil
		}
	}
	if pending == 0 {
		return 0, nil
	}
	// Idempotent per node: a later scan must not hand the same stall over twice.
	history, err := e.taskRepo.ListHistory(ctx, inst.ID)
	if err != nil {
		return 0, err
	}
	for _, item := range history {
		if item.NodeID == todo.NodeID && item.Action == "timeout_escalate" {
			return 0, nil
		}
	}
	unique := map[uuid.UUID]bool{}
	seeded := []uuid.UUID{}
	for _, id := range reviewers {
		if id == uuid.Nil || unique[id] {
			continue
		}
		if inst.BusinessType == "member_application" && id == inst.InitiatorID {
			continue
		}
		unique[id] = true
		seeded = append(seeded, id)
	}
	if len(seeded) == 0 {
		return 0, nil
	}
	reason := "审批超时未处理，已转交升级处理人"
	if err := e.cancelNodePendingTasks(ctx, inst.ID, todo.NodeID, operator, reason); err != nil {
		return 0, err
	}
	now := time.Now()
	activation := uuid.New()
	for _, assigneeID := range seeded {
		task := &model.FlowTask{
			ID:           uuid.New(),
			InstanceID:   inst.ID,
			NodeID:       todo.NodeID,
			NodeName:     todo.NodeName,
			TaskType:     "approval",
			ActivationID: &activation,
			AssigneeID:   &assigneeID,
			Status:       0,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		if err := e.taskRepo.CreateTask(ctx, nil, task); err != nil {
			return 0, err
		}
		hist := &model.FlowHistory{
			ID: uuid.New(), InstanceID: inst.ID, TaskID: &task.ID, NodeID: todo.NodeID,
			NodeName: todo.NodeName, NodeType: "approval", Action: "timeout_escalate",
			Comment: reason, CreatedAt: now,
		}
		if operator != uuid.Nil {
			hist.OperatorID = &operator
		}
		if err := e.taskRepo.CreateHistory(ctx, nil, hist); err != nil {
			return 0, err
		}
		_ = e.eventBus.Publish(ctx, events.TaskCreatedEvent{
			BusinessType: inst.BusinessType,
			BusinessKey:  inst.BusinessKey,
			InstanceID:   inst.ID,
			TaskID:       task.ID,
			AssigneeID:   assigneeID,
			NodeID:       todo.NodeID,
			NodeName:     todo.NodeName,
			TaskType:     "approval",
		})
	}
	return len(seeded), nil
}
