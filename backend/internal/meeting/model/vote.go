package model

import (
	"time"

	"github.com/google/uuid"
)

type Vote struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	MeetingID   uuid.UUID  `gorm:"type:uuid;not null" json:"meeting_id"`
	Title       string     `gorm:"type:varchar(200);not null" json:"title"`
	Description string     `gorm:"type:text" json:"description"`
	VoteType    int16      `gorm:"type:smallint;not null;default:1" json:"vote_type"`
	Status      int16      `gorm:"type:smallint;not null;default:0" json:"status"`
	IsAnonymous bool       `gorm:"not null;default:false" json:"is_anonymous"`
	StartTime   *time.Time `json:"start_time"`
	EndTime     *time.Time `json:"end_time"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (Vote) TableName() string { return "meeting_votes" }

type VoteOption struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	VoteID     uuid.UUID `gorm:"type:uuid;not null" json:"vote_id"`
	OptionText string    `gorm:"type:varchar(200);not null" json:"option_text"`
	OptionKey  string    `gorm:"type:varchar(64);not null" json:"option_key"`
	SortOrder  int       `gorm:"not null;default:0" json:"sort_order"`
}

func (VoteOption) TableName() string { return "meeting_vote_options" }

type VoteRecord struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	VoteID    uuid.UUID `gorm:"type:uuid;not null" json:"vote_id"`
	VoterID   uuid.UUID `gorm:"type:uuid;not null" json:"voter_id"`
	OptionID  uuid.UUID `gorm:"type:uuid;not null" json:"option_id"`
	OptionKey string    `gorm:"type:varchar(64);not null;default:''" json:"option_key"`
	Weight    float64   `gorm:"type:numeric(8,2);not null;default:1" json:"weight"`
	VotedAt   time.Time `json:"voted_at"`
}

func (VoteRecord) TableName() string { return "meeting_vote_records" }

type VoteRecordNamed struct {
	VoteRecord
	VoterName string `gorm:"column:voter_name"`
}

type SystemConfig struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey"`
	ConfigKey   string     `gorm:"column:config_key;type:varchar(100);uniqueIndex"`
	ConfigValue string     `gorm:"column:config_value;type:text"`
	ConfigType  string     `gorm:"column:config_type;type:varchar(20)"`
	Description string     `gorm:"type:varchar(255)"`
	Category    string     `gorm:"type:varchar(50)"`
	IsPublic    bool       `gorm:"column:is_public"`
	UpdatedBy   *uuid.UUID `gorm:"type:uuid"`
	UpdatedAt   time.Time
	CreatedAt   time.Time
}

func (SystemConfig) TableName() string { return "configs" }

type VoteWeightConfig struct {
	Weights       map[string]float64 `json:"weights"`
	DefaultWeight float64            `json:"default_weight"`
}
