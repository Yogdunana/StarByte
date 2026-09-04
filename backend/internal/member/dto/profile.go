package dto

import "time"

// NamedRef 部门/职位摘要。
type NamedRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ProjectItem 项目经历。
type ProjectItem struct {
	Name   string `json:"name"`
	Role   string `json:"role"`
	Period string `json:"period"`
}

// ListProfileRequest 档案列表查询。
type ListProfileRequest struct {
	Page         int    `form:"page"`
	PageSize     int    `form:"page_size"`
	DepartmentID string `form:"department_id"`
	MemberType   *int16 `form:"member_type"`
	Status       *int16 `form:"status"`
	Keyword      string `form:"keyword"`
	IDs          string `form:"ids"`
}

// UpdateProfileRequest 更新档案。
type UpdateProfileRequest struct {
	RealName     string        `json:"real_name" binding:"omitempty,max=50"`
	Gender       *int16        `json:"gender" binding:"omitempty,oneof=0 1 2"`
	Grade        string        `json:"grade" binding:"omitempty,max=20"`
	Major        string        `json:"major" binding:"omitempty,max=100"`
	Skills       []string      `json:"skills"`
	Projects     []ProjectItem `json:"projects"`
	Bio          string        `json:"bio" binding:"omitempty,max=2000"`
	ContactPhone string        `json:"contact_phone" binding:"omitempty,max=20"`
	ContactEmail string        `json:"contact_email" binding:"omitempty,email,max=100"`
}

// UpdateProfileStatusRequest 变更档案状态。
type UpdateProfileStatusRequest struct {
	Status int16  `json:"status" binding:"required,oneof=0 1 2"`
	Reason string `json:"reason" binding:"required,max=500"`
}

// ProfileResponse 档案详情。
type ProfileResponse struct {
	ID           string        `json:"id"`
	UserID       string        `json:"user_id"`
	Username     string        `json:"username,omitempty"`
	RealName     string        `json:"real_name"`
	StudentNo    string        `json:"student_no"`
	Gender       int16         `json:"gender"`
	Grade        string        `json:"grade"`
	Major        string        `json:"major"`
	Department   *NamedRef     `json:"department,omitempty"`
	Position     *NamedRef     `json:"position,omitempty"`
	MemberType   int16         `json:"member_type"`
	Status       int16         `json:"status"`
	JoinDate     *time.Time    `json:"join_date,omitempty"`
	LeaveDate    *time.Time    `json:"leave_date,omitempty"`
	Skills       []string      `json:"skills"`
	Projects     []ProjectItem `json:"projects"`
	Bio          string        `json:"bio"`
	ContactPhone string        `json:"contact_phone"`
	ContactEmail string        `json:"contact_email"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

// ProfileHistoryResponse 字段级变更历史。
type ProfileHistoryResponse struct {
	ID         string    `json:"id"`
	FieldName  string    `json:"field_name"`
	OldValue   string    `json:"old_value"`
	NewValue   string    `json:"new_value"`
	OperatorID string    `json:"operator_id,omitempty"`
	Reason     string    `json:"reason"`
	CreatedAt  time.Time `json:"created_at"`
}
