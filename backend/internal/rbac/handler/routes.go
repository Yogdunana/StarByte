package handler

import (
	rbacRepo "github.com/Yogdunana/StarByte/backend/internal/rbac/repo"
	rbacService "github.com/Yogdunana/StarByte/backend/internal/rbac/service"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// withPermission 创建一个带有权限校验的路由组
// 等效于依次注册 RequirePermission 和 PermissionRequired 中间件
func withPermission(group *gin.RouterGroup, permCode string, cacheService rbacService.PermissionCacheService) *gin.RouterGroup {
	g := group.Group("")
	g.Use(middleware.RequirePermission(permCode))
	g.Use(middleware.PermissionRequired(cacheService))
	return g
}

// RegisterRoutes 注册 RBAC 系统管理路由。
//
// 权限校验中间件（PermissionRequired）在每个子路由组内注册，
// 位于 RequirePermission 之后，确保执行顺序正确：
// 1. RequirePermission 设置权限码 → 2. PermissionRequired 读取并校验
//
// 数据权限中间件（DataScopeMiddleware）仅在查询用户数据的接口上注册，
// 因为角色/权限/部门/职位等系统管理数据为全局可见，无需数据权限过滤。
// 数据权限执行顺序：RequireDataScope → PermissionRequired → DataScopeMiddleware
func RegisterRoutes(
	r *gin.RouterGroup,
	db *gorm.DB,
	roleHandler *RoleHandler,
	permHandler *PermissionHandler,
	deptHandler *DepartmentHandler,
	posHandler *PositionHandler,
	cacheService rbacService.PermissionCacheService,
	deptRepo rbacRepo.DepartmentRepo,
) {
	system := r.Group("/system")
	{
		// ========== 角色 ==========
		roles := system.Group("/roles")
		{
			withPermission(roles, "role:read", cacheService).GET("", roleHandler.List)
			withPermission(roles, "role:read", cacheService).GET("/:id", roleHandler.GetByID)
			withPermission(roles, "role:create", cacheService).POST("", roleHandler.Create)
			withPermission(roles, "role:update", cacheService).PUT("/:id", roleHandler.Update)
			withPermission(roles, "role:delete", cacheService).DELETE("/:id", roleHandler.Delete)
			withPermission(roles, "role:assign", cacheService).PUT("/:id/permissions", roleHandler.AssignPermissions)
		}
		// 角色用户列表：查询用户数据，需应用数据权限
		roleUsers := system.Group("/roles")
		roleUsers.Use(middleware.RequirePermission("role:read"))
		roleUsers.Use(middleware.RequireDataScope("user"))
		roleUsers.Use(middleware.PermissionRequired(cacheService))
		roleUsers.Use(middleware.DataScopeMiddleware(db, deptRepo, cacheService))
		{
			roleUsers.GET("/:id/users", roleHandler.GetRoleUsers)
		}

		// ========== 权限 ==========
		perms := system.Group("/permissions")
		{
			withPermission(perms, "permission:read", cacheService).GET("", permHandler.GetTree)
			withPermission(perms, "permission:read", cacheService).GET("/:id", permHandler.GetByID)
			withPermission(perms, "permission:create", cacheService).POST("", permHandler.Create)
			withPermission(perms, "permission:update", cacheService).PUT("/:id", permHandler.Update)
			withPermission(perms, "permission:delete", cacheService).DELETE("/:id", permHandler.Delete)
		}

		// ========== 部门 ==========
		depts := system.Group("/departments")
		{
			withPermission(depts, "department:read", cacheService).GET("", deptHandler.GetTree)
			withPermission(depts, "department:read", cacheService).GET("/:id", deptHandler.GetByID)
			withPermission(depts, "department:create", cacheService).POST("", deptHandler.Create)
			withPermission(depts, "department:update", cacheService).PUT("/:id", deptHandler.Update)
			withPermission(depts, "department:delete", cacheService).DELETE("/:id", deptHandler.Delete)
		}

		// ========== 职位 ==========
		positions := system.Group("/positions")
		{
			withPermission(positions, "position:read", cacheService).GET("", posHandler.List)
			withPermission(positions, "position:read", cacheService).GET("/:id", posHandler.GetByID)
			withPermission(positions, "position:create", cacheService).POST("", posHandler.Create)
			withPermission(positions, "position:update", cacheService).PUT("/:id", posHandler.Update)
			withPermission(positions, "position:delete", cacheService).DELETE("/:id", posHandler.Delete)
		}
	}
}
