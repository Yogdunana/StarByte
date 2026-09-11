package engine

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

// BusinessApprovers isolates policies by business type; adding a module cannot
// replace another module's resolver. Transaction copies bind every policy.
type BusinessApprovers struct {
	mu        sync.RWMutex
	resolvers map[string]BusinessApprover
}

func NewBusinessApprovers() *BusinessApprovers {
	return &BusinessApprovers{resolvers: map[string]BusinessApprover{}}
}
func (r *BusinessApprovers) Register(kind string, resolver BusinessApprover) error {
	if kind == "" || resolver == nil {
		return fmt.Errorf("business type and resolver are required")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.resolvers[kind]; exists {
		return fmt.Errorf("business resolver already registered: %s", kind)
	}
	r.resolvers[kind] = resolver
	return nil
}
func (r *BusinessApprovers) Resolve(ctx context.Context, inst *model.FlowInstance, node *FlowNode) ([]uuid.UUID, error) {
	if inst == nil || node == nil || node.Config["businessType"] != inst.BusinessType {
		return nil, response.NewError(response.CodeWorkflowInvalidNode, "审批节点与业务类型不一致")
	}
	r.mu.RLock()
	resolver := r.resolvers[inst.BusinessType]
	r.mu.RUnlock()
	if resolver == nil {
		return nil, response.NewError(response.CodeWorkflowInvalidNode, "该业务的审批人解析器未配置")
	}
	return resolver.Resolve(ctx, inst, node)
}
func (r *BusinessApprovers) ForTransaction(tx *gorm.DB) BusinessApprover {
	out := NewBusinessApprovers()
	r.mu.RLock()
	defer r.mu.RUnlock()
	for kind, resolver := range r.resolvers {
		out.resolvers[kind] = resolver.ForTransaction(tx)
	}
	return out
}
