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
	AdmissionVersion         int16         `json:"admission_version"`
	AdmissionRevision        int           `json:"admission_revision"`
	AdmissionStage           string        `json:"admission_stage"`
	HistoricalReviewRequired bool          `json:"historical_review_required"`
	ProbationUntil           *time.Time    `json:"probation_until,omitempty"`
	ID                       string        `json:"id"`
	UserID                   string        `json:"user_id"`
	Username                 string        `json:"username,omitempty"`
	ApplicantType            int16         `json:"applicant_type"`
	RealName                 string        `json:"real_name"`
	StudentNo                string        `json:"student_no"`
	DepartmentID             string        `json:"department_id,omitempty"`
	DepartmentName           string        `json:"department_name,omitempty"`
	Reason                   string        `json:"reason"`
	Skills                   []string      `json:"skills"`
	Experience               string        `json:"experience"`
	ContactPhone             string        `json:"contact_phone"`
	ContactEmail             string        `json:"contact_email"`
	Status                   int16         `json:"status"`
	CurrentStage             string        `json:"current_stage,omitempty"`
	FlowInstanceID           string        `json:"flow_instance_id,omitempty"`
	WorkflowKey              string        `json:"workflow_key,omitempty"`
	Reviewer                 *ReviewerInfo `json:"reviewer,omitempty"`
	ReviewComment            string        `json:"review_comment,omitempty"`
	RequiredFields           []string      `json:"required_fields,omitempty"`
	ReviewedAt               *time.Time    `json:"reviewed_at,omitempty"`
	SubmittedAt              time.Time     `json:"submitted_at"`
	CreatedAt                time.Time     `json:"created_at"`
	UpdatedAt                time.Time     `json:"updated_at"`
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

// TransferApplicationRequest 转交当前审批环节。
type TransferApplicationRequest struct {
	TargetUserID string `json:"target_user_id" binding:"required,uuid"`
	Comment      string `json:"comment" binding:"max=1000"`
}

// TransferCandidate 可转交的处理人。
type TransferCandidate struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	DepartmentName string `json:"department_name,omitempty"`
}

// ApplicationProgressStep 入会流程可视化步骤。
type ApplicationProgressStep struct {
	ID            string `json:"id"`
	Label         string `json:"label"`
	Type          string `json:"type"`
	Role          string `json:"role,omitempty"`
	State         string `json:"state"`
	ApprovalType  string `json:"approval_type,omitempty"`
	AllowTransfer bool   `json:"allow_transfer,omitempty"`
}

// ApplicationProgressResponse 入会流程进度。
type ApplicationProgressResponse struct {
	ApplicationID  string                    `json:"application_id"`
	InstanceID     string                    `json:"instance_id"`
	Status         int                       `json:"status"`
	CurrentNodeIDs []string                  `json:"current_node_ids"`
	Completed      bool                      `json:"completed"`
	Terminated     bool                      `json:"terminated"`
	Editable       bool                      `json:"editable"`
	Steps          []ApplicationProgressStep `json:"steps"`
}
