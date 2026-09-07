package dto

import "time"

type CreateContractRequest struct {
	Title        string     `json:"title" binding:"required,max=200"`
	ContractType int16      `json:"contract_type" binding:"required,oneof=1 2 3 4"`
	PartyName    string     `json:"party_name" binding:"required,max=200"`
	Amount       *float64   `json:"amount"`
	TemplateID   string     `json:"template_id"`
	FileID       string     `json:"file_id"`
	StartAt      *time.Time `json:"start_at"`
	ExpiredAt    *time.Time `json:"expired_at"`
	Status       *int16     `json:"status" binding:"omitempty,oneof=0 1 2 3"`
}

type UpdateContractRequest struct {
	Title        *string    `json:"title" binding:"omitempty,max=200"`
	ContractType *int16     `json:"contract_type" binding:"omitempty,oneof=1 2 3 4"`
	PartyName    *string    `json:"party_name" binding:"omitempty,max=200"`
	Amount       *float64   `json:"amount"`
	TemplateID   *string    `json:"template_id"`
	FileID       *string    `json:"file_id"`
	StartAt      *time.Time `json:"start_at"`
	ExpiredAt    *time.Time `json:"expired_at"`
	Status       *int16     `json:"status" binding:"omitempty,oneof=0 1 2 3"`
}

type ListContractRequest struct {
	Page         int    `form:"page"`
	PageSize     int    `form:"page_size"`
	Keyword      string `form:"keyword"`
	Status       *int16 `form:"status"`
	ContractType *int16 `form:"contract_type"`
}

type ExpiringQuery struct {
	Days int `form:"days"`
}

type Person struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type TemplateResponse struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Code    string `json:"code"`
	Content string `json:"content"`
}

type ContractResponse struct {
	ID           string     `json:"id"`
	Title        string     `json:"title"`
	Status       int16      `json:"status"`
	ContractType int16      `json:"contract_type"`
	PartyName    string     `json:"party_name"`
	Amount       *float64   `json:"amount,omitempty"`
	TemplateID   string     `json:"template_id,omitempty"`
	TemplateName string     `json:"template_name,omitempty"`
	FileID       string     `json:"file_id,omitempty"`
	FileName     string     `json:"file_name,omitempty"`
	Owner        Person     `json:"owner"`
	StartAt      *time.Time `json:"start_at,omitempty"`
	SignedAt     *time.Time `json:"signed_at,omitempty"`
	ExpiredAt    *time.Time `json:"expired_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}
