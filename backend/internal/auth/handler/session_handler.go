package handler

import (
	rbacService "github.com/Yogdunana/StarByte/backend/internal/rbac/service"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

func withPermission(group *gin.RouterGroup, permCode string, cacheService rbacService.PermissionCacheService) *gin.RouterGroup {
	g := group.Group("")
	g.Use(middleware.RequirePermission(permCode))
	g.Use(middleware.PermissionRequired(cacheService))
	return g
}

func registerSessionRoutes(
	authProtected *gin.RouterGroup,
	handler *AuthHandler,
	cacheService rbacService.PermissionCacheService,
) {
	if cacheService == nil {
		return
	}
	sessions := authProtected.Group("/sessions")
	withPermission(sessions, "session:read", cacheService).GET("", handler.ListSessions)
	withPermission(sessions, "session:read", cacheService).GET("/:user_id", handler.GetUserSessions)
	// /user/:user_id 必须在 /:token 之前，避免 "user" 被当成 token
	withPermission(sessions, "session:delete", cacheService).DELETE("/user/:user_id", handler.KickUserSessions)
	withPermission(sessions, "session:delete", cacheService).DELETE("/:token", handler.KickSession)
}

// ListSessions handles GET /api/v1/auth/sessions
// @Summary 在线会话列表
// @Description 管理员查看当前在线 Access Token 会话，可按关键词或用户筛选
// @Tags 认证
// @Produce json
// @Param keyword query string false "关键词"
// @Param user_id query string false "用户 ID"
// @Success 200 {object} response.Response{data=dto.SessionListResponse}
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /auth/sessions [get]
// @Security BearerAuth
func (h *AuthHandler) ListSessions(c *gin.Context) {
	result, err := h.authService.ListSessions(c.Request.Context(), c.Query("keyword"), c.Query("user_id"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

// GetUserSessions handles GET /api/v1/auth/sessions/:user_id
// @Summary 指定用户的在线会话
// @Description 查看某一用户当前全部在线会话
// @Tags 认证
// @Produce json
// @Param user_id path string true "用户 ID"
// @Success 200 {object} response.Response{data=dto.UserSessionsResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /auth/sessions/{user_id} [get]
// @Security BearerAuth
func (h *AuthHandler) GetUserSessions(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		response.BadRequest(c, "缺少用户 ID")
		return
	}
	result, err := h.authService.GetUserSessions(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

// KickSession handles DELETE /api/v1/auth/sessions/:token
// @Summary 踢出单个会话
// @Description 按 token 标识强制下线一个会话
// @Tags 认证
// @Produce json
// @Param token path string true "会话 token 标识"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /auth/sessions/{token} [delete]
// @Security BearerAuth
func (h *AuthHandler) KickSession(c *gin.Context) {
	tokenID := c.Param("token")
	if tokenID == "" || tokenID == "user" {
		response.BadRequest(c, "缺少会话标识")
		return
	}
	if err := h.authService.KickSession(c.Request.Context(), tokenID); err != nil {
		response.Error(c, err)
		return
	}
	response.OKWithoutData(c)
}

// KickUserSessions handles DELETE /api/v1/auth/sessions/user/:user_id
// @Summary 踢出用户全部会话
// @Description 强制下线指定用户的全部在线会话
// @Tags 认证
// @Produce json
// @Param user_id path string true "用户 ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /auth/sessions/user/{user_id} [delete]
// @Security BearerAuth
func (h *AuthHandler) KickUserSessions(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		response.BadRequest(c, "缺少用户 ID")
		return
	}
	if err := h.authService.KickUserSessions(c.Request.Context(), userID); err != nil {
		response.Error(c, err)
		return
	}
	response.OKWithoutData(c)
}
