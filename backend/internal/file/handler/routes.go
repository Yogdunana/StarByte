package handler

import (
	rbacService "github.com/Yogdunana/StarByte/backend/internal/rbac/service"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware"
	"github.com/gin-gonic/gin"
)

func withPermission(group *gin.RouterGroup, permCode string, cacheService rbacService.PermissionCacheService) *gin.RouterGroup {
	g := group.Group("")
	g.Use(middleware.RequirePermission(permCode))
	g.Use(middleware.PermissionRequired(cacheService))
	return g
}

// RegisterRoutes 注册 /api/v1/files
func RegisterRoutes(
	r *gin.RouterGroup,
	fileHandler *FileHandler,
	cacheService rbacService.PermissionCacheService,
) {
	files := r.Group("/files")

	create := withPermission(files, "file:create", cacheService)
	create.POST("/upload", fileHandler.Upload)
	create.POST("/upload-batch", fileHandler.UploadBatch)

	read := withPermission(files, "file:read", cacheService)
	read.GET("", fileHandler.List)
	read.GET("/:id", fileHandler.GetByID)
	read.GET("/:id/download", fileHandler.Download)

	// 删除：登录即可进入，服务层校验上传者或 file:delete
	files.DELETE("/:id", fileHandler.Delete)
}
