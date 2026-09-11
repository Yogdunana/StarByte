package dto

import "time"

// ========== Request DTOs ==========

// LoginRequest 登录请求（username 可为用户名或学号）
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

// CASExchangeRequest 用一次性 code 换取系统 JWT
type CASExchangeRequest struct {
	Code string `json:"code" binding:"required"`
}

// CASRegisterRequest 用 CAS 续传凭证创建本地账号并绑定学号
type CASRegisterRequest struct {
	Token    string `json:"token" binding:"required"`
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=8,max=50"`
	RealName string `json:"real_name" binding:"omitempty,max=50"`
	Email    string `json:"email" binding:"omitempty,email"`
}

// CASStatusResponse 学校统一认证是否开通
type CASStatusResponse struct {
	Enabled bool `json:"enabled"`
}

// CASLoginStart 跳转学校 CAS 所需的 Location 与短时 state（写入 Cookie）
type CASLoginStart struct {
	Location string `json:"location"`
	State    string `json:"state"`
	Service  string `json:"service"`
}

// CASExchangeResponse 兑换 CAS 回调 code 后的登录结果；未知用户返回 needs_registration。
type CASExchangeResponse struct {
	LoginResponse
	Redirect          string `json:"redirect"`
	NeedsRegistration bool   `json:"needs_registration,omitempty"`
	RegistrationToken string `json:"registration_token,omitempty"`
	StudentNo         string `json:"student_no,omitempty"`
	RealName          string `json:"real_name,omitempty"`
	Email             string `json:"email,omitempty"`
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

// UserInfo 用户信息（登录与 /auth/me 返回，含档案学号/姓名等）
type UserInfo struct {
	ID             string    `json:"id"`
	Username       string    `json:"username"`
	RealName       string    `json:"real_name"`
	StudentNo      string    `json:"student_no"`
	Grade          string    `json:"grade"`
	Major          string    `json:"major"`
	DepartmentID   string    `json:"department_id,omitempty"`
	DepartmentName string    `json:"department_name"`
	PositionID     string    `json:"position_id,omitempty"`
	PositionName   string    `json:"position_name"`
	AvatarURL      string    `json:"avatar_url"`
	Email          string    `json:"email"`
	Phone          string    `json:"phone"`
	Gender         int       `json:"gender"`
	Status         int       `json:"status"`
	Roles          []string  `json:"roles"`
	Permissions    []string  `json:"permissions"`
	CreatedAt      time.Time `json:"created_at"`
}

// RefreshResponse 刷新 Token 响应
type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}
