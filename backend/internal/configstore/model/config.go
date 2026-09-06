package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	TypeString  = "string"
	TypeNumber  = "number"
	TypeBoolean = "boolean"
	TypeJSON    = "json"
)

// Config 对应已有 configs 表，不改列名。
type Config struct {
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

func (Config) TableName() string { return "configs" }

func ValidType(t string) bool {
	switch t {
	case TypeString, TypeNumber, TypeBoolean, TypeJSON:
		return true
	default:
		return false
	}
}

func ValidCategory(c string) bool {
	switch c {
	case "system", "business", "notification", "security", "internship", "meeting":
		return true
	default:
		return false
	}
}

// ProtectedKeys 业务模块已占用，禁止删除。
func ProtectedKeys() map[string]struct{} {
	return map[string]struct{}{
		"internship_config":  {},
		"vote_weight_config": {},
	}
}
