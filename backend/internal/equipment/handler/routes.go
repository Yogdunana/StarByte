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

// RegisterRoutes 注册 /api/v1/equipment 路由
func RegisterRoutes(
	r *gin.RouterGroup,
	h *EquipmentHandler,
	cacheService rbacService.PermissionCacheService,
) {
	equip := r.Group("/equipment")

	// 读取权限
	read := withPermission(equip, "equipment:read", cacheService)
	read.GET("", h.ListEquipment)
	read.GET("/:id/records", h.ListBorrows)

	// 入库
	create := withPermission(equip, "equipment:create", cacheService)
	create.POST("", h.CreateEquipment)

	// 更新（归还确认也用 update 权限）
	update := withPermission(equip, "equipment:update", cacheService)
	update.PUT("/:id", h.UpdateEquipment)
	update.PUT("/borrow/:id/return", h.ReturnBorrow)
	update.PUT("/borrow/:id/approve", h.ActionBorrow)
	update.POST("/maintenance", h.CreateMaintenance)
	update.GET("/maintenance", h.ListMaintenance)
	update.POST("/inventory", h.CreateInventory)
	update.GET("/inventory", h.ListInventory)

	// 删除
	del := withPermission(equip, "equipment:delete", cacheService)
	del.DELETE("/:id", h.DeleteEquipment)

	// 借用申请（登录即可）
	equip.POST("/borrow", h.CreateBorrow)
	// 借用记录列表（登录即可查自己的）
	equip.GET("/borrow", h.ListBorrows)
}
