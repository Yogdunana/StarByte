package handler

import "github.com/gin-gonic/gin"

// RegisterUserRoutes 注册用户路由（需要鉴权）
func RegisterUserRoutes(r *gin.RouterGroup, handler *UserHandler) {
	user := r.Group("/user")
	{
		user.GET("/me", handler.GetCurrentUser)
		user.PUT("/profile", handler.UpdateProfile)
		user.PUT("/password", handler.ChangePassword)
	}

	users := r.Group("/users")
	{
		users.GET("", handler.ListUser)
		users.GET("/:id", handler.GetUser)
		users.POST("", handler.CreateUser)
		users.PUT("/:id", handler.UpdateUser)
		users.DELETE("/:id", handler.DeleteUser)
	}
}
