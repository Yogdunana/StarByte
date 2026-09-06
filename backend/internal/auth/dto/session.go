package dto

import "time"

// SessionView is one online access-token session shown in the admin UI.
type SessionView struct {
	TokenID     string    `json:"token_id"`
	UserID      string    `json:"user_id"`
	Username    string    `json:"username"`
	RealName    string    `json:"real_name"`
	IP          string    `json:"ip"`
	UserAgent   string    `json:"user_agent"`
	Browser     string    `json:"browser"`
	OS          string    `json:"os"`
	Device      string    `json:"device"`
	LoginAt     time.Time `json:"login_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	MultiDevice bool      `json:"multi_device"`
	MultiIP     bool      `json:"multi_ip"`
}

// SessionListResponse is GET /auth/sessions.
type SessionListResponse struct {
	List  []SessionView `json:"list"`
	Total int64         `json:"total"`
}

// UserSessionsResponse is GET /auth/sessions/:user_id.
type UserSessionsResponse struct {
	UserID      string        `json:"user_id"`
	Username    string        `json:"username"`
	RealName    string        `json:"real_name"`
	MultiDevice bool          `json:"multi_device"`
	MultiIP     bool          `json:"multi_ip"`
	Sessions    []SessionView `json:"sessions"`
}
