package nodes

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/engine"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/model"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/events"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

// ApprovalNode handles the "approval" node type.
// When entered, it creates a FlowTask for the assignee(s) and pauses the flow.
// When the task is completed, the flow resumes to the next node.
type ApprovalNode struct {
	BusinessApprovers engine.BusinessApprover
	TaskRepo          repo.TaskRepo
	EventBus          *events.EventBus
	Approvers         repo.ApproverRepo
}

func (n *ApprovalNode) Type() string            { return "approval" }
func (n *ApprovalNode) WaitForCompletion() bool { return true }

func (n *ApprovalNode) Execute(ctx context.Context, inst *model.FlowInstance, node *engine.FlowNode, graph *engine.FlowGraph, vars map[string]interface{}) ([]string, error) {
	// Execute is called after task completion. Return next edges.
	edges := graph.GetNextNodes(node.ID, "")
	result := make([]string, 0, len(edges))
	for _, e := range edges {
		result = append(result, e.Target)
	}
	return result, nil
}

func (n *ApprovalNode) OnEnter(ctx context.Context, inst *model.FlowInstance, node *engine.FlowNode, vars map[string]interface{}) error {
	config := n.parseConfig(node)
	assignees := n.resolveAssignees(config, inst, vars)
	if config["assigneeStrategy"] == "business_role" {
		if n.BusinessApprovers == nil {
			return response.NewAppError(response.CodeWorkflowInvalidNode, "业务审批人解析器未配置")
		}
		var err error
		assignees, err = n.BusinessApprovers.Resolve(ctx, inst, node)
		if err != nil {
			return err
		}
	} else if n.Approvers != nil {
		var err error
		assignees, err = n.resolveRuntime(ctx, config, inst.InitiatorID, vars, assignees)
		if err != nil {
			return err
		}
	}
	unique := []uuid.UUID{}
	seen := map[uuid.UUID]bool{}
	for _, id := range assignees {
		if id != uuid.Nil && !seen[id] {
			seen[id] = true
			unique = append(unique, id)
		}
	}
	assignees = unique
	if config["approvalType"] == "single" && len(assignees) != 1 {
		return response.NewAppError(response.CodeWorkflowInvalidNode, "单人审批必须且只能有一名处理人")
	}
	if len(assignees) == 0 {
		return response.NewAppError(response.CodeWorkflowInvalidNode,
			"审批节点没有处理人")
	}

	activation := uuid.New()
	for _, assigneeID := range assignees {
		task := &model.FlowTask{
			ActivationID: &activation,
			ID:           uuid.New(),
			InstanceID:   inst.ID,
			NodeID:       node.ID,
			NodeName:     node.Label,
			TaskType:     "approval",
			AssigneeID:   &assigneeID,
			Status:       0,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		if dueDays, ok := config["dueDays"].(float64); ok && dueDays > 0 {
			due := time.Now().Add(time.Duration(dueDays) * 24 * time.Hour)
			task.DueDate = &due
		}

		if err := n.TaskRepo.CreateTask(ctx, nil, task); err != nil {
			return err
		}

		// Publish TaskCreatedEvent.
		n.EventBus.Publish(ctx, events.TaskCreatedEvent{
			InstanceID: inst.ID,
			TaskID:     task.ID,
			AssigneeID: assigneeID,
			NodeID:     node.ID,
			NodeName:   node.Label,
			TaskType:   "approval",
		})
	}

	return nil
}

func (n *ApprovalNode) OnLeave(ctx context.Context, inst *model.FlowInstance, node *engine.FlowNode, vars map[string]interface{}) error {
	return nil
}

func (n *ApprovalNode) Validate(node *engine.FlowNode) error {
	config := n.parseConfig(node)
	if config == nil {
		return response.NewAppError(response.CodeWorkflowInvalidNode,
			"审批节点缺少配置")
	}

	strategy, ok := config["assigneeStrategy"].(string)
	if !ok || strategy == "" {
		return response.NewAppError(response.CodeWorkflowInvalidNode,
			"审批节点缺少 assigneeStrategy 配置")
	}

	// Validate strategy-specific fields.
	switch strategy {
	case "business_role":
		switch config["businessType"] {
		case "member_application":
			if config["admissionRole"] == nil || config["admissionStage"] == nil {
				return response.NewAppError(response.CodeWorkflowInvalidNode, "业务审批缺少入会环节与签字职务")
			}
		case engine.TaskTransferBusinessType:
			if config["transferStage"] != "handover" || config["transferRole"] == nil {
				return response.NewAppError(response.CodeWorkflowInvalidNode, "转办签字节点缺少必要的签字要求")
			}
		case engine.TaskBusinessType:
			switch config["taskStage"] {
			case "assignment", "execution", "review", "acceptance":
			default:
				return response.NewAppError(response.CodeWorkflowInvalidNode, "无效的任务审批环节")
			}
		default:
			return response.NewAppError(response.CodeWorkflowInvalidNode, "不支持的业务审批类型")
		}
	case "static":
		assignees, ok := config["assignees"].([]interface{})
		if !ok || len(assignees) == 0 {
			return response.NewAppError(response.CodeWorkflowInvalidNode,
				"静态处理人策略需要非空的 assignees 列表")
		}
	case "role":
		roleID, _ := config["roleId"].(string)
		roleCode, _ := config["roleCode"].(string)
		if roleID == "" && roleCode == "" {
			return response.NewAppError(response.CodeWorkflowInvalidNode,
				"角色处理人策略需要 roleId 或 roleCode 配置")
		}
	case "dept_leader", "initiator":
		// No additional fields required.
	default:
		return response.NewAppErrorf(response.CodeWorkflowNodeType,
			"不支持的处理人策略: %s", strategy)
	}

	return engine.ValidateApprovalConfig(config)
}

func (n *ApprovalNode) parseConfig(node *engine.FlowNode) map[string]interface{} {
	if node.Config == nil {
		return nil
	}
	return node.Config
}

func (n *ApprovalNode) resolveAssignees(config map[string]interface{}, inst *model.FlowInstance, vars map[string]interface{}) []uuid.UUID {
	var result []uuid.UUID

	strategy, _ := config["assigneeStrategy"].(string)
	switch strategy {
	case "static":
		if assignees, ok := config["assignees"].([]interface{}); ok {
			for _, a := range assignees {
				if s, ok := a.(string); ok {
					if id, err := uuid.Parse(s); err == nil {
						result = append(result, id)
					}
				}
			}
		}
	case "initiator":
		result = append(result, inst.InitiatorID)
	case "role", "dept_leader":
		// Active role holders are resolved through ApproverRepo at runtime.
	}

	return result
}
