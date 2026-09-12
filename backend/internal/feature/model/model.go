package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Flag is a feature switch evaluated without process restart.
type Flag struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	FlagKey     string     `gorm:"column:flag_key;type:varchar(128);uniqueIndex;not null" json:"flag_key"`
	Name        string     `gorm:"type:varchar(128);not null" json:"name"`
	Description string     `gorm:"type:text;not null;default:''" json:"description"`
	FlagType    string     `gorm:"type:varchar(32);not null" json:"flag_type"`
	Enabled     bool       `gorm:"not null;default:false" json:"enabled"`
	GroupName   string     `gorm:"type:varchar(64);not null;default:''" json:"group_name"`
	Priority    int        `gorm:"not null;default:0" json:"priority"`
	Rules       Rules      `gorm:"type:jsonb;not null" json:"rules"`
	IsSystem    bool       `gorm:"not null;default:false" json:"is_system"`
	CreatedBy   *uuid.UUID `gorm:"type:uuid" json:"created_by,omitempty"`
	UpdatedBy   *uuid.UUID `gorm:"type:uuid" json:"updated_by,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (Flag) TableName() string { return "feature_flags" }

// Rules is type-specific targeting. Unused fields are ignored.
type Rules struct {
	UserIDs       []string `json:"user_ids,omitempty"`
	RoleCodes     []string `json:"role_codes,omitempty"`
	DepartmentIDs []string `json:"department_ids,omitempty"`
	Percent       int      `json:"percent,omitempty"`
	Salt          string   `json:"salt,omitempty"`
}

func (r Rules) Value() (driver.Value, error) {
	b, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (r *Rules) Scan(value interface{}) error {
	if value == nil {
		*r = Rules{}
		return nil
	}
	var raw []byte
	switch v := value.(type) {
	case []byte:
		raw = v
	case string:
		raw = []byte(v)
	default:
		return fmt.Errorf("feature rules: unsupported type %T", value)
	}
	if len(raw) == 0 {
		*r = Rules{}
		return nil
	}
	return json.Unmarshal(raw, r)
}

// Audit records a toggle or mutation.
type Audit struct {
	ID         uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	FlagID     *uuid.UUID      `gorm:"type:uuid" json:"flag_id,omitempty"`
	FlagKey    string          `gorm:"type:varchar(128);not null" json:"flag_key"`
	Action     string          `gorm:"type:varchar(32);not null" json:"action"`
	ActorID    *uuid.UUID      `gorm:"type:uuid" json:"actor_id,omitempty"`
	BeforeJSON json.RawMessage `gorm:"type:jsonb" json:"before_json,omitempty"`
	AfterJSON  json.RawMessage `gorm:"type:jsonb" json:"after_json,omitempty"`
	Reason     string          `gorm:"type:text;not null;default:''" json:"reason"`
	CreatedAt  time.Time       `json:"created_at"`
}

func (Audit) TableName() string { return "feature_flag_audits" }
