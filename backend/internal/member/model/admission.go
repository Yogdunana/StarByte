package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	AdmissionMaterials  = "materials"
	AdmissionRound1     = "round1"
	AdmissionRound2     = "round2"
	AdmissionPresident  = "president"
	AdmissionProbation  = "probation"
	AdmissionApproved   = "approved"
	AdmissionRejected   = "rejected"
	AdmissionSupplement = "supplement"
	AdmissionLegacy     = "legacy_review"
	AdmissionEngine     = "engine"
)

type AdmissionSignature struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ApplicationID    uuid.UUID `gorm:"type:uuid;not null;index" json:"application_id"`
	Revision         int       `gorm:"not null;default:1" json:"revision"`
	Stage            string    `gorm:"size:40;not null" json:"stage"`
	SignerID         uuid.UUID `gorm:"type:uuid;not null" json:"signer_id"`
	SignerName       string    `gorm:"->" json:"signer_name"`
	SignerRole       string    `gorm:"size:40;not null" json:"signer_role"`
	Decision         string    `gorm:"size:20;not null" json:"decision"`
	Comment          string    `gorm:"type:text" json:"comment"`
	Delegated        bool      `json:"delegated"`
	DelegationReason string    `gorm:"type:text" json:"delegation_reason"`
	CreatedAt        time.Time `json:"created_at"`
}

func (AdmissionSignature) TableName() string { return "admission_signatures" }

type AdmissionActor struct {
	ID           uuid.UUID
	DepartmentID *uuid.UUID
	Roles        []string `gorm:"-"`
}

type AdmissionObjection struct {
	ID               uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	ApplicationID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"application_id"`
	RaisedBy         uuid.UUID  `gorm:"type:uuid;not null" json:"raised_by"`
	Reason           string     `gorm:"type:text;not null" json:"reason"`
	Status           string     `gorm:"size:30;not null" json:"status"`
	CenterReviewerID *uuid.UUID `gorm:"type:uuid" json:"center_reviewer_id,omitempty"`
	CenterComment    string     `gorm:"type:text" json:"center_comment"`
	FinalReviewerID  *uuid.UUID `gorm:"type:uuid" json:"final_reviewer_id,omitempty"`
	FinalComment     string     `gorm:"type:text" json:"final_comment"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

func (AdmissionObjection) TableName() string { return "admission_objections" }
