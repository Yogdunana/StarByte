package handler

import (
	rbacRepo "github.com/Yogdunana/StarByte/backend/internal/rbac/repo"
	rbacService "github.com/Yogdunana/StarByte/backend/internal/rbac/service"
	"github.com/Yogdunana/StarByte/backend/internal/task/service"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TaskHandler struct {
	svc service.TaskService
}

func NewTaskHandler(svc service.TaskService) *TaskHandler {
	return &TaskHandler{svc: svc}
}

func withPermission(group *gin.RouterGroup, permCode string, cacheService rbacService.PermissionCacheService) *gin.RouterGroup {
	g := group.Group("")
	g.Use(middleware.RequirePermission(permCode))
	g.Use(middleware.PermissionRequired(cacheService))
	return g
}

func withReadScope(
	group *gin.RouterGroup,
	cacheService rbacService.PermissionCacheService,
	db *gorm.DB,
	deptRepo rbacRepo.DepartmentRepo,
) *gin.RouterGroup {
	g := group.Group("")
	g.Use(middleware.RequirePermission("task:read"))
	g.Use(middleware.RequireDataScope("task"))
	g.Use(middleware.PermissionRequired(cacheService))
	g.Use(middleware.DataScopeMiddleware(db, deptRepo, cacheService))
	return g
}

// RegisterRoutes 注册 /api/v1/tasks。静态路径必须在 /:id 之前。
func RegisterRoutes(
	r *gin.RouterGroup,
	h *TaskHandler,
	cacheService rbacService.PermissionCacheService,
	db *gorm.DB,
	deptRepo rbacRepo.DepartmentRepo,
) {
	g := r.Group("/tasks")
	g.GET("/my/todo", h.MyTodo)
	g.GET("/my/done", h.MyDone)
	g.GET("/my/created", h.MyCreated)
	g.GET("/my/overdue", h.MyOverdue)

	read := withReadScope(g, cacheService, db, deptRepo)
	read.GET("/stats", h.Stats)
	read.GET("", h.ListTasks)
	read.GET("/:id", h.GetTask)
	read.GET("/:id/logs", h.ListLogs)
	read.GET("/:id/comments", h.ListComments)
	read.GET("/:id/attachments", h.ListAttachments)
	read.GET("/:id/attachments/:aid", h.DownloadAttachment)

	create := withPermission(g, "task:create", cacheService)
	create.POST("", h.CreateTask)
	create.POST("/:id/urge", h.Urge)

	update := withPermission(g, "task:update", cacheService)
	update.PUT("/:id", h.UpdateTask)
	update.POST("/:id/status", h.ChangeStatus)
	update.POST("/:id/attachments", h.UploadAttachment)
	update.DELETE("/:id/attachments/:aid", h.DeleteAttachment)

	del := withPermission(g, "task:delete", cacheService)
	del.DELETE("/:id", h.DeleteTask)

	assign := withPermission(g, "task:assign", cacheService)
	assign.POST("/:id/assign", h.Assign)

	transfer := withPermission(g, "task:transfer", cacheService)
	transfer.POST("/:id/transfer", h.Transfer)

	comment := withPermission(g, "task:comment", cacheService)
	comment.POST("/:id/comments", h.AddComment)
	comment.PUT("/:id/comments/:cid", h.UpdateComment)
	comment.DELETE("/:id/comments/:cid", h.DeleteComment)
}
