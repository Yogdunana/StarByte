package model

import "time"

// Session represents a user session stored in Redis.
// It is created on login and used for session management.
type Session struct {
	UserID       string    `json:"user_id"`
	Username     string    `json:"username"`
	RefreshToken string    `json:"refresh_token"` // The random refresh token string
	UserAgent    string    `json:"user_agent"`
	IP           string    `json:"ip"`
	LoginAt      time.Time `json:"login_at"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// LoginAttempt tracks login attempts for lockout detection.
type LoginAttempt struct {
	Username  string    `json:"username"`
	IP        string    `json:"ip"`
	Success   bool      `json:"success"`
	AttemptAt time.Time `json:"attempt_at"`
}
