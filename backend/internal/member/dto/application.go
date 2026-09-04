package dto

import "time"

// SubmitApplicationRequest 提交入会申请。
type SubmitApplicationRequest struct {
	ApplicantType int      `json:"applicant_type" binding:"required,oneof=1 2"`
	RealName      string   `json:"real_name" binding:"required,max=50"`
	StudentNo     string   `json:"student_no" binding:"required,max=30"`
	DepartmentID  string   `json:"department_id" binding:"omitempty,uuid"`
	Reason        string   `json:"reason" binding:"required,max=2000"`
	Skills        []string `json:"skills"`
	Experience    string   `json:"experience" binding:"max=4000"`
	ContactPhone  string   `json:"contact_phone" binding:"required,max=20"`
	ContactEmail  string   `json:"contact_email" binding:"required,email,max=100"`
}

// ResubmitApplicationRequest 补充材料后重新提交。
type ResubmitApplicationRequest struct {
	RealName     string   `json:"real_name" binding:"omitempty,max=50"`
	StudentNo    string   `json:"student_no" binding:"omitempty,max=30"`
	DepartmentID string   `json:"department_id" binding:"omitempty,uuid"`
	Reason       string   `json:"reason" binding:"omitempty,max=2000"`
	Skills       []string `json:"skills"`
	Experience   string   `json:"experience" binding:"omitempty,max=4000"`
	ContactPhone string   `json:"contact_phone" binding:"omitempty,max=20"`
	ContactEmail string   `json:"contact_email" binding:"omitempty,email,max=100"`
}

// ReviewCommentRequest 通过/拒绝。
type ReviewCommentRequest struct {
	Comment string `json:"comment" binding:"max=1000"`
}

// SupplementRequest 要求补充材料。
type SupplementRequest struct {
	Comment        string   `json:"comment" binding:"required,max=1000"`
	RequiredFields []string `json:"required_fields"`
}

// ListApplicationRequest 申请列表查询。
type ListApplicationRequest struct {
	Page          int    `form:"page"`
	PageSize      int    `form:"page_size"`
	Status        *int16 `form:"status"`
	ApplicantType *int16 `form:"applicant_type"`
	DepartmentID  string `form:"department_id"`
	Keyword       string `form:"keyword"`
}

// ReviewerInfo 审核人摘要。
type ReviewerInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ApplicationResponse 申请详情/列表项。
type ApplicationResponse struct {
	ID             string        `json:"id"`
	UserID         string        `json:"user_id"`
	Username       string        `json:"username,omitempty"`
	ApplicantType  int16         `json:"applicant_type"`
	RealName       string        `json:"real_name"`
	StudentNo      string        `json:"student_no"`
	DepartmentID   string        `json:"department_id,omitempty"`
	DepartmentName string        `json:"department_name,omitempty"`
	Reason         string        `json:"reason"`
	Skills         []string      `json:"skills"`
	Experience     string        `json:"experience"`
	ContactPhone   string        `json:"contact_phone"`
	ContactEmail   string        `json:"contact_email"`
	Status         int16         `json:"status"`
	CurrentStage   string        `json:"current_stage,omitempty"`
	FlowInstanceID string        `json:"flow_instance_id,omitempty"`
	Reviewer       *ReviewerInfo `json:"reviewer,omitempty"`
	ReviewComment  string        `json:"review_comment,omitempty"`
	RequiredFields []string      `json:"required_fields,omitempty"`
	ReviewedAt     *time.Time    `json:"reviewed_at,omitempty"`
	SubmittedAt    time.Time     `json:"submitted_at"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
}

// ApplicationHistoryResponse 申请历史。
type ApplicationHistoryResponse struct {
	ID         string    `json:"id"`
	FromStatus int16     `json:"from_status"`
	ToStatus   int16     `json:"to_status"`
	OperatorID string    `json:"operator_id,omitempty"`
	Comment    string    `json:"comment"`
	CreatedAt  time.Time `json:"created_at"`
}

// DepartmentOption 意向部门下拉。
type DepartmentOption struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
