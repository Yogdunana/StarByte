package dto

import "time"

type Person struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type CreateMeetingRequest struct {
	Title       string    `json:"title" binding:"required,max=200"`
	Description string    `json:"description"`
	StartTime   time.Time `json:"start_time" binding:"required"`
	EndTime     time.Time `json:"end_time" binding:"required"`
	Location    string    `json:"location" binding:"max=200"`
	OnlineLink  string    `json:"online_link" binding:"max=500"`
	MeetingType int16     `json:"meeting_type" binding:"required,oneof=1 2 3"`
	UserIDs     []string  `json:"user_ids"`
}

type UpdateMeetingRequest struct {
	Title       *string    `json:"title"`
	Description *string    `json:"description"`
	StartTime   *time.Time `json:"start_time"`
	EndTime     *time.Time `json:"end_time"`
	Location    *string    `json:"location"`
	OnlineLink  *string    `json:"online_link"`
	MeetingType *int16     `json:"meeting_type"`
}

type ListMeetingRequest struct {
	Page      int    `form:"page"`
	PageSize  int    `form:"page_size"`
	Status    *int16 `form:"status"`
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
	Keyword   string `form:"keyword"`
}

type CancelMeetingRequest struct {
	Reason string `json:"reason" binding:"max=500"`
}

type MinutesRequest struct {
	Minutes string `json:"minutes" binding:"required"`
}

type MeetingResponse struct {
	ID             string    `json:"id"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time"`
	Location       string    `json:"location"`
	OnlineLink     string    `json:"online_link,omitempty"`
	Organizer      Person    `json:"organizer"`
	Status         int16     `json:"status"`
	MeetingType    int16     `json:"meeting_type"`
	Minutes        string    `json:"minutes,omitempty"`
	CancelReason   string    `json:"cancel_reason,omitempty"`
	AttendeeCount  int64     `json:"attendee_count"`
	CheckedInCount int64     `json:"checked_in_count"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type QRCodeResponse struct {
	MeetingID   string `json:"meeting_id"`
	Token       string `json:"token"`
	CheckinPath string `json:"checkin_path"`
	PNGBase64   string `json:"png_base64"`
}

type CreateAgendaRequest struct {
	Title     string `json:"title" binding:"required,max=200"`
	Content   string `json:"content"`
	Duration  int    `json:"duration"`
	Presenter string `json:"presenter" binding:"max=100"`
	SortOrder int    `json:"sort_order"`
}

type UpdateAgendaRequest struct {
	Title     *string `json:"title"`
	Content   *string `json:"content"`
	Duration  *int    `json:"duration"`
	Presenter *string `json:"presenter"`
	SortOrder *int    `json:"sort_order"`
}

type SortAgendasRequest struct {
	AgendaIDs []string `json:"agenda_ids" binding:"required,min=1"`
}

type AgendaResponse struct {
	ID        string `json:"id"`
	MeetingID string `json:"meeting_id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Duration  int    `json:"duration"`
	SortOrder int    `json:"sort_order"`
	Presenter string `json:"presenter"`
}

type AddAttendeesRequest struct {
	UserIDs []string `json:"user_ids" binding:"required,min=1"`
}

type AttendeeResponse struct {
	ID           string  `json:"id"`
	UserID       string  `json:"user_id"`
	Name         string  `json:"name"`
	PositionCode string  `json:"position_code,omitempty"`
	Attended     bool    `json:"attended"`
	CheckedInAt  *string `json:"checked_in_at,omitempty"`
}
