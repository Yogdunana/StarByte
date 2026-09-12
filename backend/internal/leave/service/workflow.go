package service

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/engine"
	wfmodel "github.com/Yogdunana/StarByte/backend/internal/workflow/model"
)

// LeaveFlow is the transaction-bound engine surface used by leave mutations.
type LeaveFlow interface {
	Start(ctx context.Context, key, businessKey, businessType string, initiator uuid.UUID, vars map[string]interface{}) (*wfmodel.FlowInstance, error)
	CompleteLeaveApproval(ctx context.Context, instanceID, reviewer uuid.UUID, action engine.TaskAction, comment string) error
	RunningApprovalNode(ctx context.Context, id uuid.UUID) (string, bool, error)
}

// LeaveEngine binds the generic engine into the same leave transaction.
type LeaveEngine interface {
	BindTransaction(tx *gorm.DB) (LeaveFlow, func(context.Context), error)
}

type engineAdapter struct{ inner *engine.FlowEngine }

// AdaptEngine wraps the process engine for leave start/approve.
func AdaptEngine(flow *engine.FlowEngine) LeaveEngine {
	if flow == nil {
		return nil
	}
	return engineAdapter{inner: flow}
}

func (a engineAdapter) BindTransaction(tx *gorm.DB) (LeaveFlow, func(context.Context), error) {
	bound, deliver, err := a.inner.BindTransaction(tx)
	if err != nil {
		return nil, nil, err
	}
	return bound, deliver, nil
}
