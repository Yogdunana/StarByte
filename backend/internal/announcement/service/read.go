package service

import (
	"context"
	"fmt"

	"github.com/Yogdunana/StarByte/backend/internal/announcement/dto"
	"github.com/Yogdunana/StarByte/backend/internal/announcement/model"
	"github.com/google/uuid"
)

func (s *announcementService) MarkRead(ctx context.Context, userID, id uuid.UUID) error {
	a, err := s.load(ctx, id)
	if err != nil {
		return err
	}
	if a.Status != model.StatusPublished && a.Status != model.StatusArchived {
		return invalidState("草稿不能标记已读")
	}
	return s.rows.MarkRead(ctx, id, userID, s.clock())
}

func (s *announcementService) UnreadCount(ctx context.Context, userID uuid.UUID) (*dto.UnreadCountResponse, error) {
	n, err := s.rows.UnreadCount(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("unread count: %w", err)
	}
	return &dto.UnreadCountResponse{Count: n}, nil
}

func (s *announcementService) ReadStatus(ctx context.Context, viewer Viewer, id uuid.UUID) (*dto.ReadStatusResponse, error) {
	if !viewer.CanManage {
		return nil, noAccess("无权查看阅读回执")
	}
	a, err := s.load(ctx, id)
	if err != nil {
		return nil, err
	}
	readers, err := s.rows.ListReaders(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("list readers: %w", err)
	}
	total, err := s.rows.CountActiveUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("count users: %w", err)
	}
	readCount := int64(len(readers))
	unread := total - readCount
	if unread < 0 {
		unread = 0
	}
	out := &dto.ReadStatusResponse{
		AnnouncementID: a.ID.String(),
		ReadCount:      readCount,
		UnreadCount:    unread,
		Readers:        make([]dto.ReaderResponse, 0, len(readers)),
	}
	for _, r := range readers {
		name := r.RealName
		if name == "" {
			name = r.Username
		}
		out.Readers = append(out.Readers, dto.ReaderResponse{
			User:   dto.Person{ID: r.UserID.String(), Name: name},
			ReadAt: formatTime(r.ReadAt),
		})
	}
	return out, nil
}
