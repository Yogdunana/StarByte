package model

import (
	"time"

	"github.com/google/uuid"
)

// 申请类型
const (
	ApplicantMember  int16 = 1
	ApplicantOfficer int16 = 2
)

// 申请状态（Issue #6）
const (
	AppPending      int16 = 0
	AppReviewing    int16 = 1
	AppInterviewing int16 = 2
	AppApproved     int16 = 3
	AppRejected     int16 = 4
	AppSupplement   int16 = 5
)

// MemberApplication 入会申请。type 列沿用 000010，API 对外叫 applicant_type。
type MemberApplication struct {
	ID             uuid.UUID   `gorm:"type:uuid;primaryKey" json:"id"`
	UserID         uuid.UUID   `gorm:"type:uuid;not null;index" json:"user_id"`
	Type           int16       `gorm:"column:type;type:smallint;not null;default:1" json:"applicant_type"`
	RealName       string      `gorm:"type:varchar(50);not null" json:"real_name"`
	StudentNo      string      `gorm:"type:varchar(30);not null" json:"student_no"`
	DepartmentID   *uuid.UUID  `gorm:"type:uuid" json:"department_id"`
	Reason         string      `gorm:"type:text" json:"reason"`
	Skills         JSONStrings `gorm:"type:jsonb" json:"skills"`
	Experience     string      `gorm:"type:text" json:"experience"`
	ContactPhone   string      `gorm:"type:varchar(20)" json:"contact_phone"`
	ContactEmail   string      `gorm:"type:varchar(100)" json:"contact_email"`
	Status         int16       `gorm:"type:smallint;not null;default:0;index" json:"status"`
	CurrentStage   string      `gorm:"type:varchar(50)" json:"current_stage"`
	FlowInstanceID *uuid.UUID  `gorm:"type:uuid" json:"flow_instance_id"`
	ReviewerID     *uuid.UUID  `gorm:"type:uuid" json:"reviewer_id"`
	ReviewComment  string      `gorm:"type:text" json:"review_comment"`
	RequiredFields JSONStrings `gorm:"type:jsonb" json:"required_fields"`
	SubmittedAt    time.Time   `json:"submitted_at"`
	ReviewedAt     *time.Time  `json:"reviewed_at"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
}

func (MemberApplication) TableName() string {
	return "member_applications"
}

// ApplicationHistory 申请状态变更历史。
type ApplicationHistory struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	ApplicationID uuid.UUID  `gorm:"type:uuid;not null;index" json:"application_id"`
	FromStatus    int16      `gorm:"type:smallint;not null" json:"from_status"`
	ToStatus      int16      `gorm:"type:smallint;not null" json:"to_status"`
	OperatorID    *uuid.UUID `gorm:"type:uuid" json:"operator_id"`
	Comment       string     `gorm:"type:text" json:"comment"`
	Extra         []byte     `gorm:"type:jsonb" json:"extra"`
	CreatedAt     time.Time  `json:"created_at"`
}

func (ApplicationHistory) TableName() string {
	return "member_application_histories"
}

// ApplicationWithNames 列表/详情联表结果。
type ApplicationWithNames struct {
	MemberApplication
	Username       string `gorm:"column:username"`
	DepartmentName string `gorm:"column:department_name"`
	ReviewerName   string `gorm:"column:reviewer_name"`
}
