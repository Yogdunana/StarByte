package dto

import (
	"encoding/json"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/feature/model"
)

type ListQuery struct {
	Page      int    `form:"page"`
	PageSize  int    `form:"page_size"`
	Keyword   string `form:"keyword"`
	GroupName string `form:"group"`
	Enabled   *bool  `form:"enabled"`
}

type CreateFlagRequest struct {
	FlagKey     string      `json:"flag_key" binding:"required,max=128"`
	Name        string      `json:"name" binding:"required,max=128"`
	Description string      `json:"description" binding:"max=2000"`
	FlagType    string      `json:"flag_type" binding:"required"`
	Enabled     bool        `json:"enabled"`
	GroupName   string      `json:"group_name" binding:"max=64"`
	Priority    int         `json:"priority"`
	Rules       model.Rules `json:"rules"`
}

type UpdateFlagRequest struct {
	Name        *string      `json:"name" binding:"omitempty,max=128"`
	Description *string      `json:"description" binding:"omitempty,max=2000"`
	FlagType    *string      `json:"flag_type"`
	Enabled     *bool        `json:"enabled"`
	GroupName   *string      `json:"group_name" binding:"omitempty,max=64"`
	Priority    *int         `json:"priority"`
	Rules       *model.Rules `json:"rules"`
}

type ToggleRequest struct {
	Enabled *bool  `json:"enabled"`
	Reason  string `json:"reason" binding:"max=500"`
}

type EvaluateQuery struct {
	UserID string `form:"user_id"`
}

type EvaluateMeQuery struct {
	Keys string `form:"keys"`
}

type AuditQuery struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	FlagKey  string `form:"flag_key"`
}

type FlagResponse struct {
	ID          string      `json:"id"`
	FlagKey     string      `json:"flag_key"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	FlagType    string      `json:"flag_type"`
	Enabled     bool        `json:"enabled"`
	GroupName   string      `json:"group_name"`
	Priority    int         `json:"priority"`
	Rules       model.Rules `json:"rules"`
	IsSystem    bool        `json:"is_system"`
	CreatedBy   string      `json:"created_by,omitempty"`
	UpdatedBy   string      `json:"updated_by,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

type EvaluateResponse struct {
	Key     string `json:"key"`
	Enabled bool   `json:"enabled"`
	Reason  string `json:"reason"`
	Type    string `json:"flag_type,omitempty"`
}

type AuditResponse struct {
	ID         string          `json:"id"`
	FlagID     string          `json:"flag_id,omitempty"`
	FlagKey    string          `json:"flag_key"`
	Action     string          `json:"action"`
	ActorID    string          `json:"actor_id,omitempty"`
	BeforeJSON json.RawMessage `json:"before_json,omitempty"`
	AfterJSON  json.RawMessage `json:"after_json,omitempty"`
	Reason     string          `json:"reason"`
	CreatedAt  time.Time       `json:"created_at"`
}

func ToFlag(f *model.Flag) FlagResponse {
	out := FlagResponse{
		ID: f.ID.String(), FlagKey: f.FlagKey, Name: f.Name, Description: f.Description,
		FlagType: f.FlagType, Enabled: f.Enabled, GroupName: f.GroupName, Priority: f.Priority,
		Rules: f.Rules, IsSystem: f.IsSystem, CreatedAt: f.CreatedAt, UpdatedAt: f.UpdatedAt,
	}
	if f.CreatedBy != nil {
		out.CreatedBy = f.CreatedBy.String()
	}
	if f.UpdatedBy != nil {
		out.UpdatedBy = f.UpdatedBy.String()
	}
	return out
}

func ToAudit(a *model.Audit) AuditResponse {
	out := AuditResponse{
		ID: a.ID.String(), FlagKey: a.FlagKey, Action: a.Action,
		BeforeJSON: a.BeforeJSON, AfterJSON: a.AfterJSON, Reason: a.Reason, CreatedAt: a.CreatedAt,
	}
	if a.FlagID != nil {
		out.FlagID = a.FlagID.String()
	}
	if a.ActorID != nil {
		out.ActorID = a.ActorID.String()
	}
	return out
}
