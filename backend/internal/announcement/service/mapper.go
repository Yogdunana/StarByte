package service

import (
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/announcement/dto"
	"github.com/Yogdunana/StarByte/backend/internal/announcement/model"
)

func toResponse(a *model.AnnouncementNamed) *dto.AnnouncementResponse {
	resp := &dto.AnnouncementResponse{
		ID:          a.ID.String(),
		Title:       a.Title,
		Content:     a.Content,
		ContentType: a.ContentType,
		Category:    a.Category,
		Pinned:      a.Pinned,
		Required:    a.Required,
		Status:      a.Status,
		Author:      dto.Person{ID: a.AuthorID.String(), Name: a.AuthorName},
		IsRead:      a.IsRead,
		CreatedAt:   formatTime(a.CreatedAt),
		UpdatedAt:   formatTime(a.UpdatedAt),
	}
	if a.ScheduledAt != nil {
		resp.ScheduledAt = formatTime(*a.ScheduledAt)
	}
	if a.PublishedAt != nil {
		resp.PublishedAt = formatTime(*a.PublishedAt)
	}
	return resp
}

func formatTime(t time.Time) string {
	return t.Format(time.RFC3339)
}
