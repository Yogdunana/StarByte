package model

import (
	"context"

	"github.com/google/uuid"

	rbacModel "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
)

type Viewer struct {
	ID    uuid.UUID
	Scope *rbacModel.DataScopeCondition
}
type viewerKey struct{}

func WithViewer(ctx context.Context, viewer Viewer) context.Context {
	return context.WithValue(ctx, viewerKey{}, viewer)
}
func ViewerFromContext(ctx context.Context) Viewer {
	viewer, _ := ctx.Value(viewerKey{}).(Viewer)
	return viewer
}

type ApproverOption struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	DepartmentName string    `json:"department_name"`
}
