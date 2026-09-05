package dto

import "time"

type Person struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar,omitempty"`
}

type ParentRef struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type CreateTaskRequest struct {
	Title        string     `json:"title" binding:"required,max=200"`
	Description  string     `json:"description"`
	Priority     int16      `json:"priority" binding:"oneof=0 1 2 3"`
	AssigneeID   string     `json:"assignee_id"`
	DepartmentID string     `json:"department_id"`
	DueDate      *time.Time `json:"due_date"`
	Tags         []string   `json:"tags"`
	ParentID     string     `json:"parent_id"`
}

type UpdateTaskRequest struct {
	Title       *string    `json:"title"`
	Description *string    `json:"description"`
	Priority    *int16     `json:"priority"`
	DueDate     *time.Time `json:"due_date"`
	Tags        []string   `json:"tags"`
}

type ListTaskRequest struct {
	Page         int    `form:"page"`
	PageSize     int    `form:"page_size"`
	Status       *int16 `form:"status"`
	Priority     *int16 `form:"priority"`
	AssigneeID   string `form:"assignee_id"`
	DepartmentID string `form:"department_id"`
	Keyword      string `form:"keyword"`
	Tags         string `form:"tags"`
	SortBy       string `form:"sort_by"`
	SortOrder    string `form:"sort_order"`
	ParentID     string `form:"parent_id"`
}

type MyTaskRequest struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Priority *int16 `form:"priority"`
	Keyword  string `form:"keyword"`
	Status   *int16 `form:"status"`
}

type StatsRequest struct {
	DepartmentID string `form:"department_id"`
	StartDate    string `form:"start_date"`
	EndDate      string `form:"end_date"`
}

type TaskResponse struct {
	ID              string      `json:"id"`
	Title           string      `json:"title"`
	Description     string      `json:"description"`
	Priority        int16       `json:"priority"`
	Status          int16       `json:"status"`
	Assignee        *Person     `json:"assignee,omitempty"`
	Creator         Person      `json:"creator"`
	Department      *Person     `json:"department,omitempty"`
	Parent          *ParentRef  `json:"parent,omitempty"`
	Children        []TaskBrief `json:"children"`
	DueDate         *time.Time  `json:"due_date,omitempty"`
	CompletedAt     *time.Time  `json:"completed_at,omitempty"`
	Tags            []string    `json:"tags"`
	CommentCount    int64       `json:"comment_count"`
	AttachmentCount int64       `json:"attachment_count"`
	Progress        int16       `json:"progress"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

type TaskBrief struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Status int16  `json:"status"`
}

type AssignRequest struct {
	AssigneeID string `json:"assignee_id" binding:"required"`
}

type TransferRequest struct {
	NewAssigneeID string `json:"new_assignee_id" binding:"required"`
	Reason        string `json:"reason" binding:"required,max=500"`
}

type StatusRequest struct {
	Status  int16  `json:"status" binding:"oneof=0 1 2 3 4"`
	Comment string `json:"comment" binding:"max=500"`
}

type UrgeRequest struct {
	Message string `json:"message" binding:"max=500"`
}
