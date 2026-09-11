package model

import (
	"time"

	"github.com/google/uuid"
)

type Calendar struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	Name         string     `gorm:"type:varchar(200);not null" json:"name"`
	Description  string     `gorm:"type:text;not null;default:''" json:"description"`
	CalendarType int16      `gorm:"type:smallint;not null;default:1" json:"calendar_type"`
	Source       string     `gorm:"type:varchar(32);not null;default:personal" json:"source"`
	SourceKey    string     `gorm:"type:varchar(200);not null;default:''" json:"source_key"`
	Color        string     `gorm:"type:varchar(16);not null;default:'#2563eb'" json:"color"`
	OwnerID      uuid.UUID  `gorm:"type:uuid;not null" json:"owner_id"`
	DepartmentID *uuid.UUID `gorm:"type:uuid" json:"department_id"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func (Calendar) TableName() string { return "calendars" }

type CalendarNamed struct {
	Calendar
	OwnerName      string `gorm:"column:owner_name"`
	DepartmentName string `gorm:"column:department_name"`
	MemberRole     int16  `gorm:"column:member_role"`
}

type CalendarMember struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	CalendarID uuid.UUID `gorm:"type:uuid;not null" json:"calendar_id"`
	UserID     uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	Role       int16     `gorm:"type:smallint;not null;default:1" json:"role"`
	CreatedAt  time.Time `json:"created_at"`
}

func (CalendarMember) TableName() string { return "calendar_members" }

type CalendarMemberNamed struct {
	CalendarMember
	RealName string `gorm:"column:real_name"`
	Username string `gorm:"column:username"`
}

type Event struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	CalendarID      uuid.UUID  `gorm:"type:uuid;not null" json:"calendar_id"`
	Title           string     `gorm:"type:varchar(200);not null" json:"title"`
	Description     string     `gorm:"type:text;not null;default:''" json:"description"`
	Location        string     `gorm:"type:varchar(200);not null;default:''" json:"location"`
	StartAt         time.Time  `json:"start_at"`
	EndAt           time.Time  `json:"end_at"`
	AllDay          bool       `json:"all_day"`
	Color           string     `gorm:"type:varchar(16);not null;default:''" json:"color"`
	Status          int16      `gorm:"type:smallint;not null;default:0" json:"status"`
	Recurrence      string     `gorm:"type:varchar(16);not null;default:'none'" json:"recurrence"`
	RecurrenceUntil *time.Time `json:"recurrence_until"`
	MeetingID       *uuid.UUID `gorm:"type:uuid" json:"meeting_id"`
	Origin          string     `gorm:"type:varchar(32);not null;default:manual" json:"origin"`
	ExternalUID     string     `gorm:"type:varchar(200);not null;default:''" json:"external_uid"`
	CreatedBy       uuid.UUID  `gorm:"type:uuid;not null" json:"created_by"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func (Event) TableName() string { return "schedule_events" }

type EventNamed struct {
	Event
	CalendarName    string     `gorm:"column:calendar_name"`
	CalendarColor   string     `gorm:"column:calendar_color"`
	CalendarType    int16      `gorm:"column:calendar_type"`
	CalendarSource  string     `gorm:"column:calendar_source"`
	OwnerID         uuid.UUID  `gorm:"column:owner_id"`
	DepartmentID    *uuid.UUID `gorm:"column:department_id"`
	CreatorName     string     `gorm:"column:creator_name"`
	AttendeeCount   int64      `gorm:"column:attendee_count"`
	MemberRole      int16      `gorm:"column:member_role"`
	OccurrenceStart *time.Time `gorm:"-" json:"-"`
}

type Attendee struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	EventID        uuid.UUID `gorm:"type:uuid;not null" json:"event_id"`
	UserID         uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	ResponseStatus int16     `gorm:"type:smallint;not null;default:0" json:"response_status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (Attendee) TableName() string { return "schedule_event_attendees" }

type AttendeeNamed struct {
	Attendee
	RealName string `gorm:"column:real_name"`
	Username string `gorm:"column:username"`
}

type Reminder struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	EventID       uuid.UUID  `gorm:"type:uuid;not null" json:"event_id"`
	MinutesBefore int        `gorm:"not null" json:"minutes_before"`
	Method        int16      `gorm:"type:smallint;not null;default:1" json:"method"`
	TriggeredAt   *time.Time `json:"triggered_at"`
	CreatedAt     time.Time  `json:"created_at"`
}

func (Reminder) TableName() string { return "schedule_reminders" }

type DueReminder struct {
	Reminder
	Title     string
	StartAt   time.Time
	OwnerID   uuid.UUID
	CreatedBy uuid.UUID
}

type NamedUser struct {
	ID           uuid.UUID
	RealName     string
	Username     string
	DepartmentID *uuid.UUID
}

type GoogleAccount struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UserID       uuid.UUID  `gorm:"type:uuid;not null" json:"user_id"`
	CalendarID   *uuid.UUID `gorm:"type:uuid" json:"calendar_id"`
	AccessToken  string     `gorm:"type:text;not null;default:''" json:"-"`
	RefreshToken string     `gorm:"type:text;not null;default:''" json:"-"`
	TokenExpiry  *time.Time `json:"-"`
	GoogleEmail  string     `gorm:"type:varchar(200);not null;default:''" json:"google_email"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func (GoogleAccount) TableName() string { return "schedule_google_accounts" }
