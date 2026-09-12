package dto

import "time"

type HandoverRequest struct {
	TargetID string `json:"target_id" binding:"required,uuid"`
	Reason   string `json:"reason" binding:"required,max=2000"`
	Revision int64  `json:"revision" binding:"required,min=1"`
}
type HandoverDecision struct {
	TransferID  string `json:"transfer_id" binding:"omitempty,uuid"`
	Requirement string `json:"requirement" binding:"required"`
	Decision    string `json:"decision" binding:"required,oneof=approve reject"`
	Comment     string `json:"comment" binding:"required,max=2000"`
	Revision    int64  `json:"revision" binding:"required,min=1"`
}
type HandoverSign struct {
	ID          string    `json:"id"`
	Requirement string    `json:"requirement"`
	Signer      Person    `json:"signer"`
	SignerRole  string    `json:"signer_role"`
	Waived      bool      `json:"waived"`
	Decision    string    `json:"decision"`
	Comment     string    `json:"comment"`
	CreatedAt   time.Time `json:"created_at"`
}
type HandoverResponse struct {
	ID               string         `json:"id"`
	TaskID           string         `json:"task_id"`
	TaskTitle        string         `json:"task_title"`
	Kind             string         `json:"kind"`
	Status           string         `json:"status"`
	Revision         int64          `json:"revision"`
	From             Person         `json:"from"`
	To               Person         `json:"to"`
	Reason           string         `json:"reason"`
	SourceDepartment string         `json:"source_department"`
	TargetDepartment string         `json:"target_department"`
	Requirements     []string       `json:"requirements"`
	CanSign          []string       `json:"can_sign"`
	Signatures       []HandoverSign `json:"signatures"`
	CreatedAt        time.Time      `json:"created_at"`
}
