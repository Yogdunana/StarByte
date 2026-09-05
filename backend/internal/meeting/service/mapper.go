package service

import (
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/meeting/dto"
	"github.com/Yogdunana/StarByte/backend/internal/meeting/model"
)

func mapMeeting(row *model.MeetingWithNames) *dto.MeetingResponse {
	return &dto.MeetingResponse{
		ID:             row.ID.String(),
		Title:          row.Title,
		Description:    row.Description,
		StartTime:      row.StartTime,
		EndTime:        row.EndTime,
		Location:       row.Location,
		OnlineLink:     row.OnlineLink,
		Organizer:      dto.Person{ID: row.OrganizerID.String(), Name: row.OrganizerName},
		Status:         row.Status,
		MeetingType:    row.MeetingType,
		Minutes:        row.Minutes,
		CancelReason:   row.CancelReason,
		AttendeeCount:  row.AttendeeCount,
		CheckedInCount: row.CheckedInCount,
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
	}
}

func mapAgenda(a model.Agenda) dto.AgendaResponse {
	dur := 0
	if a.Duration != nil {
		dur = *a.Duration
	}
	return dto.AgendaResponse{
		ID:        a.ID.String(),
		MeetingID: a.MeetingID.String(),
		Title:     a.Title,
		Content:   a.Description,
		Duration:  dur,
		SortOrder: a.SortOrder,
		Presenter: a.Presenter,
	}
}

func mapAttendee(a model.AttendeeNamed) dto.AttendeeResponse {
	out := dto.AttendeeResponse{
		ID:           a.ID.String(),
		UserID:       a.UserID.String(),
		Name:         a.RealName,
		PositionCode: a.PositionCode,
		Attended:     a.Attended,
	}
	if a.CheckedInAt != nil {
		s := a.CheckedInAt.Format(time.RFC3339)
		out.CheckedInAt = &s
	}
	return out
}

func mapVote(v *model.Vote, options []model.VoteOption, hasVoted bool) *dto.VoteResponse {
	opts := make([]dto.VoteOptionResponse, 0, len(options))
	for _, o := range options {
		opts = append(opts, dto.VoteOptionResponse{Key: o.OptionKey, Label: o.OptionText})
	}
	return &dto.VoteResponse{
		ID:          v.ID.String(),
		MeetingID:   v.MeetingID.String(),
		Title:       v.Title,
		Description: v.Description,
		VoteType:    v.VoteType,
		IsAnonymous: v.IsAnonymous,
		Options:     opts,
		Status:      v.Status,
		StartTime:   v.StartTime,
		EndTime:     v.EndTime,
		HasVoted:    hasVoted,
		CreatedAt:   v.CreatedAt,
	}
}

func displayName(u *model.NamedUser) string {
	if u == nil {
		return ""
	}
	if u.RealName != "" {
		return u.RealName
	}
	return u.Username
}
