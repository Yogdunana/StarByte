package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/auth/dto"
	"github.com/Yogdunana/StarByte/backend/internal/auth/service"
	authmiddleware "github.com/Yogdunana/StarByte/backend/pkg/middleware/auth"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication-related HTTP requests.
type AuthHandler struct {
	authService service.AuthService
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Login handles POST /api/v1/auth/login
// @Summary 用户登录
// @Description 用户名密码登录，返回 Access Token 和 Refresh Token
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "登录信息"
// @Success 200 {object} response.Response{data=dto.LoginResponse}
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	ip := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	result, err := h.authService.Login(c.Request.Context(), &req, ip, userAgent)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, result)
}

// RefreshToken handles POST /api/v1/auth/refresh
// @Summary 刷新 Token
// @Description 使用 Refresh Token 刷新 Access Token（旋转机制：旧 Refresh Token 失效）
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body dto.RefreshTokenRequest true "Refresh Token"
// @Success 200 {object} response.Response{data=dto.RefreshResponse}
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.authService.RefreshToken(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, result)
}

// Logout handles POST /api/v1/auth/logout
// @Summary 用户登出
// @Description 退出登录，使当前 Access Token 加入黑名单，Refresh Token 失效
// @Tags 认证
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body dto.LogoutRequest false "登出信息（refresh_token 可选）"
// @Success 200 {object} response.Response
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	userID := authmiddleware.GetUserID(c)
	tokenID := authmiddleware.GetTokenID(c)

	var req dto.LogoutRequest
	_ = c.ShouldBindJSON(&req) // best-effort: body is optional

	err := h.authService.Logout(c.Request.Context(), userID, tokenID, req.RefreshToken)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OKWithoutData(c)
}

// GetCurrentUser handles GET /api/v1/auth/me
// @Summary 获取当前用户信息
// @Description 获取当前登录用户的详细信息（含角色和权限）
// @Tags 认证
// @Produce json
// @Security Bearer
// @Success 200 {object} response.Response{data=dto.UserInfo}
// @Router /auth/me [get]
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID := authmiddleware.GetUserID(c)

	result, err := h.authService.GetCurrentUser(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, result)
}

// ChangePassword handles PUT /api/v1/auth/password
// @Summary 修改密码
// @Description 修改当前用户的密码（需校验原密码，新密码需满足强度要求）
// @Tags 认证
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body dto.ChangePasswordRequest true "密码信息"
// @Success 200 {object} response.Response
// @Router /auth/password [put]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID := authmiddleware.GetUserID(c)

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	err := h.authService.ChangePassword(c.Request.Context(), userID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OKWithoutData(c)
}

// WechatQRCode handles POST /api/v1/auth/wechat/qrcode
// @Summary 获取微信登录二维码（预留）
// @Description 返回微信扫码登录二维码链接（接口预留，一期不实现）
// @Tags 认证
// @Produce json
// @Success 200 {object} response.Response
// @Router /auth/wechat/qrcode [post]
func (h *AuthHandler) WechatQRCode(c *gin.Context) {
	response.NotImplemented(c, "微信扫码登录功能暂未开通")
}

// WechatCallback handles POST /api/v1/auth/wechat/callback
// @Summary 微信登录回调（预留）
// @Description 微信扫码登录回调接口（接口预留，一期不实现）
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body dto.WechatLoginRequest true "微信授权码"
// @Success 200 {object} response.Response
// @Router /auth/wechat/callback [post]
func (h *AuthHandler) WechatCallback(c *gin.Context) {
	response.NotImplemented(c, "微信登录回调功能暂未开通")
}

// OAuthLogin handles POST /api/v1/auth/oauth/{provider}
// @Summary 第三方 OAuth 登录（预留）
// @Description 通用第三方 OAuth 登录接口（接口预留，一期不实现）
// @Tags 认证
// @Accept json
// @Produce json
// @Param provider path string true "OAuth 提供者（github/google 等）"
// @Param request body dto.OAuthLoginRequest true "OAuth 授权码"
// @Success 200 {object} response.Response
// @Router /auth/oauth/{provider} [post]
func (h *AuthHandler) OAuthLogin(c *gin.Context) {
	provider := c.Param("provider")
	if provider == "" {
		response.NotImplemented(c, "第三方 OAuth 登录功能暂未开通")
		return
	}
	response.NotImplemented(c, "第三方 OAuth 登录功能暂未开通: "+provider)
}

// RegisterRoutes registers all authentication routes.
// public routes (no auth): login, refresh, wechat, oauth
// protected routes (auth): logout, me, password
// loginRateLimiter is an optional middleware applied to the login endpoint
// for brute-force protection (e.g., 5 req/min). Pass nil to skip.
func RegisterRoutes(
	public *gin.RouterGroup,
	protected *gin.RouterGroup,
	handler *AuthHandler,
	loginRateLimiter gin.HandlerFunc,
) {
	if public != nil {
		authGroup := public.Group("/auth")
		{
			if loginRateLimiter != nil {
				authGroup.POST("/login", loginRateLimiter, handler.Login)
			} else {
				authGroup.POST("/login", handler.Login)
			}
			authGroup.POST("/refresh", handler.RefreshToken)
			authGroup.POST("/wechat/qrcode", handler.WechatQRCode)
			authGroup.POST("/wechat/callback", handler.WechatCallback)
			authGroup.POST("/oauth/:provider", handler.OAuthLogin)
		}
	}

	if protected != nil {
		authProtected := protected.Group("/auth")
		{
			authProtected.POST("/logout", handler.Logout)
			authProtected.GET("/me", handler.GetCurrentUser)
			authProtected.PUT("/password", handler.ChangePassword)
		}
	}
}
