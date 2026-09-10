package model

import (
	"context"

	"github.com/google/uuid"

	rbac "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
)

type Viewer struct {
	CanUpdate   bool
	CanDelete   bool
	UpdateScope *rbac.DataScopeCondition
	DeleteScope *rbac.DataScopeCondition
	CanManage   bool
	ManageScope *rbac.DataScopeCondition
	ID          uuid.UUID
	Scope       *rbac.DataScopeCondition
	Manage      bool
}
type viewerKey struct{}

func WithViewer(ctx context.Context, viewer Viewer) context.Context {
	return context.WithValue(ctx, viewerKey{}, viewer)
}
func ViewerFromContext(ctx context.Context) Viewer {
	viewer, _ := ctx.Value(viewerKey{}).(Viewer)
	return viewer
}
