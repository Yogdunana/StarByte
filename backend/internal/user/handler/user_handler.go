package handler

import (
	"net/http"

	"github.com/Yogdunana/StarByte/backend/internal/user/dto"
	"github.com/Yogdunana/StarByte/backend/internal/user/service"
	authmiddleware "github.com/Yogdunana/StarByte/backend/pkg/middleware/auth"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UserHandler 用户处理器
type UserHandler struct {
	userService service.UserService
}

// NewUserHandler 创建用户处理器
func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// ========== 认证相关接口 ==========

// Register 注册
// @Summary 用户注册
// @Description 新用户注册
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "注册信息"
// @Success 200 {object} response.Response
// @Router /auth/register [post]
func (h *UserHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	user, err := h.userService.Register(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, gin.H{
		"id":       user.ID.String(),
		"username": user.Username,
	})
}

// Login 登录
// @Summary 用户登录
// @Description 用户名密码登录
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "登录信息"
// @Success 200 {object} response.Response{data=dto.LoginResponse}
// @Router /auth/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	ip := c.ClientIP()
	result, err := h.userService.Login(c.Request.Context(), &req, ip)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, result)
}

// RefreshToken 刷新 Token
// @Summary 刷新 Token
// @Description 使用 Refresh Token 刷新 Access Token
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body dto.RefreshTokenRequest true "Refresh Token"
// @Success 200 {object} response.Response{data=dto.LoginResponse}
// @Router /auth/refresh [post]
func (h *UserHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.userService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, result)
}

// Logout 登出
// @Summary 用户登出
// @Description 退出登录，使当前 Token 失效
// @Tags 认证
// @Produce json
// @Security Bearer
// @Success 200 {object} response.Response
// @Router /auth/logout [post]
func (h *UserHandler) Logout(c *gin.Context) {
	userID := authmiddleware.GetUserID(c)
	tokenID := "" // TODO: 从 JWT claims 中获取 jti

	err := h.userService.Logout(c.Request.Context(), userID, tokenID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OKWithoutData(c)
}

// ========== 当前用户相关接口 ==========

// GetCurrentUser 获取当前用户信息
// @Summary 获取当前用户信息
// @Description 获取当前登录用户的详细信息
// @Tags 用户
// @Produce json
// @Security Bearer
// @Success 200 {object} response.Response{data=dto.UserInfoResponse}
// @Router /user/me [get]
func (h *UserHandler) GetCurrentUser(c *gin.Context) {
	userID := authmiddleware.GetUserID(c)

	result, err := h.userService.GetCurrentUser(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, result)
}

// UpdateProfile 更新个人资料
// @Summary 更新个人资料
// @Description 更新当前用户的个人信息
// @Tags 用户
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body dto.UpdateProfileRequest true "个人信息"
// @Success 200 {object} response.Response
// @Router /user/profile [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := authmiddleware.GetUserID(c)

	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	err := h.userService.UpdateProfile(c.Request.Context(), userID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OKWithoutData(c)
}

// ChangePassword 修改密码
// @Summary 修改密码
// @Description 修改当前用户的密码
// @Tags 用户
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body dto.ChangePasswordRequest true "密码信息"
// @Success 200 {object} response.Response
// @Router /user/password [put]
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID := authmiddleware.GetUserID(c)

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	err := h.userService.ChangePassword(c.Request.Context(), userID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OKWithoutData(c)
}

// ========== 用户管理接口（管理员） ==========

// ListUser 用户列表
// @Summary 获取用户列表
// @Description 分页查询用户列表
// @Tags 用户管理
// @Produce json
// @Security Bearer
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param keyword query string false "关键词"
// @Param status query int false "状态"
// @Param department_id query string false "部门ID"
// @Success 200 {object} response.Response{data=response.PageResponse{list=[]dto.UserListResponse}}
// @Router /users [get]
func (h *UserHandler) ListUser(c *gin.Context) {
	var req dto.ListUserRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	list, total, err := h.userService.List(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Page(c, list, total, req.Page, req.PageSize)
}

// GetUser 获取用户详情
// @Summary 获取用户详情
// @Description 根据ID获取用户详细信息
// @Tags 用户管理
// @Produce json
// @Security Bearer
// @Param id path string true "用户ID"
// @Success 200 {object} response.Response{data=dto.UserInfoResponse}
// @Router /users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}

	result, err := h.userService.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, result)
}

// CreateUser 创建用户
// @Summary 创建用户
// @Description 管理员创建新用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body dto.CreateUserRequest true "用户信息"
// @Success 200 {object} response.Response{data=dto.UserInfoResponse}
// @Router /users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.userService.Create(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	c.JSON(http.StatusCreated, response.Response{
		Code:      response.CodeSuccess,
		Message:   "success",
		Data:      result,
		RequestID: c.GetString("request_id"),
		Timestamp: 0,
	})
}

// UpdateUser 更新用户
// @Summary 更新用户
// @Description 更新用户信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "用户ID"
// @Param request body dto.UpdateUserRequest true "用户信息"
// @Success 200 {object} response.Response{data=dto.UserInfoResponse}
// @Router /users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.userService.Update(c.Request.Context(), id, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, result)
}

// DeleteUser 删除用户
// @Summary 删除用户
// @Description 删除指定用户
// @Tags 用户管理
// @Produce json
// @Security Bearer
// @Param id path string true "用户ID"
// @Success 200 {object} response.Response
// @Router /users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}

	err = h.userService.Delete(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OKWithoutData(c)
}

// ========== 路由注册 ==========

// RegisterAuthRoutes 注册认证路由（公开）
func RegisterAuthRoutes(r *gin.RouterGroup, handler *UserHandler) {
	auth := r.Group("/auth")
	{
		auth.POST("/register", handler.Register)
		auth.POST("/login", handler.Login)
		auth.POST("/refresh", handler.RefreshToken)
		auth.POST("/logout", handler.Logout) // 注意：实际需要鉴权，这里只是分组
	}
}

// RegisterUserRoutes 注册用户路由（需要鉴权）
func RegisterUserRoutes(r *gin.RouterGroup, handler *UserHandler) {
	// 当前用户相关
	user := r.Group("/user")
	{
		user.GET("/me", handler.GetCurrentUser)
		user.PUT("/profile", handler.UpdateProfile)
		user.PUT("/password", handler.ChangePassword)
	}

	// 用户管理（管理员）
	users := r.Group("/users")
	{
		users.GET("", handler.ListUser)
		users.GET("/:id", handler.GetUser)
		users.POST("", handler.CreateUser)
		users.PUT("/:id", handler.UpdateUser)
		users.DELETE("/:id", handler.DeleteUser)
	}
}
