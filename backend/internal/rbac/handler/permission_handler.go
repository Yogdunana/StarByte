package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/rbac/dto"
	"github.com/Yogdunana/StarByte/backend/internal/rbac/service"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// PermissionHandler 权限处理器
type PermissionHandler struct {
	permissionService service.PermissionService
}

// NewPermissionHandler 创建权限处理器
func NewPermissionHandler(permissionService service.PermissionService) *PermissionHandler {
	return &PermissionHandler{permissionService: permissionService}
}

// GetTree 权限树
// @Summary 获取权限树
// @Description 获取完整的权限树形结构
// @Tags 权限管理
// @Produce json
// @Security Bearer
// @Success 200 {object} response.Response{data=[]dto.PermissionTreeResponse}
// @Router /system/permissions [get]
func (h *PermissionHandler) GetTree(c *gin.Context) {
	tree, err := h.permissionService.GetTree(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, tree)
}

// GetByID 获取权限详情
// @Summary 获取权限详情
// @Description 根据ID获取权限详细信息
// @Tags 权限管理
// @Produce json
// @Security Bearer
// @Param id path string true "权限ID"
// @Success 200 {object} response.Response{data=dto.PermissionResponse}
// @Router /system/permissions/{id} [get]
func (h *PermissionHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "无效的权限ID")
		return
	}

	result, err := h.permissionService.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, result)
}

// Create 创建权限
// @Summary 创建权限
// @Description 创建新权限
// @Tags 权限管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body dto.CreatePermissionRequest true "权限信息"
// @Success 200 {object} response.Response{data=dto.PermissionResponse}
// @Router /system/permissions [post]
func (h *PermissionHandler) Create(c *gin.Context) {
	var req dto.CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	result, err := h.permissionService.Create(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, result)
}

// Update 更新权限
// @Summary 更新权限
// @Description 更新权限信息
// @Tags 权限管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "权限ID"
// @Param request body dto.UpdatePermissionRequest true "权限信息"
// @Success 200 {object} response.Response{data=dto.PermissionResponse}
// @Router /system/permissions/{id} [put]
func (h *PermissionHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "无效的权限ID")
		return
	}

	var req dto.UpdatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	result, err := h.permissionService.Update(c.Request.Context(), id, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, result)
}

// Delete 删除权限
// @Summary 删除权限
// @Description 删除指定权限
// @Tags 权限管理
// @Produce json
// @Security Bearer
// @Param id path string true "权限ID"
// @Success 200 {object} response.Response
// @Router /system/permissions/{id} [delete]
func (h *PermissionHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "无效的权限ID")
		return
	}

	if err := h.permissionService.Delete(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}

	response.OKWithoutData(c)
}
