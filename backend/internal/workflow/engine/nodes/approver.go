package nodes

import (
	"context"

	"github.com/google/uuid"

	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

func (n *ApprovalNode) resolveRuntime(ctx context.Context, config map[string]interface{}, initiator uuid.UUID, vars map[string]interface{}, fallback []uuid.UUID) ([]uuid.UUID, error) {
	switch config["assigneeStrategy"] {
	case "role":
		return n.resolveRoleAssignees(ctx, config, vars)
	case "dept_leader":
		return n.Approvers.DepartmentLeaders(ctx, initiator)
	default:
		users, err := n.Approvers.ActiveUsers(ctx, fallback)
		if err != nil {
			return nil, err
		}
		unique := map[uuid.UUID]bool{}
		for _, id := range fallback {
			unique[id] = true
		}
		if len(users) != len(unique) {
			return nil, response.NewAppError(response.CodeWorkflowInvalidNode, "审批人不存在或已停用，请更新流程配置")
		}
		return users, nil
	}
}

func (n *ApprovalNode) resolveRoleAssignees(ctx context.Context, config map[string]interface{}, vars map[string]interface{}) ([]uuid.UUID, error) {
	if n.Approvers == nil {
		return nil, response.NewAppError(response.CodeWorkflowInvalidNode, "角色审批人解析器未配置")
	}
	if code, _ := config["roleCode"].(string); code != "" {
		department := scopedDepartment(config, vars)
		if required, _ := config["requireDepartment"].(bool); required && department == nil {
			return nil, response.NewAppError(response.CodeWorkflowInvalidNode, "审批缺少部门或中心范围")
		}
		return n.Approvers.ByRoleCode(ctx, code, department)
	}
	role, _ := config["roleId"].(string)
	id, err := uuid.Parse(role)
	if err != nil {
		return nil, response.NewAppError(response.CodeWorkflowInvalidNode, "角色 ID 无效")
	}
	return n.Approvers.ByRole(ctx, id)
}

func scopedDepartment(config, vars map[string]interface{}) *uuid.UUID {
	scoped, _ := config["departmentScope"].(bool)
	if !scoped || vars == nil {
		return nil
	}
	key := "department_id"
	if variable, _ := config["departmentVariable"].(string); variable != "" {
		key = variable
	}
	raw, _ := vars[key].(string)
	id, err := uuid.Parse(raw)
	if err != nil {
		return nil
	}
	return &id
}
