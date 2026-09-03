package dto

import "time"

// ========== Request DTOs ==========

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RefreshTokenRequest 刷新 Token 请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8,max=50"`
}

// LogoutRequest 登出请求
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// WechatLoginRequest 微信扫码登录请求（预留）
type WechatLoginRequest struct {
	Code string `json:"code" binding:"required"`
}

// OAuthLoginRequest 第三方 OAuth 登录请求（预留）
type OAuthLoginRequest struct {
	Code  string `json:"code" binding:"required"`
	State string `json:"state"`
}

// ========== Response DTOs ==========

// LoginResponse 登录响应
type LoginResponse struct {
	AccessToken      string    `json:"access_token"`
	RefreshToken     string    `json:"refresh_token"`
	ExpiresIn        int64     `json:"expires_in"` // access token expiry in seconds
	RefreshExpiresIn int64     `json:"refresh_expires_in"`
	User             *UserInfo `json:"user"`
}

// UserInfo 用户信息（登录响应中返回）
type UserInfo struct {
	ID          string    `json:"id"`
	Username    string    `json:"username"`
	RealName    string    `json:"real_name"`
	AvatarURL   string    `json:"avatar_url"`
	Email       string    `json:"email"`
	Phone       string    `json:"phone"`
	Gender      int       `json:"gender"`
	Status      int       `json:"status"`
	Roles       []string  `json:"roles"`
	Permissions []string  `json:"permissions"`
	CreatedAt   time.Time `json:"created_at"`
}

// RefreshResponse 刷新 Token 响应
type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}
