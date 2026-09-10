package dto

import "time"

// Person 通用人员信息
type Person struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// CreateActivityRequest 创建活动
type CreateActivityRequest struct {
	Title           string    `json:"title" binding:"required,max=200"`
	Description     string    `json:"description"`
	CoverImageID    string    `json:"cover_image_id"`
	Category        string    `json:"category" binding:"max=50"`
	Tags            []string  `json:"tags"`
	StartTime       time.Time `json:"start_time" binding:"required"`
	EndTime         time.Time `json:"end_time" binding:"required"`
	Location        string    `json:"location" binding:"max=200"`
	MaxParticipants int       `json:"max_participants" binding:"min=0"`
}

// UpdateActivityRequest 更新活动
type UpdateActivityRequest struct {
	Title           *string    `json:"title"`
	Description     *string    `json:"description"`
	CoverImageID    *string    `json:"cover_image_id"`
	Category        *string    `json:"category"`
	Tags            []string   `json:"tags"`
	StartTime       *time.Time `json:"start_time"`
	EndTime         *time.Time `json:"end_time"`
	Location        *string    `json:"location"`
	MaxParticipants *int       `json:"max_participants"`
}

// ListActivityRequest 活动列表
type ListActivityRequest struct {
	Page      int    `form:"page"`
	PageSize  int    `form:"page_size"`
	Status    *int16 `form:"status"`
	Category  string `form:"category"`
	Keyword   string `form:"keyword"`
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
}

// ActivityResponse 活动响应
type ActivityResponse struct {
	ID              string   `json:"id"`
	Title           string   `json:"title"`
	Description     string   `json:"description"`
	CoverImageID    string   `json:"cover_image_id,omitempty"`
	Category        string   `json:"category"`
	Tags            []string `json:"tags"`
	StartTime       string   `json:"start_time"`
	EndTime         string   `json:"end_time"`
	Location        string   `json:"location"`
	MaxParticipants int      `json:"max_participants"`
	Status          int16    `json:"status"`
	Organizer       Person   `json:"organizer"`
	RegisteredCount int64    `json:"registered_count"`
	CheckedInCount  int64    `json:"checked_in_count"`
	CreatedAt       string   `json:"created_at"`
	UpdatedAt       string   `json:"updated_at"`
}

// RegisterRequest 报名
type RegisterRequest struct {
	// 预留扩展字段
}

// ApproveRegistrationRequest 审批报名
type ApproveRegistrationRequest struct {
	Approve bool   `json:"approve"`
	Reason  string `json:"reason" binding:"max=500"`
}

// CheckinRequest 签到
type CheckinRequest struct {
	Method    int16    `json:"method" binding:"required,oneof=1 2"`
	Token     string   `json:"token"`
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
}

// SurveyRequest 满意度调查
type SurveyRequest struct {
	Rating  int16  `json:"rating" binding:"required,min=1,max=5"`
	Comment string `json:"comment" binding:"max=1000"`
}

// RegistrationResponse 报名记录响应
type RegistrationResponse struct {
	ID            string `json:"id"`
	ActivityID    string `json:"activity_id"`
	User          Person `json:"user"`
	Status        int16  `json:"status"`
	CheckinStatus int16  `json:"checkin_status"`
	CheckedInAt   string `json:"checked_in_at,omitempty"`
	CheckinMethod *int16 `json:"checkin_method,omitempty"`
	CreatedAt     string `json:"created_at"`
}

// ActivityStatsResponse 活动统计
type ActivityStatsResponse struct {
	ActivityID      string           `json:"activity_id"`
	MaxParticipants int              `json:"max_participants"`
	RegisteredCount int64            `json:"registered_count"`
	ApprovedCount   int64            `json:"approved_count"`
	WaitlistCount   int64            `json:"waitlist_count"`
	CheckedInCount  int64            `json:"checked_in_count"`
	RegisterRate    float64          `json:"register_rate"`
	AttendRate      float64          `json:"attend_rate"`
	SurveyCount     int64            `json:"survey_count"`
	AvgRating       float64          `json:"avg_rating"`
	RatingDist      map[string]int64 `json:"rating_distribution"`
}
