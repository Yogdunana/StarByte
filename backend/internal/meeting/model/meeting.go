package model

import (
	"time"

	"github.com/google/uuid"
)

type Meeting struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Title        string    `gorm:"type:varchar(200);not null" json:"title"`
	Description  string    `gorm:"type:text" json:"description"`
	Status       int16     `gorm:"type:smallint;not null;default:0" json:"status"`
	MeetingType  int16     `gorm:"type:smallint;not null;default:1" json:"meeting_type"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	Location     string    `gorm:"type:varchar(200)" json:"location"`
	OnlineLink   string    `gorm:"type:varchar(500);not null;default:''" json:"online_link"`
	OrganizerID  uuid.UUID `gorm:"type:uuid;not null" json:"organizer_id"`
	Minutes      string    `gorm:"type:text;not null;default:''" json:"minutes"`
	QRToken      string    `gorm:"column:qr_token;type:varchar(64);not null;default:''" json:"-"`
	CancelReason string    `gorm:"type:varchar(500);not null;default:''" json:"cancel_reason"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (Meeting) TableName() string { return "meetings" }

type MeetingWithNames struct {
	Meeting
	OrganizerName  string `gorm:"column:organizer_name"`
	AttendeeCount  int64  `gorm:"column:attendee_count"`
	CheckedInCount int64  `gorm:"column:checked_in_count"`
}

type NamedUser struct {
	ID           uuid.UUID
	RealName     string
	Username     string
	PositionID   *uuid.UUID
	PositionCode string
}

type Agenda struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	MeetingID   uuid.UUID  `gorm:"type:uuid;not null" json:"meeting_id"`
	Title       string     `gorm:"type:varchar(200);not null" json:"title"`
	Description string     `gorm:"type:text" json:"description"`
	SortOrder   int        `gorm:"not null;default:0" json:"sort_order"`
	Duration    *int       `json:"duration"`
	SpeakerID   *uuid.UUID `gorm:"type:uuid" json:"speaker_id"`
	Presenter   string     `gorm:"type:varchar(100);not null;default:''" json:"presenter"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (Agenda) TableName() string { return "meeting_agendas" }

type Attendee struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	MeetingID   uuid.UUID  `gorm:"type:uuid;not null" json:"meeting_id"`
	UserID      uuid.UUID  `gorm:"type:uuid;not null" json:"user_id"`
	Attended    bool       `gorm:"default:false" json:"attended"`
	CheckedInAt *time.Time `json:"checked_in_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

func (Attendee) TableName() string { return "meeting_attendees" }

type AttendeeNamed struct {
	Attendee
	RealName     string `gorm:"column:real_name"`
	Username     string `gorm:"column:username"`
	PositionCode string `gorm:"column:position_code"`
}
