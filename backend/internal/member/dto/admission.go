package dto

import (
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/member/model"
)

type SignAdmissionRequest struct {
	Stage            string   `json:"stage" binding:"required"`
	Revision         int      `json:"revision" binding:"required,min=1"`
	RequiredFields   []string `json:"required_fields,omitempty"`
	Role             string   `json:"role" binding:"required,oneof=materials minister center president"`
	Decision         string   `json:"decision" binding:"required,oneof=approve reject supplement"`
	Comment          string   `json:"comment" binding:"max=1000"`
	DelegationReason string   `json:"delegation_reason" binding:"max=1000"`
}

type AdmissionObjectionView struct {
	ID               string    `json:"id"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"created_at"`
	RaisedBy         string    `json:"raised_by,omitempty"`
	Reason           string    `json:"reason,omitempty"`
	CenterReviewerID string    `json:"center_reviewer_id,omitempty"`
	CenterComment    string    `json:"center_comment,omitempty"`
	FinalReviewerID  string    `json:"final_reviewer_id,omitempty"`
	FinalComment     string    `json:"final_comment,omitempty"`
}

type AdmissionResponse struct {
	Objections               []AdmissionObjectionView   `json:"objections"`
	AllowedObjectionActions  []string                   `json:"allowed_objection_actions"`
	ApplicationID            string                     `json:"application_id"`
	Revision                 int                        `json:"revision"`
	Stage                    string                     `json:"stage"`
	HistoricalReviewRequired bool                       `json:"historical_review_required"`
	Signatures               []model.AdmissionSignature `json:"signatures"`
	AllowedRoles             []string                   `json:"allowed_roles"`
	InterviewCompleted       bool                       `json:"interview_completed"`
}

type AdmissionObjectionRequest struct {
	Action  string `json:"action" binding:"required,oneof=raise center_review uphold dismiss"`
	Comment string `json:"comment" binding:"required,max=2000"`
}
