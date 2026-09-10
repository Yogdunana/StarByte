package dto

import "github.com/Yogdunana/StarByte/backend/internal/member/model"

type SignAdmissionRequest struct {
	Stage            string   `json:"stage" binding:"required"`
	Revision         int      `json:"revision" binding:"required,min=1"`
	RequiredFields   []string `json:"required_fields,omitempty"`
	Role             string   `json:"role" binding:"required,oneof=materials minister center president"`
	Decision         string   `json:"decision" binding:"required,oneof=approve reject supplement"`
	Comment          string   `json:"comment" binding:"max=1000"`
	DelegationReason string   `json:"delegation_reason" binding:"max=1000"`
}

type AdmissionResponse struct {
	Objections               []model.AdmissionObjection `json:"objections"`
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
