package model

import "github.com/google/uuid"

// Elector contains frozen eligibility and weight, never a ballot ID or choice.
type Elector struct {
	VoteID uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID uuid.UUID `gorm:"type:uuid;primaryKey"`
	Weight float64   `gorm:"type:numeric(8,2);not null"`
}

func (Elector) TableName() string { return "meeting_vote_electorates" }

type ElectorCandidate struct {
	UserID       uuid.UUID
	PositionCode string
	RoleCode     string
}
