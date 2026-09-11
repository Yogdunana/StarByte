package engine

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/model"
)

// BusinessApprover delegates role resolution to the owning business module.
// Implementations must use the same transaction and exclude the applicant.
type BusinessApprover interface {
	Resolve(context.Context, *model.FlowInstance, *FlowNode) ([]uuid.UUID, error)
	ForTransaction(*gorm.DB) BusinessApprover
}
