package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/Yogdunana/StarByte/backend/internal/meeting/dto"
	"github.com/Yogdunana/StarByte/backend/internal/meeting/model"
	rbac "github.com/Yogdunana/StarByte/backend/internal/rbac/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

func (s *meetingService) requireMeetingAccess(ctx context.Context, id uuid.UUID) error {
	if s.access == nil {
		return nil
	} // In-memory unit repositories; production always binds AccessRepo.
	allowed, err := s.access.CanAccess(ctx, id, model.ViewerFromContext(ctx))
	if err != nil {
		return err
	}
	if !allowed {
		return response.NewError(response.CodeForbidden, "无权访问或管理该会议")
	}
	return nil
}

// canPreviewLiveResult is true only when the viewer holds meeting:manage for
// this meeting. A global CanManage flag (any department) is not enough.
func (s *meetingService) canPreviewLiveResult(ctx context.Context, meetingID uuid.UUID) (bool, error) {
	viewer := model.ViewerFromContext(ctx)
	if !viewer.CanManage {
		return false, nil
	}
	if s.access == nil {
		return true, nil
	}
	candidate := viewer
	candidate.Scope = viewer.ManageScope
	candidate.Manage = true
	return s.access.CanAccess(ctx, meetingID, candidate)
}

// Batch capabilities per page, using the same scope policy as mutation endpoints.
func (s *meetingService) meetingCapabilities(ctx context.Context, rows []*dto.MeetingResponse) error {
	if s.access == nil || len(rows) == 0 {
		return nil
	}
	viewer := model.ViewerFromContext(ctx)
	ids := make([]uuid.UUID, len(rows))
	for i, row := range rows {
		ids[i] = uuid.MustParse(row.ID)
	}
	actions := []struct {
		enabled bool
		scope   *rbac.DataScopeCondition
		set     func(*dto.MeetingResponse, bool)
	}{
		{viewer.CanManage, viewer.ManageScope, func(r *dto.MeetingResponse, v bool) { r.CanManage = v }},
		{viewer.CanUpdate, viewer.UpdateScope, func(r *dto.MeetingResponse, v bool) { r.CanUpdate = v }},
		{viewer.CanDelete, viewer.DeleteScope, func(r *dto.MeetingResponse, v bool) { r.CanDelete = v }},
	}
	for _, action := range actions {
		if !action.enabled {
			continue
		}
		candidate := viewer
		candidate.Scope = action.scope
		candidate.Manage = true
		allowed, err := s.access.AllowedIDs(ctx, ids, candidate)
		if err != nil {
			return fmt.Errorf("resolve meeting actions: %w", err)
		}
		for i, row := range rows {
			action.set(row, allowed[ids[i]])
		}
	}
	return nil
}
