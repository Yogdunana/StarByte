package dto

import "time"

type TypeResponse struct {
	ID          string    `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	SortOrder   int       `json:"sort_order"`
	Status      int16     `json:"status"`
	IsSystem    bool      `json:"is_system"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ItemResponse struct {
	ID        string    `json:"id"`
	TypeID    string    `json:"type_id"`
	TypeCode  string    `json:"type_code"`
	ItemValue string    `json:"item_value"`
	ItemLabel string    `json:"item_label"`
	SortOrder int       `json:"sort_order"`
	Status    int16     `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateTypeRequest struct {
	Code        string `json:"code" binding:"required,max=50"`
	Name        string `json:"name" binding:"required,max=100"`
	Description string `json:"description" binding:"max=255"`
	SortOrder   int    `json:"sort_order"`
	Status      int16  `json:"status" binding:"omitempty,oneof=0 1"`
}

type UpdateTypeRequest struct {
	Name        *string `json:"name" binding:"omitempty,max=100"`
	Description *string `json:"description" binding:"omitempty,max=255"`
	SortOrder   *int    `json:"sort_order"`
	Status      *int16  `json:"status" binding:"omitempty,oneof=0 1"`
}

type CreateItemRequest struct {
	TypeCode  string `json:"type_code" binding:"required,max=50"`
	ItemValue string `json:"item_value" binding:"required,max=50"`
	ItemLabel string `json:"item_label" binding:"required,max=100"`
	SortOrder int    `json:"sort_order"`
	Status    int16  `json:"status" binding:"omitempty,oneof=0 1"`
}

type UpdateItemRequest struct {
	ItemLabel *string `json:"item_label" binding:"omitempty,max=100"`
	SortOrder *int    `json:"sort_order"`
	Status    *int16  `json:"status" binding:"omitempty,oneof=0 1"`
}
