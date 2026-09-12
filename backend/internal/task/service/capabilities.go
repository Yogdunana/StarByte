package service

import (
	"context"

	"github.com/Yogdunana/StarByte/backend/internal/task/dto"
	"github.com/Yogdunana/StarByte/backend/internal/task/model"
)

func taskCapabilities(ctx context.Context, t *model.Task, out *dto.TaskResponse) {
	v, ok := model.ViewerFromContext(ctx)
	if !ok {
		return
	}
	can := func(permission string) bool {
		return v.Allowed[permission] && canViewTask(t, v.ID, v.Scopes[permission])
	}
	participant := t.CreatorID == v.ID || (t.AssigneeID != nil && *t.AssigneeID == v.ID)
	live := !model.IsClosed(t.Status)
	out.CanUpdate = live && participant && can("task:update")
	out.CanCancel = out.CanUpdate && (t.Status == model.StatusPending || t.Status == model.StatusDoing)
	out.CanDelete = t.Status != model.StatusDoing && participant && can("task:delete")
	out.CanAssign = live && can("task:assign")
	out.CanTransfer = live && participant && can("task:transfer")
	out.CanComment = live && can("task:comment")
	if t.WorkflowInstanceID != nil {
		out.CanUpdate = out.CanUpdate && (t.WorkflowStage == "assignment" || t.WorkflowStage == "execution")
		out.CanDelete = out.CanDelete && !live
		out.CanAssign = out.CanAssign && t.WorkflowStage == "assignment" && t.CreatorID == v.ID
		out.CanTransfer = live && t.WorkflowStage == "execution" && t.AssigneeID != nil && *t.AssigneeID == v.ID && can("task:transfer")
		out.CanCancel = out.CanCancel && t.CreatorID == v.ID
	}
	out.CanUrge = live && t.CreatorID == v.ID && t.AssigneeID != nil && can("task:create")
}
