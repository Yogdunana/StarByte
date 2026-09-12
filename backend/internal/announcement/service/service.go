package service

import (
	"context"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/announcement/dto"
	"github.com/Yogdunana/StarByte/backend/internal/announcement/model"
	"github.com/Yogdunana/StarByte/backend/internal/announcement/repo"
	"github.com/google/uuid"
)

// Notifier 复用站内信 / WebSocket 渠道（#4），不重建通知栈。
type Notifier interface {
	Send(ctx context.Context, userIDs []uuid.UUID, template string, vars map[string]interface{}) error
}

type Viewer struct {
	UserID     uuid.UUID
	Staff      bool
	CanPublish bool
	CanManage  bool
}

type Service interface {
	Create(ctx context.Context, viewer Viewer, req *dto.CreateAnnouncementRequest) (*dto.AnnouncementResponse, error)
	Update(ctx context.Context, viewer Viewer, id uuid.UUID, req *dto.UpdateAnnouncementRequest) (*dto.AnnouncementResponse, error)
	Delete(ctx context.Context, viewer Viewer, id uuid.UUID) error
	Get(ctx context.Context, viewer Viewer, id uuid.UUID) (*dto.AnnouncementResponse, error)
	List(ctx context.Context, viewer Viewer, req *dto.ListAnnouncementRequest) ([]*dto.AnnouncementResponse, int64, error)

	Publish(ctx context.Context, viewer Viewer, id uuid.UUID) (*dto.AnnouncementResponse, error)
	Pin(ctx context.Context, viewer Viewer, id uuid.UUID, pinned *bool, sortOrder *int) (*dto.AnnouncementResponse, error)
	Archive(ctx context.Context, viewer Viewer, id uuid.UUID) (*dto.AnnouncementResponse, error)

	MarkRead(ctx context.Context, userID, id uuid.UUID, duration int) error
	UnreadCount(ctx context.Context, userID uuid.UUID) (*dto.UnreadCountResponse, error)
	ReadStatus(ctx context.Context, viewer Viewer, id uuid.UUID) (*dto.ReadStatusResponse, error)

	DispatchDuePublishes(ctx context.Context, payload string, logf func(string)) error
	DispatchExpired(ctx context.Context, payload string, logf func(string)) error
}

type announcementService struct {
	rows   repo.Repository
	notify Notifier
	now    func() time.Time
}

func New(rows repo.Repository, notify Notifier) Service {
	return &announcementService{rows: rows, notify: notify, now: time.Now}
}

func (s *announcementService) clock() time.Time {
	if s.now != nil {
		return s.now()
	}
	return time.Now()
}

func (s *announcementService) load(ctx context.Context, id uuid.UUID) (*model.Announcement, error) {
	a, err := s.rows.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if a == nil {
		return nil, notFound()
	}
	return a, nil
}
