package service

import (
	"context"
	"fmt"

	"github.com/Yogdunana/StarByte/backend/internal/announcement/dto"
	"github.com/Yogdunana/StarByte/backend/internal/announcement/model"
	"github.com/Yogdunana/StarByte/backend/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (s *announcementService) Publish(ctx context.Context, viewer Viewer, id uuid.UUID) (*dto.AnnouncementResponse, error) {
	a, err := s.load(ctx, id)
	if err != nil {
		return nil, err
	}
	if !viewer.Staff && viewer.UserID != a.AuthorID {
		return nil, noAccess("无权发布该公告")
	}
	if a.Status != model.StatusDraft {
		return nil, invalidState("只有草稿可以发布")
	}
	if err := s.publishLocked(ctx, a); err != nil {
		return nil, err
	}
	return s.getResponse(ctx, a.ID, viewer.UserID)
}

func (s *announcementService) publishLocked(ctx context.Context, a *model.Announcement) error {
	now := s.clock()
	a.Status = model.StatusPublished
	a.PublishedAt = &now
	a.ScheduledAt = nil
	a.UpdatedAt = now
	if err := s.rows.Update(ctx, a); err != nil {
		return fmt.Errorf("publish announcement: %w", err)
	}
	s.notifyPublished(ctx, a)
	return nil
}

func (s *announcementService) Pin(ctx context.Context, viewer Viewer, id uuid.UUID, pinned *bool) (*dto.AnnouncementResponse, error) {
	if !viewer.Staff {
		return nil, noAccess("无权置顶公告")
	}
	a, err := s.load(ctx, id)
	if err != nil {
		return nil, err
	}
	if pinned == nil {
		a.Pinned = !a.Pinned
	} else {
		a.Pinned = *pinned
	}
	a.UpdatedAt = s.clock()
	if err := s.rows.Update(ctx, a); err != nil {
		return nil, fmt.Errorf("pin announcement: %w", err)
	}
	return s.getResponse(ctx, a.ID, viewer.UserID)
}

func (s *announcementService) Archive(ctx context.Context, viewer Viewer, id uuid.UUID) (*dto.AnnouncementResponse, error) {
	if !viewer.Staff {
		return nil, noAccess("无权归档公告")
	}
	a, err := s.load(ctx, id)
	if err != nil {
		return nil, err
	}
	if a.Status != model.StatusPublished {
		return nil, invalidState("只有已发布公告可以归档")
	}
	a.Status = model.StatusArchived
	a.UpdatedAt = s.clock()
	if err := s.rows.Update(ctx, a); err != nil {
		return nil, fmt.Errorf("archive announcement: %w", err)
	}
	return s.getResponse(ctx, a.ID, viewer.UserID)
}

func (s *announcementService) DispatchDuePublishes(ctx context.Context, _ string, logf func(string)) error {
	if logf == nil {
		logf = func(string) {}
	}
	rows, err := s.rows.ListDueDrafts(ctx, s.clock(), 50)
	if err != nil {
		return fmt.Errorf("list due announcement drafts: %w", err)
	}
	published := 0
	for i := range rows {
		a := &rows[i]
		if err := s.publishLocked(ctx, a); err != nil {
			logger.Warn("scheduled announcement publish failed", zap.Error(err), zap.String("id", a.ID.String()))
			continue
		}
		published++
	}
	logf(fmt.Sprintf("published %d scheduled announcements", published))
	return nil
}
