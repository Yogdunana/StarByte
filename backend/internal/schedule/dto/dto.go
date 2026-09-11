package dto

import "time"

type Person struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type CreateCalendarRequest struct {
	Name         string `json:"name" binding:"required,max=200"`
	Description  string `json:"description"`
	CalendarType int16  `json:"calendar_type" binding:"required,oneof=1 2 3"`
	Color        string `json:"color" binding:"omitempty,max=16"`
	DepartmentID string `json:"department_id"`
}

type UpdateCalendarRequest struct {
	Name         *string `json:"name" binding:"omitempty,max=200"`
	Description  *string `json:"description"`
	Color        *string `json:"color" binding:"omitempty,max=16"`
	DepartmentID *string `json:"department_id"`
}

type ListCalendarRequest struct {
	Page         int    `form:"page"`
	PageSize     int    `form:"page_size"`
	Keyword      string `form:"keyword"`
	CalendarType *int16 `form:"calendar_type"`
	// OwnerID 故意不接收：列表所有权一律由 JWT + DataScope 决定，避免 IDOR。
}

type AddMemberRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Role   int16  `json:"role" binding:"required,oneof=1 2"`
}

type CalendarResponse struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	CalendarType   int16     `json:"calendar_type"`
	Color          string    `json:"color"`
	Owner          Person    `json:"owner"`
	DepartmentID   string    `json:"department_id,omitempty"`
	DepartmentName string    `json:"department_name,omitempty"`
	MemberRole     int16     `json:"member_role"`
	CanEdit        bool      `json:"can_edit"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type MemberResponse struct {
	ID        string    `json:"id"`
	User      Person    `json:"user"`
	Role      int16     `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateEventRequest struct {
	CalendarID      string     `json:"calendar_id"`
	Title           string     `json:"title" binding:"required,max=200"`
	Description     string     `json:"description"`
	Location        string     `json:"location" binding:"max=200"`
	StartAt         time.Time  `json:"start_at" binding:"required"`
	EndAt           time.Time  `json:"end_at" binding:"required"`
	AllDay          bool       `json:"all_day"`
	Color           string     `json:"color" binding:"omitempty,max=16"`
	Recurrence      string     `json:"recurrence"`
	RecurrenceUntil *time.Time `json:"recurrence_until"`
	MeetingID       string     `json:"meeting_id"`
	AttendeeIDs     []string   `json:"attendee_ids"`
	RemindMinutes   []int      `json:"remind_minutes"`
}

type UpdateEventRequest struct {
	Title           *string    `json:"title" binding:"omitempty,max=200"`
	Description     *string    `json:"description"`
	Location        *string    `json:"location" binding:"omitempty,max=200"`
	StartAt         *time.Time `json:"start_at"`
	EndAt           *time.Time `json:"end_at"`
	AllDay          *bool      `json:"all_day"`
	Color           *string    `json:"color" binding:"omitempty,max=16"`
	Status          *int16     `json:"status" binding:"omitempty,oneof=0 1"`
	Recurrence      *string    `json:"recurrence"`
	RecurrenceUntil *time.Time `json:"recurrence_until"`
	MeetingID       *string    `json:"meeting_id"`
}

type ListEventRequest struct {
	Page       int       `form:"page"`
	PageSize   int       `form:"page_size"`
	Keyword    string    `form:"keyword"`
	CalendarID string    `form:"calendar_id"`
	Status     *int16    `form:"status"`
	Start      time.Time `form:"start"`
	End        time.Time `form:"end"`
	// OwnerID 故意不接收。
}

type RangeEventRequest struct {
	Start      time.Time `form:"start" binding:"required"`
	End        time.Time `form:"end" binding:"required"`
	CalendarID string    `form:"calendar_id"`
}

type RemindRequest struct {
	Minutes []int `json:"minutes" binding:"required,min=1"`
}

type RSVPRequest struct {
	Response int16 `json:"response" binding:"oneof=0 1 2 3"`
}

type EventResponse struct {
	ID              string             `json:"id"`
	CalendarID      string             `json:"calendar_id"`
	CalendarName    string             `json:"calendar_name"`
	CalendarColor   string             `json:"calendar_color"`
	Title           string             `json:"title"`
	Description     string             `json:"description"`
	Location        string             `json:"location"`
	StartAt         time.Time          `json:"start_at"`
	EndAt           time.Time          `json:"end_at"`
	AllDay          bool               `json:"all_day"`
	Color           string             `json:"color"`
	Status          int16              `json:"status"`
	Recurrence      string             `json:"recurrence"`
	RecurrenceUntil *time.Time         `json:"recurrence_until,omitempty"`
	MeetingID       string             `json:"meeting_id,omitempty"`
	Creator         Person             `json:"creator"`
	AttendeeCount   int64              `json:"attendee_count"`
	Attendees       []MemberResponse   `json:"attendees,omitempty"`
	Reminders       []ReminderResponse `json:"reminders,omitempty"`
	OccurrenceStart *time.Time         `json:"occurrence_start,omitempty"`
	CanEdit         bool               `json:"can_edit"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
}

type ReminderResponse struct {
	ID            string     `json:"id"`
	MinutesBefore int        `json:"minutes_before"`
	Method        int16      `json:"method"`
	TriggeredAt   *time.Time `json:"triggered_at,omitempty"`
}

type AttendeeResponse struct {
	ID       string `json:"id"`
	User     Person `json:"user"`
	Response int16  `json:"response"`
}
