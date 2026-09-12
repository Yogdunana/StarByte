package dto

import "time"

type Person struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type LeaveTypeResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Code        string  `json:"code"`
	Deductible  bool    `json:"deductible"`
	DefaultDays float64 `json:"default_days"`
	Description string  `json:"description"`
}

type LeaveBalanceResponse struct {
	ID            string            `json:"id"`
	UserID        string            `json:"user_id"`
	Year          int               `json:"year"`
	TotalDays     float64           `json:"total_days"`
	UsedDays      float64           `json:"used_days"`
	RemainingDays float64           `json:"remaining_days"`
	LeaveType     LeaveTypeResponse `json:"leave_type"`
}

type SubmitLeaveRequest struct {
	LeaveTypeID string    `json:"leave_type_id" binding:"required,uuid"`
	StartTime   time.Time `json:"start_time" binding:"required"`
	EndTime     time.Time `json:"end_time" binding:"required"`
	Reason      string    `json:"reason" binding:"required,min=1,max=500"`
}

type DecisionRequest struct {
	Remark string `json:"remark" binding:"max=500"`
}

type ListLeaveRequest struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Status   string `form:"status"`
	UserID   string `form:"user_id"`
	Year     int    `form:"year"`
}

type LeaveApplicationResponse struct {
	ID            string            `json:"id"`
	Applicant     Person            `json:"applicant"`
	LeaveType     LeaveTypeResponse `json:"leave_type"`
	StartTime     string            `json:"start_time"`
	EndTime       string            `json:"end_time"`
	DurationDays  float64           `json:"duration_days"`
	Reason        string            `json:"reason"`
	Status        string            `json:"status"`
	Approver      *Person           `json:"approver,omitempty"`
	ApproveRemark string            `json:"approve_remark,omitempty"`
	ApprovedAt    string            `json:"approved_at,omitempty"`
	CreatedAt     string            `json:"created_at"`
	UpdatedAt     string            `json:"updated_at"`
}

type TypeStat struct {
	LeaveTypeID string  `json:"leave_type_id"`
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Count       int64   `json:"count"`
	Days        float64 `json:"days"`
}

type LeaveStatsResponse struct {
	Total    int64            `json:"total"`
	ByStatus map[string]int64 `json:"by_status"`
	ByType   []TypeStat       `json:"by_type"`
}
