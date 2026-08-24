package handler

import (
	"github.com/Yogdunana/StarByte/backend/pkg/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册 RBAC 系统管理路由。每个资源按操作类型分组，并通过
// middleware.RequirePermission 声明所需权限码，实际的权限校验由上层路由组
// 注册的 middleware.PermissionRequired 中间件统一执行。
func RegisterRoutes(
	r *gin.RouterGroup,
	roleHandler *RoleHandler,
	permHandler *PermissionHandler,
	deptHandler *DepartmentHandler,
	posHandler *PositionHandler,
) {
	system := r.Group("/system")
	{
		// ========== 角色 ==========
		roles := system.Group("/roles")
		roles.Use(middleware.RequirePermission("role:read"))
		{
			roles.GET("", roleHandler.List)
			roles.GET("/:id", roleHandler.GetByID)
			roles.GET("/:id/users", roleHandler.GetRoleUsers)
		}
		rolesCreate := system.Group("/roles")
		rolesCreate.Use(middleware.RequirePermission("role:create"))
		{
			rolesCreate.POST("", roleHandler.Create)
		}
		rolesUpdate := system.Group("/roles")
		rolesUpdate.Use(middleware.RequirePermission("role:update"))
		{
			rolesUpdate.PUT("/:id", roleHandler.Update)
		}
		rolesDelete := system.Group("/roles")
		rolesDelete.Use(middleware.RequirePermission("role:delete"))
		{
			rolesDelete.DELETE("/:id", roleHandler.Delete)
		}
		rolesAssign := system.Group("/roles")
		rolesAssign.Use(middleware.RequirePermission("role:assign"))
		{
			rolesAssign.PUT("/:id/permissions", roleHandler.AssignPermissions)
		}

		// ========== 权限 ==========
		perms := system.Group("/permissions")
		perms.Use(middleware.RequirePermission("permission:read"))
		{
			perms.GET("", permHandler.GetTree)
			perms.GET("/:id", permHandler.GetByID)
		}
		permsCreate := system.Group("/permissions")
		permsCreate.Use(middleware.RequirePermission("permission:create"))
		{
			permsCreate.POST("", permHandler.Create)
		}
		permsUpdate := system.Group("/permissions")
		permsUpdate.Use(middleware.RequirePermission("permission:update"))
		{
			permsUpdate.PUT("/:id", permHandler.Update)
		}
		permsDelete := system.Group("/permissions")
		permsDelete.Use(middleware.RequirePermission("permission:delete"))
		{
			permsDelete.DELETE("/:id", permHandler.Delete)
		}

		// ========== 部门 ==========
		depts := system.Group("/departments")
		depts.Use(middleware.RequirePermission("dept:read"))
		{
			depts.GET("", deptHandler.GetTree)
			depts.GET("/:id", deptHandler.GetByID)
		}
		deptsCreate := system.Group("/departments")
		deptsCreate.Use(middleware.RequirePermission("dept:create"))
		{
			deptsCreate.POST("", deptHandler.Create)
		}
		deptsUpdate := system.Group("/departments")
		deptsUpdate.Use(middleware.RequirePermission("dept:update"))
		{
			deptsUpdate.PUT("/:id", deptHandler.Update)
		}
		deptsDelete := system.Group("/departments")
		deptsDelete.Use(middleware.RequirePermission("dept:delete"))
		{
			deptsDelete.DELETE("/:id", deptHandler.Delete)
		}

		// ========== 职位 ==========
		positions := system.Group("/positions")
		positions.Use(middleware.RequirePermission("position:read"))
		{
			positions.GET("", posHandler.List)
			positions.GET("/:id", posHandler.GetByID)
		}
		positionsCreate := system.Group("/positions")
		positionsCreate.Use(middleware.RequirePermission("position:create"))
		{
			positionsCreate.POST("", posHandler.Create)
		}
		positionsUpdate := system.Group("/positions")
		positionsUpdate.Use(middleware.RequirePermission("position:update"))
		{
			positionsUpdate.PUT("/:id", posHandler.Update)
		}
		positionsDelete := system.Group("/positions")
		positionsDelete.Use(middleware.RequirePermission("position:delete"))
		{
			positionsDelete.DELETE("/:id", posHandler.Delete)
		}
	}
}
