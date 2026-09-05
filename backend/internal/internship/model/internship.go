package model

import (
	"time"

	"github.com/google/uuid"
)

type Internship struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UserID        uuid.UUID  `gorm:"type:uuid;not null" json:"user_id"`
	DepartmentID  *uuid.UUID `gorm:"type:uuid" json:"department_id"`
	Title         string     `gorm:"type:varchar(200);not null" json:"title"`
	Organization  string     `gorm:"type:varchar(200);not null" json:"organization"`
	Description   string     `gorm:"type:text;not null;default:''" json:"description"`
	StartDate     time.Time  `gorm:"type:date;not null" json:"start_date"`
	EndDate       *time.Time `gorm:"type:date" json:"end_date"`
	Status        int16      `gorm:"type:smallint;not null;default:0" json:"status"`
	Type          int16      `gorm:"type:smallint;not null;default:0" json:"type"`
	MentorID      *uuid.UUID `gorm:"type:uuid" json:"mentor_id"`
	SupervisorID  *uuid.UUID `gorm:"type:uuid" json:"supervisor_id"`
	Skills        string     `gorm:"type:text;not null;default:'[]'" json:"skills"`
	Achievements  string     `gorm:"type:text;not null;default:''" json:"achievements"`
	Report        string     `gorm:"type:text;not null;default:''" json:"report"`
	MentorComment string     `gorm:"type:text;not null;default:''" json:"mentor_comment"`
	CreatedBy     uuid.UUID  `gorm:"type:uuid" json:"created_by"`
	UpdatedBy     *uuid.UUID `gorm:"type:uuid" json:"updated_by"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

func (Internship) TableName() string { return "internships" }

type InternshipWithNames struct {
	Internship
	UserName       string `gorm:"column:user_name"`
	UserAvatar     string `gorm:"column:user_avatar"`
	DepartmentName string `gorm:"column:department_name"`
	MentorName     string `gorm:"column:mentor_name"`
	SupervisorName string `gorm:"column:supervisor_name"`
}

type NamedUser struct {
	ID           uuid.UUID
	RealName     string
	Username     string
	Avatar       string
	DepartmentID *uuid.UUID
}

type SystemConfig struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey"`
	ConfigKey   string     `gorm:"column:config_key;type:varchar(100);uniqueIndex"`
	ConfigValue string     `gorm:"column:config_value;type:text"`
	ConfigType  string     `gorm:"column:config_type;type:varchar(20)"`
	Description string     `gorm:"type:varchar(255)"`
	Category    string     `gorm:"type:varchar(50)"`
	IsPublic    bool       `gorm:"column:is_public"`
	UpdatedBy   *uuid.UUID `gorm:"type:uuid"`
	UpdatedAt   time.Time
	CreatedAt   time.Time
}

func (SystemConfig) TableName() string { return "configs" }

type InternshipConfig struct {
	AllowStudentEdit  bool `json:"allow_student_edit"`
	AllowMinisterEdit bool `json:"allow_minister_edit"`
	RankingVisible    bool `json:"ranking_visible"`
}

type DurationRow struct {
	Key          string `gorm:"column:group_key"`
	Name         string `gorm:"column:group_name"`
	DurationDays int    `gorm:"column:duration_days"`
	Count        int64  `gorm:"column:item_count"`
}

type RankingRow struct {
	UserID           uuid.UUID  `gorm:"column:user_id"`
	UserName         string     `gorm:"column:user_name"`
	UserAvatar       string     `gorm:"column:user_avatar"`
	DepartmentID     *uuid.UUID `gorm:"column:department_id"`
	DepartmentName   string     `gorm:"column:department_name"`
	TotalDuration    int        `gorm:"column:total_duration"`
	InternshipCount  int64      `gorm:"column:internship_count"`
	LatestInternship string     `gorm:"column:latest_internship"`
}

type DepartmentStatRow struct {
	DepartmentID   *uuid.UUID `gorm:"column:department_id"`
	DepartmentName string     `gorm:"column:department_name"`
	DurationDays   int        `gorm:"column:duration_days"`
	Count          int64      `gorm:"column:item_count"`
	Ongoing        int64      `gorm:"column:ongoing_count"`
}
