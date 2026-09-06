package dto

import "time"

type ConfigResponse struct {
	ID          string    `json:"id"`
	ConfigKey   string    `json:"config_key"`
	ConfigValue string    `json:"config_value"`
	ConfigType  string    `json:"config_type"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	IsPublic    bool      `json:"is_public"`
	UpdatedAt   time.Time `json:"updated_at"`
	CreatedAt   time.Time `json:"created_at"`
}

type CreateConfigRequest struct {
	ConfigKey   string `json:"config_key" binding:"required,max=100"`
	ConfigValue string `json:"config_value"`
	ConfigType  string `json:"config_type" binding:"required,oneof=string number boolean json"`
	Description string `json:"description" binding:"max=255"`
	Category    string `json:"category" binding:"required,max=50"`
	IsPublic    bool   `json:"is_public"`
}

type UpdateConfigRequest struct {
	ConfigValue *string `json:"config_value"`
	ConfigType  *string `json:"config_type" binding:"omitempty,oneof=string number boolean json"`
	Description *string `json:"description" binding:"omitempty,max=255"`
	Category    *string `json:"category" binding:"omitempty,max=50"`
	IsPublic    *bool   `json:"is_public"`
}

type ListQuery struct {
	Category string `form:"category"`
	Keyword  string `form:"keyword"`
}
