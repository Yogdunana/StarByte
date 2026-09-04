package service

import (
	"context"

	wfmodel "github.com/Yogdunana/StarByte/backend/internal/workflow/model"
	wfsvc "github.com/Yogdunana/StarByte/backend/internal/workflow/service"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
)

const officerInterviewKey = "officer_interview"

// InterviewStarter 启动干事面试流程。
type InterviewStarter interface {
	StartOfficerInterview(ctx context.Context, applicationID, initiatorID uuid.UUID, vars map[string]interface{}) (uuid.UUID, error)
}

type definitionLookup interface {
	GetByKey(ctx context.Context, key string) (*wfmodel.FlowDefinition, error)
}

type engineInterviewStarter struct {
	defs definitionLookup
	inst wfsvc.InstanceService
}

// NewInterviewStarter 用已发布的 officer_interview 定义启动实例。
func NewInterviewStarter(defs definitionLookup, inst wfsvc.InstanceService) InterviewStarter {
	if defs == nil || inst == nil {
		return nil
	}
	return &engineInterviewStarter{defs: defs, inst: inst}
}

func (s *engineInterviewStarter) StartOfficerInterview(ctx context.Context, applicationID, initiatorID uuid.UUID, vars map[string]interface{}) (uuid.UUID, error) {
	def, err := s.defs.GetByKey(ctx, officerInterviewKey)
	if err != nil {
		return uuid.Nil, response.NewAppErrorf(response.CodeWorkflowNotFound, "查询面试流程失败: %v", err)
	}
	if def == nil {
		return uuid.Nil, response.NewAppError(response.CodeWorkflowNotFound, "未找到已发布的干事面试流程 officer_interview")
	}
	inst, err := s.inst.Start(ctx, def.ID, applicationID.String(), "member_application", initiatorID, vars)
	if err != nil {
		return uuid.Nil, err
	}
	return inst.ID, nil
}
