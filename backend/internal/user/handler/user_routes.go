package handler

import (
	rbacService "github.com/Yogdunana/StarByte/backend/internal/rbac/service"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterUserRoutes 注册用户路由（需要鉴权）
func RegisterUserRoutes(r *gin.RouterGroup, handler *UserHandler, cache rbacService.PermissionCacheService) {
	user := r.Group("/user")
	{
		user.GET("/me", handler.GetCurrentUser)
		user.PUT("/profile", handler.UpdateProfile)
		user.PUT("/password", handler.ChangePassword)
	}

	users := r.Group("/users")
	{
		users.GET("", middleware.RequirePermission("user:read"), middleware.PermissionRequired(cache), handler.ListUser)
		users.GET("/:id", middleware.RequirePermission("user:read"), middleware.PermissionRequired(cache), handler.GetUser)
		users.POST("", middleware.RequirePermission("user:create"), middleware.PermissionRequired(cache), handler.CreateUser)
		users.PUT("/:id", middleware.RequirePermission("user:update"), middleware.PermissionRequired(cache), handler.UpdateUser)
		users.DELETE("/:id", middleware.RequirePermission("user:delete"), middleware.PermissionRequired(cache), handler.DeleteUser)
	}
}
