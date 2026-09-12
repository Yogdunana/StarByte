package service

import (
	"context"
	"fmt"

	"github.com/Yogdunana/StarByte/backend/internal/announcement/dto"
	"github.com/Yogdunana/StarByte/backend/internal/announcement/model"
	"github.com/google/uuid"
)

func (s *announcementService) MarkRead(ctx context.Context, viewer Viewer, id uuid.UUID, duration int) error {
	a, err := s.load(ctx, id)
	if err != nil {
		return err
	}
	if a.Status != model.StatusPublished && a.Status != model.StatusArchived {
		return invalidState("草稿不能标记已读")
	}
	ok, err := s.canViewPublished(ctx, viewer, a)
	if err != nil {
		return err
	}
	if !ok {
		return noAccess("无权阅读该公告")
	}
	return s.rows.MarkRead(ctx, id, viewer.UserID, s.clock(), duration)
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
	targets, err := s.recipients(ctx, a)
	if err != nil {
		return nil, fmt.Errorf("list recipients: %w", err)
	}
	targetSet := map[uuid.UUID]struct{}{}
	for _, uid := range targets {
		targetSet[uid] = struct{}{}
	}
	readSet := map[uuid.UUID]struct{}{}
	var durationSum int
	var durationN int
	outReaders := make([]dto.ReaderResponse, 0, len(readers))
	for _, r := range readers {
		if _, ok := targetSet[r.UserID]; !ok {
			continue
		}
		readSet[r.UserID] = struct{}{}
		if r.DurationSeconds > 0 {
			durationSum += r.DurationSeconds
			durationN++
		}
		name := r.RealName
		if name == "" {
			name = r.Username
		}
		outReaders = append(outReaders, dto.ReaderResponse{
			User:            dto.Person{ID: r.UserID.String(), Name: name},
			ReadAt:          formatTime(r.ReadAt),
			DurationSeconds: r.DurationSeconds,
		})
	}
	unreadIDs := make([]uuid.UUID, 0)
	for _, uid := range targets {
		if _, ok := readSet[uid]; !ok {
			unreadIDs = append(unreadIDs, uid)
		}
	}
	unreadUsers, err := s.rows.ListNamedUsers(ctx, unreadIDs)
	if err != nil {
		return nil, fmt.Errorf("list unread users: %w", err)
	}
	unreadPeople := make([]dto.Person, 0, len(unreadUsers))
	for _, u := range unreadUsers {
		name := u.RealName
		if name == "" {
			name = u.Username
		}
		unreadPeople = append(unreadPeople, dto.Person{ID: u.ID.String(), Name: name})
	}
	readCount := int64(len(outReaders))
	total := int64(len(targets))
	unread := int64(len(unreadPeople))
	if unread < 0 {
		unread = 0
	}
	var rate float64
	if total > 0 {
		rate = float64(readCount) / float64(total)
	}
	var avg float64
	if durationN > 0 {
		avg = float64(durationSum) / float64(durationN)
	}
	return &dto.ReadStatusResponse{
		AnnouncementID:     a.ID.String(),
		ReadCount:          readCount,
		UnreadCount:        unread,
		ReadRate:           rate,
		AvgDurationSeconds: avg,
		Readers:            outReaders,
		UnreadUsers:        unreadPeople,
	}, nil
}
