package model

import "time"

// Session represents a user session stored in Redis.
// It is created on login and used for session management.
type Session struct {
	UserID    string    `json:"user_id"`
	TokenID   string    `json:"token_id"`
	UserAgent string    `json:"user_agent"`
	IP        string    `json:"ip"`
	LoginAt   time.Time `json:"login_at"`
	ExpiresAt time.Time `json:"expires_at"`
}
