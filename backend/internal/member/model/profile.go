package model

import (
	"time"

	"github.com/google/uuid"
)

// 人员类型
const (
	MemberTypeMember    int16 = 1
	MemberTypeOfficer   int16 = 2
	MemberTypeMinister  int16 = 3
	MemberTypePresident int16 = 4
)

// 档案状态
const (
	ProfileActive   int16 = 0
	ProfileDisabled int16 = 1
	ProfileLeft     int16 = 2
)

// MemberProfile 人员档案。
type MemberProfile struct {
	ID           uuid.UUID    `gorm:"type:uuid;primaryKey" json:"id"`
	UserID       uuid.UUID    `gorm:"type:uuid;not null;uniqueIndex" json:"user_id"`
	RealName     string       `gorm:"type:varchar(50)" json:"real_name"`
	StudentNo    string       `gorm:"type:varchar(30)" json:"student_no"`
	Gender       int16        `gorm:"type:smallint;default:0" json:"gender"`
	Grade        string       `gorm:"type:varchar(20)" json:"grade"`
	Major        string       `gorm:"type:varchar(100)" json:"major"`
	DepartmentID *uuid.UUID   `gorm:"type:uuid;index" json:"department_id"`
	PositionID   *uuid.UUID   `gorm:"type:uuid" json:"position_id"`
	MemberType   int16        `gorm:"type:smallint;not null;default:1" json:"member_type"`
	Status       int16        `gorm:"type:smallint;not null;default:0;index" json:"status"`
	JoinDate     *time.Time   `json:"join_date"`
	LeaveDate    *time.Time   `json:"leave_date"`
	Skills       JSONStrings  `gorm:"type:jsonb" json:"skills"`
	Projects     JSONProjects `gorm:"type:jsonb" json:"projects"`
	Bio          string       `gorm:"type:text" json:"bio"`
	ContactPhone string       `gorm:"type:varchar(20)" json:"contact_phone"`
	ContactEmail string       `gorm:"type:varchar(100)" json:"contact_email"`
	Points       int          `gorm:"not null;default:0" json:"points"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

func (MemberProfile) TableName() string {
	return "member_profiles"
}

// ProfileHistory 档案字段级变更。
type ProfileHistory struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	ProfileID  uuid.UUID  `gorm:"type:uuid;not null;index" json:"profile_id"`
	FieldName  string     `gorm:"type:varchar(50);not null" json:"field_name"`
	OldValue   string     `gorm:"type:text" json:"old_value"`
	NewValue   string     `gorm:"type:text" json:"new_value"`
	OperatorID *uuid.UUID `gorm:"type:uuid" json:"operator_id"`
	Reason     string     `gorm:"type:text" json:"reason"`
	CreatedAt  time.Time  `json:"created_at"`
}

func (ProfileHistory) TableName() string {
	return "member_profile_histories"
}

// ProfileWithNames 档案联表结果。
type ProfileWithNames struct {
	MemberProfile
	Username       string `gorm:"column:username"`
	DepartmentName string `gorm:"column:department_name"`
	PositionName   string `gorm:"column:position_name"`
}

// NamedItem 下拉/统计用 id+name。
type NamedItem struct {
	ID   uuid.UUID
	Name string
}

// StatBucket 分组统计。
type StatBucket struct {
	Key   string
	Label string
	Count int64
}
