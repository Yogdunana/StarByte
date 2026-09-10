package model

import (
	"context"

	"github.com/google/uuid"

	rbac "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
)

type Viewer struct {
	Allowed map[string]bool
	Scopes  map[string]*rbac.DataScopeCondition
	ID      uuid.UUID
	Scope   *rbac.DataScopeCondition
}
type viewerKey struct{}

func WithViewer(ctx context.Context, viewer Viewer) context.Context {
	return context.WithValue(ctx, viewerKey{}, viewer)
}
func ViewerFromContext(ctx context.Context) (Viewer, bool) {
	v, ok := ctx.Value(viewerKey{}).(Viewer)
	return v, ok
}
