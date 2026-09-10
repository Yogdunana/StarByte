package service

import (
	"encoding/json"

	"github.com/Yogdunana/StarByte/backend/internal/activity/dto"
	"github.com/Yogdunana/StarByte/backend/internal/activity/model"
)

func toActivityResponse(a *model.ActivityWithNames) *dto.ActivityResponse {
	coverID := ""
	if a.CoverImageID != nil {
		coverID = a.CoverImageID.String()
	}
	var tags []string
	_ = json.Unmarshal(a.Tags, &tags)
	return &dto.ActivityResponse{
		ID:              a.ID.String(),
		Title:           a.Title,
		Description:     a.Description,
		CoverImageID:    coverID,
		Category:        a.Category,
		Tags:            tags,
		StartTime:       formatTime(a.StartTime),
		EndTime:         formatTime(a.EndTime),
		Location:        a.Location,
		MaxParticipants: a.MaxParticipants,
		Status:          a.Status,
		Organizer: dto.Person{
			ID:   a.OrganizerID.String(),
			Name: a.OrganizerName,
		},
		RegisteredCount: a.RegisteredCount,
		CheckedInCount:  a.CheckedInCount,
		CreatedAt:       formatTime(a.CreatedAt),
		UpdatedAt:       formatTime(a.UpdatedAt),
	}
}

func toRegistrationResponse(r *model.RegistrationNamed) dto.RegistrationResponse {
	resp := dto.RegistrationResponse{
		ID:            r.ID.String(),
		ActivityID:    r.ActivityID.String(),
		User:          dto.Person{ID: r.UserID.String(), Name: r.RealName},
		Status:        r.Status,
		CheckinStatus: r.CheckinStatus,
		CheckinMethod: r.CheckinMethod,
		CreatedAt:     formatTime(r.CreatedAt),
	}
	if r.CheckedInAt != nil {
		resp.CheckedInAt = formatTime(*r.CheckedInAt)
	}
	return resp
}
