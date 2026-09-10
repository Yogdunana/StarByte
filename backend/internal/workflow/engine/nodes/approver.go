package nodes

import (
	"context"

	"github.com/google/uuid"

	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

func (n *ApprovalNode) resolveRuntime(ctx context.Context, config map[string]interface{}, initiator uuid.UUID, fallback []uuid.UUID) ([]uuid.UUID, error) {
	switch config["assigneeStrategy"] {
	case "role":
		role, _ := config["roleId"].(string)
		id, err := uuid.Parse(role)
		if err != nil {
			return nil, response.NewAppError(response.CodeWorkflowInvalidNode, "角色 ID 无效")
		}
		return n.Approvers.ByRole(ctx, id)
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
