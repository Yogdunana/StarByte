package dto

// ========== 请求 DTO ==========

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6,max=50"`
	RealName string `json:"real_name" binding:"omitempty,max=50"`
	Email    string `json:"email" binding:"omitempty,email"`
	Phone    string `json:"phone" binding:"omitempty,max=20"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RefreshTokenRequest 刷新 Token 请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// UpdateProfileRequest 更新个人信息请求
type UpdateProfileRequest struct {
	RealName  string `json:"real_name" binding:"omitempty,max=50"`
	AvatarURL string `json:"avatar_url" binding:"omitempty,max=500"`
	Email     string `json:"email" binding:"omitempty,email"`
	Phone     string `json:"phone" binding:"omitempty,max=20"`
	Gender    *int   `json:"gender" binding:"omitempty,oneof=0 1 2"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=50"`
}

// ListUserRequest 用户列表请求
type ListUserRequest struct {
	Page         int    `form:"page,default=1" binding:"min=1"`
	PageSize     int    `form:"page_size,default=10" binding:"min=1,max=100"`
	Keyword      string `form:"keyword"`
	Status       *int   `form:"status" binding:"omitempty,oneof=0 1 2"`
	DepartmentID string `form:"department_id"`
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username     string   `json:"username" binding:"required,min=3,max=50"`
	Password     string   `json:"password" binding:"required,min=6,max=50"`
	RealName     string   `json:"real_name" binding:"omitempty,max=50"`
	Email        string   `json:"email" binding:"omitempty,email"`
	Phone        string   `json:"phone" binding:"omitempty,max=20"`
	Gender       *int     `json:"gender" binding:"omitempty,oneof=0 1 2"`
	DepartmentID string   `json:"department_id"`
	PositionID   string   `json:"position_id"`
	RoleIDs      []string `json:"role_ids"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	RealName     string   `json:"real_name" binding:"omitempty,max=50"`
	Email        string   `json:"email" binding:"omitempty,email"`
	Phone        string   `json:"phone" binding:"omitempty,max=20"`
	Gender       *int     `json:"gender" binding:"omitempty,oneof=0 1 2"`
	Status       *int     `json:"status" binding:"omitempty,oneof=0 1 2"`
	DepartmentID string   `json:"department_id"`
	PositionID   string   `json:"position_id"`
	RoleIDs      []string `json:"role_ids"`
}

// ========== 响应 DTO ==========

// LoginResponse 登录响应
type LoginResponse struct {
	AccessToken         string            `json:"access_token"`
	RefreshToken        string            `json:"refresh_token"`
	AccessTokenExpires  int64             `json:"access_token_expires"`
	RefreshTokenExpires int64             `json:"refresh_token_expires"`
	User                *UserInfoResponse `json:"user"`
}

// UserInfoResponse 用户信息响应
type UserInfoResponse struct {
	ID           string             `json:"id"`
	Username     string             `json:"username"`
	RealName     string             `json:"real_name"`
	AvatarURL    string             `json:"avatar_url"`
	Email        string             `json:"email"`
	Phone        string             `json:"phone"`
	Gender       int                `json:"gender"`
	Status       int                `json:"status"`
	DepartmentID string             `json:"department_id"`
	PositionID   string             `json:"position_id"`
	Roles        []RoleInfoResponse `json:"roles"`
	Permissions  []string           `json:"permissions"`
	CreatedAt    string             `json:"created_at"`
}

// RoleInfoResponse 角色信息响应
type RoleInfoResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

// UserListResponse 用户列表项响应
type UserListResponse struct {
	ID             string `json:"id"`
	Username       string `json:"username"`
	RealName       string `json:"real_name"`
	AvatarURL      string `json:"avatar_url"`
	Email          string `json:"email"`
	Phone          string `json:"phone"`
	Gender         int    `json:"gender"`
	Status         int    `json:"status"`
	DepartmentID   string `json:"department_id"`
	DepartmentName string `json:"department_name"`
	PositionID     string `json:"position_id"`
	PositionName   string `json:"position_name"`
	LastLoginAt    string `json:"last_login_at"`
	CreatedAt      string `json:"created_at"`
}
