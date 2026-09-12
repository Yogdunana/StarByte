package dto

import "time"

type AutoAssignmentConfig struct {
	Mode   string `json:"mode"`
	RoleID string `json:"role_id,omitempty"`
}

type WorkflowConfig struct {
	Assignment *AutoAssignmentConfig `json:"assignment,omitempty"`
	ReviewerID string                `json:"reviewer_id" binding:"required,uuid"`
	AcceptorID string                `json:"acceptor_id" binding:"required,uuid"`
}
type WorkflowActionRequest struct {
	Revision int64  `json:"revision" binding:"required,min=1"`
	Action   string `json:"action" binding:"required,oneof=start pause resume submit approve return reject claim"`
	Comment  string `json:"comment" binding:"required,max=5000"`
}
type WorkflowResponse struct {
	CanStart       bool              `json:"can_start"`
	CanPause       bool              `json:"can_pause"`
	CanResume      bool              `json:"can_resume"`
	AssignmentMode string            `json:"assignment_mode"`
	Revision       int64             `json:"revision"`
	TaskID         string            `json:"task_id"`
	Title          string            `json:"title"`
	InstanceID     string            `json:"instance_id"`
	Stage          string            `json:"stage"`
	Submission     string            `json:"submission"`
	Creator        Person            `json:"creator"`
	Assignee       *Person           `json:"assignee,omitempty"`
	Reviewer       Person            `json:"reviewer"`
	Acceptor       Person            `json:"acceptor"`
	CanSubmit      bool              `json:"can_submit"`
	CanApprove     bool              `json:"can_approve"`
	CanReturn      bool              `json:"can_return"`
	CanReject      bool              `json:"can_reject"`
	CanClaim       bool              `json:"can_claim"`
	CanDelegate    bool              `json:"can_delegate"`
	Handover       *HandoverResponse `json:"handover,omitempty"`
	UpdatedAt      time.Time         `json:"updated_at"`
	History        []LogResponse     `json:"history"`
}
