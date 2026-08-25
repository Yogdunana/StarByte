package handler

import (
	rbacRepo "github.com/Yogdunana/StarByte/backend/internal/rbac/repo"
	rbacService "github.com/Yogdunana/StarByte/backend/internal/rbac/service"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

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
		roles.Use(middleware.RequirePermission("role:read"))
		roles.Use(middleware.PermissionRequired(cacheService))
		{
			roles.GET("", roleHandler.List)
			roles.GET("/:id", roleHandler.GetByID)
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
		rolesCreate := system.Group("/roles")
		rolesCreate.Use(middleware.RequirePermission("role:create"))
		rolesCreate.Use(middleware.PermissionRequired(cacheService))
		{
			rolesCreate.POST("", roleHandler.Create)
		}
		rolesUpdate := system.Group("/roles")
		rolesUpdate.Use(middleware.RequirePermission("role:update"))
		rolesUpdate.Use(middleware.PermissionRequired(cacheService))
		{
			rolesUpdate.PUT("/:id", roleHandler.Update)
		}
		rolesDelete := system.Group("/roles")
		rolesDelete.Use(middleware.RequirePermission("role:delete"))
		rolesDelete.Use(middleware.PermissionRequired(cacheService))
		{
			rolesDelete.DELETE("/:id", roleHandler.Delete)
		}
		rolesAssign := system.Group("/roles")
		rolesAssign.Use(middleware.RequirePermission("role:assign"))
		rolesAssign.Use(middleware.PermissionRequired(cacheService))
		{
			rolesAssign.PUT("/:id/permissions", roleHandler.AssignPermissions)
		}

		// ========== 权限 ==========
		perms := system.Group("/permissions")
		perms.Use(middleware.RequirePermission("permission:read"))
		perms.Use(middleware.PermissionRequired(cacheService))
		{
			perms.GET("", permHandler.GetTree)
			perms.GET("/:id", permHandler.GetByID)
		}
		permsCreate := system.Group("/permissions")
		permsCreate.Use(middleware.RequirePermission("permission:create"))
		permsCreate.Use(middleware.PermissionRequired(cacheService))
		{
			permsCreate.POST("", permHandler.Create)
		}
		permsUpdate := system.Group("/permissions")
		permsUpdate.Use(middleware.RequirePermission("permission:update"))
		permsUpdate.Use(middleware.PermissionRequired(cacheService))
		{
			permsUpdate.PUT("/:id", permHandler.Update)
		}
		permsDelete := system.Group("/permissions")
		permsDelete.Use(middleware.RequirePermission("permission:delete"))
		permsDelete.Use(middleware.PermissionRequired(cacheService))
		{
			permsDelete.DELETE("/:id", permHandler.Delete)
		}

		// ========== 部门 ==========
		depts := system.Group("/departments")
		depts.Use(middleware.RequirePermission("department:read"))
		depts.Use(middleware.PermissionRequired(cacheService))
		{
			depts.GET("", deptHandler.GetTree)
			depts.GET("/:id", deptHandler.GetByID)
		}
		deptsCreate := system.Group("/departments")
		deptsCreate.Use(middleware.RequirePermission("department:create"))
		deptsCreate.Use(middleware.PermissionRequired(cacheService))
		{
			deptsCreate.POST("", deptHandler.Create)
		}
		deptsUpdate := system.Group("/departments")
		deptsUpdate.Use(middleware.RequirePermission("department:update"))
		deptsUpdate.Use(middleware.PermissionRequired(cacheService))
		{
			deptsUpdate.PUT("/:id", deptHandler.Update)
		}
		deptsDelete := system.Group("/departments")
		deptsDelete.Use(middleware.RequirePermission("department:delete"))
		deptsDelete.Use(middleware.PermissionRequired(cacheService))
		{
			deptsDelete.DELETE("/:id", deptHandler.Delete)
		}

		// ========== 职位 ==========
		positions := system.Group("/positions")
		positions.Use(middleware.RequirePermission("position:read"))
		positions.Use(middleware.PermissionRequired(cacheService))
		{
			positions.GET("", posHandler.List)
			positions.GET("/:id", posHandler.GetByID)
		}
		positionsCreate := system.Group("/positions")
		positionsCreate.Use(middleware.RequirePermission("position:create"))
		positionsCreate.Use(middleware.PermissionRequired(cacheService))
		{
			positionsCreate.POST("", posHandler.Create)
		}
		positionsUpdate := system.Group("/positions")
		positionsUpdate.Use(middleware.RequirePermission("position:update"))
		positionsUpdate.Use(middleware.PermissionRequired(cacheService))
		{
			positionsUpdate.PUT("/:id", posHandler.Update)
		}
		positionsDelete := system.Group("/positions")
		positionsDelete.Use(middleware.RequirePermission("position:delete"))
		positionsDelete.Use(middleware.PermissionRequired(cacheService))
		{
			positionsDelete.DELETE("/:id", posHandler.Delete)
		}
	}
}
