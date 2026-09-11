package service

import (
	"context"

	"github.com/google/uuid"

	wfmodel "github.com/Yogdunana/StarByte/backend/internal/workflow/model"
)

// TaskWorkflowRuntime keeps task policy separate from the generic engine.
// Production calls always use an engine bound to the task transaction.
type TaskWorkflowRuntime interface {
	CompleteTaskTransferApproval(context.Context, uuid.UUID, string, uuid.UUID, uuid.UUID, bool) error
	ReassignTaskExecution(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID, string) error

	Start(context.Context, string, string, string, uuid.UUID, map[string]interface{}) (*wfmodel.FlowInstance, error)
	TaskCheckpoint(context.Context, uuid.UUID, string, uuid.UUID, string, string) error
	BusinessStage(context.Context, uuid.UUID) (string, bool, error)
	Terminate(context.Context, uuid.UUID, uuid.UUID, string) error
}
