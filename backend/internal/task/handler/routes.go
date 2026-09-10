package handler

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	rbacRepo "github.com/Yogdunana/StarByte/backend/internal/rbac/repo"
	rbacService "github.com/Yogdunana/StarByte/backend/internal/rbac/service"
	"github.com/Yogdunana/StarByte/backend/internal/task/service"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware"
)

type TaskHandler struct {
	svc service.TaskService
}

func NewTaskHandler(svc service.TaskService) *TaskHandler {
	return &TaskHandler{svc: svc}
}

func withPermission(group *gin.RouterGroup, permCode string, cacheService rbacService.PermissionCacheService, db *gorm.DB, deptRepo rbacRepo.DepartmentRepo) *gin.RouterGroup {
	g := group.Group("")
	g.Use(middleware.RequirePermission(permCode))
	g.Use(middleware.PermissionRequired(cacheService))
	g.Use(middleware.RequireDataScope(permCode), middleware.DataScopeMiddleware(db, deptRepo, cacheService), taskViewer())
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
	g.Use(middleware.RequireDataScope("task:read"))
	g.Use(middleware.PermissionRequired(cacheService))
	g.Use(middleware.DataScopeMiddleware(db, deptRepo, cacheService), taskViewer())
	for _, permission := range []string{"task:update", "task:delete", "task:assign", "task:transfer", "task:comment", "task:create"} {
		g.Use(middleware.RequireDataScope(permission), middleware.DataScopeMiddleware(db, deptRepo, cacheService), taskCapability(permission))
	}
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
	read := withReadScope(g, cacheService, db, deptRepo)
	personal := g.Group("", personalTaskViewer(cacheService))
	for _, p := range []string{"task:update", "task:delete", "task:assign", "task:transfer", "task:comment", "task:create"} {
		personal.Use(middleware.RequireDataScope(p), middleware.DataScopeMiddleware(db, deptRepo, cacheService), taskCapability(p))
	}
	personal.GET("/:id/workflow", h.GetWorkflow)
	personal.POST("/:id/workflow/actions", h.ActWorkflow)
	personal.GET("/my/todo", h.MyTodo)
	personal.GET("/my/done", h.MyDone)
	personal.GET("/my/created", h.MyCreated)
	personal.GET("/my/overdue", h.MyOverdue)

	read.GET("/stats", h.Stats)
	read.GET("", h.ListTasks)
	read.GET("/:id", h.GetTask)
	read.GET("/:id/logs", h.ListLogs)
	read.GET("/:id/comments", h.ListComments)
	read.GET("/:id/attachments", h.ListAttachments)
	read.GET("/:id/attachments/:aid", h.DownloadAttachment)

	create := withPermission(g, "task:create", cacheService, db, deptRepo)
	create.GET("/create-candidates", h.Candidates)
	create.GET("/assignment-roles", h.AssignmentRoles)
	create.POST("", h.CreateTask)
	create.POST("/:id/urge", h.Urge)

	update := withPermission(g, "task:update", cacheService, db, deptRepo)
	update.PUT("/:id", h.UpdateTask)
	update.POST("/:id/status", h.ChangeStatus)
	update.POST("/:id/attachments", h.UploadAttachment)
	update.DELETE("/:id/attachments/:aid", h.DeleteAttachment)

	del := withPermission(g, "task:delete", cacheService, db, deptRepo)
	del.DELETE("/:id", h.DeleteTask)

	assign := withPermission(g, "task:assign", cacheService, db, deptRepo)
	assign.GET("/assign-candidates", h.Candidates)
	assign.POST("/:id/assign", h.Assign)

	transfer := withPermission(g, "task:transfer", cacheService, db, deptRepo)
	transfer.GET("/transfer-candidates", h.Candidates)
	transfer.POST("/:id/transfer", h.Transfer)

	comment := withPermission(g, "task:comment", cacheService, db, deptRepo)
	comment.POST("/:id/comments", h.AddComment)
	comment.PUT("/:id/comments/:cid", h.UpdateComment)
	comment.DELETE("/:id/comments/:cid", h.DeleteComment)
}
