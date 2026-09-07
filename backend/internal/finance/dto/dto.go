package dto

import "time"

type CreateRecordRequest struct {
	CategoryID   string    `json:"category_id" binding:"required"`
	DepartmentID string    `json:"department_id"`
	Amount       float64   `json:"amount" binding:"required"`
	Direction    int16     `json:"direction" binding:"required,oneof=1 2"`
	OccurredAt   time.Time `json:"occurred_at" binding:"required"`
	Title        string    `json:"title" binding:"required,max=200"`
	Remark       string    `json:"remark"`
}

type UpdateRecordRequest struct {
	CategoryID   *string    `json:"category_id"`
	DepartmentID *string    `json:"department_id"`
	Amount       *float64   `json:"amount"`
	Direction    *int16     `json:"direction" binding:"omitempty,oneof=1 2"`
	OccurredAt   *time.Time `json:"occurred_at"`
	Title        *string    `json:"title" binding:"omitempty,max=200"`
	Remark       *string    `json:"remark"`
}

type ListRecordRequest struct {
	Page         int    `form:"page"`
	PageSize     int    `form:"page_size"`
	Keyword      string `form:"keyword"`
	Direction    *int16 `form:"direction"`
	CategoryID   string `form:"category_id"`
	DepartmentID string `form:"department_id"`
	From         string `form:"from"`
	To           string `form:"to"`
}

type SummaryQuery struct {
	From       string `form:"from"`
	To         string `form:"to"`
	CategoryID string `form:"category_id"`
}

type Person struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type CategoryResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	Direction   int16  `json:"direction"`
	Description string `json:"description"`
}

type RecordResponse struct {
	ID             string    `json:"id"`
	CategoryID     string    `json:"category_id"`
	CategoryName   string    `json:"category_name"`
	DepartmentID   string    `json:"department_id,omitempty"`
	DepartmentName string    `json:"department_name,omitempty"`
	Amount         float64   `json:"amount"`
	Direction      int16     `json:"direction"`
	OccurredAt     time.Time `json:"occurred_at"`
	Title          string    `json:"title"`
	Remark         string    `json:"remark,omitempty"`
	CreatedBy      *Person   `json:"created_by,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type CategorySum struct {
	CategoryID   string  `json:"category_id"`
	CategoryName string  `json:"category_name"`
	Direction    int16   `json:"direction"`
	Total        float64 `json:"total"`
	Count        int64   `json:"count"`
}

type SummaryResponse struct {
	IncomeTotal  float64       `json:"income_total"`
	ExpenseTotal float64       `json:"expense_total"`
	Balance      float64       `json:"balance"`
	IncomeCount  int64         `json:"income_count"`
	ExpenseCount int64         `json:"expense_count"`
	ByCategory   []CategorySum `json:"by_category"`
}
