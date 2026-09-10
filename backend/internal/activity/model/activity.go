package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Activity 活动表
type Activity struct {
	ID              uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Title           string         `gorm:"type:varchar(200);not null" json:"title"`
	Description     string         `gorm:"type:text;not null;default:''" json:"description"`
	CoverImageID    *uuid.UUID     `gorm:"type:uuid" json:"cover_image_id"`
	Category        string         `gorm:"type:varchar(50);not null;default:''" json:"category"`
	Tags            []byte         `gorm:"type:jsonb;not null;default:'[]'" json:"tags"`
	StartTime       time.Time      `json:"start_time"`
	EndTime         time.Time      `json:"end_time"`
	Location        string         `gorm:"type:varchar(200);not null;default:''" json:"location"`
	MaxParticipants int            `gorm:"not null;default:0" json:"max_participants"`
	Status          int16          `gorm:"type:smallint;not null;default:0" json:"status"`
	OrganizerID     uuid.UUID      `gorm:"type:uuid;not null" json:"organizer_id"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Activity) TableName() string { return "activities" }

// ActivityWithNames 活动详情（带组织者姓名和统计）
type ActivityWithNames struct {
	Activity
	OrganizerName   string `gorm:"column:organizer_name" json:"organizer_name"`
	RegisteredCount int64  `gorm:"column:registered_count" json:"registered_count"`
	CheckedInCount  int64  `gorm:"column:checked_in_count" json:"checked_in_count"`
}

// ActivityRegistration 报名表
type ActivityRegistration struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	ActivityID    uuid.UUID  `gorm:"type:uuid;not null" json:"activity_id"`
	UserID        uuid.UUID  `gorm:"type:uuid;not null" json:"user_id"`
	Status        int16      `gorm:"type:smallint;not null;default:0" json:"status"`
	CheckinStatus int16      `gorm:"type:smallint;not null;default:0" json:"checkin_status"`
	CheckedInAt   *time.Time `json:"checked_in_at"`
	CheckinMethod *int16     `json:"checkin_method"`
	GPSLatitude   *float64   `gorm:"type:decimal(10,7)" json:"gps_latitude"`
	GPSLongitude  *float64   `gorm:"type:decimal(10,7)" json:"gps_longitude"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

func (ActivityRegistration) TableName() string { return "activity_registrations" }

// RegistrationNamed 报名记录（带用户姓名）
type RegistrationNamed struct {
	ActivityRegistration
	RealName string `gorm:"column:real_name" json:"real_name"`
	Username string `gorm:"column:username" json:"username"`
}

// ActivitySurvey 满意度调查
type ActivitySurvey struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ActivityID uuid.UUID `gorm:"type:uuid;not null" json:"activity_id"`
	UserID     uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	Rating     int16     `gorm:"type:smallint;not null" json:"rating"`
	Comment    string    `gorm:"type:text;not null;default:''" json:"comment"`
	CreatedAt  time.Time `json:"created_at"`
}

func (ActivitySurvey) TableName() string { return "activity_surveys" }

// NamedUser 关联查询用的用户信息
type NamedUser struct {
	ID       uuid.UUID
	RealName string
	Username string
}
