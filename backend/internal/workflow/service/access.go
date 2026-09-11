package service

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Yogdunana/StarByte/backend/internal/workflow/model"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/repo"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

func requireInstanceAccess(ctx context.Context, db *gorm.DB, inst *model.FlowInstance, participants bool) error {
	viewer := model.ViewerFromContext(ctx)
	if viewer.ID == uuid.Nil {
		return response.NewAppError(response.CodeForbidden, "缺少流程访问身份")
	}
	if inst.InitiatorID == viewer.ID {
		return nil
	}
	if db != nil {
		allowed, err := repo.NewAccessRepo(db).CanRead(ctx, inst.ID, viewer, participants)
		if err != nil {
			return err
		}
		if allowed {
			return nil
		}
	}
	return response.NewAppError(response.CodeForbidden, "无权访问此流程")
}

func (s *definitionServiceImpl) requireDefinitionAccess(ctx context.Context, def *model.FlowDefinition) error {
	viewer := model.ViewerFromContext(ctx)
	if viewer.ID == uuid.Nil {
		return response.NewAppError(response.CodeForbidden, "缺少流程管理身份")
	}
	if def.CreatedBy != nil && *def.CreatedBy == viewer.ID {
		return nil
	}
	if s.db != nil {
		allowed, err := repo.NewAccessRepo(s.db).CanManageDefinition(ctx, def.ID, viewer)
		if err != nil {
			return err
		}
		if allowed {
			return nil
		}
	}
	return response.NewAppError(response.CodeForbidden, "无权管理此流程定义")
}
