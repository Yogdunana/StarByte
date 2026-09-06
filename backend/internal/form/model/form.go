package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	StatusDraft     int16 = 0
	StatusPublished int16 = 1
	StatusDisabled  int16 = 2
)

var AllowedFieldTypes = map[string]struct{}{
	"text": {}, "textarea": {}, "number": {}, "select": {}, "radio": {},
	"checkbox": {}, "date": {}, "datetime": {}, "file": {}, "rating": {},
	"switch": {}, "cascader": {},
}

type Form struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey"`
	Name        string     `gorm:"type:varchar(100);not null"`
	Description string     `gorm:"type:varchar(500);not null;default:''"`
	Status      int16      `gorm:"type:smallint;not null;default:0;index"`
	Fields      JSONFields `gorm:"type:jsonb;not null"`
	CreatedBy   *uuid.UUID `gorm:"type:uuid"`
	UpdatedBy   *uuid.UUID `gorm:"type:uuid"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (Form) TableName() string { return "forms" }

type Submission struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey"`
	FormID      uuid.UUID  `gorm:"type:uuid;index;not null"`
	Data        JSONMap    `gorm:"type:jsonb;not null"`
	SubmittedBy *uuid.UUID `gorm:"type:uuid"`
	SubmittedAt time.Time
}

func (Submission) TableName() string { return "form_submissions" }

type FieldOption struct {
	Label    string        `json:"label"`
	Value    interface{}   `json:"value"`
	Children []FieldOption `json:"children,omitempty"`
}

type FieldValidation struct {
	MinLength *int     `json:"min_length,omitempty"`
	MaxLength *int     `json:"max_length,omitempty"`
	MinValue  *float64 `json:"min_value,omitempty"`
	MaxValue  *float64 `json:"max_value,omitempty"`
	Pattern   string   `json:"pattern,omitempty"`
	Message   string   `json:"message,omitempty"`
}

type VisibleWhen struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

type FormField struct {
	Name        string                 `json:"name"`
	Label       string                 `json:"label"`
	Type        string                 `json:"type"`
	Required    bool                   `json:"required"`
	Placeholder string                 `json:"placeholder,omitempty"`
	Default     interface{}            `json:"default,omitempty"`
	Options     []FieldOption          `json:"options,omitempty"`
	Validation  *FieldValidation       `json:"validation,omitempty"`
	VisibleWhen *VisibleWhen           `json:"visible_when,omitempty"`
	Props       map[string]interface{} `json:"props,omitempty"`
}

type JSONFields []FormField

func (j JSONFields) Value() (driver.Value, error) {
	if j == nil {
		return "[]", nil
	}
	b, err := json.Marshal(j)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

func (j *JSONFields) Scan(value interface{}) error {
	bytes, err := asBytes(value)
	if err != nil {
		return err
	}
	if len(bytes) == 0 {
		*j = JSONFields{}
		return nil
	}
	return json.Unmarshal(bytes, j)
}

type JSONMap map[string]interface{}

func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return "{}", nil
	}
	b, err := json.Marshal(j)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

func (j *JSONMap) Scan(value interface{}) error {
	bytes, err := asBytes(value)
	if err != nil {
		return err
	}
	if len(bytes) == 0 {
		*j = JSONMap{}
		return nil
	}
	return json.Unmarshal(bytes, j)
}

func asBytes(value interface{}) ([]byte, error) {
	if value == nil {
		return nil, nil
	}
	switch v := value.(type) {
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	default:
		return nil, fmt.Errorf("unsupported jsonb type %T", value)
	}
}
