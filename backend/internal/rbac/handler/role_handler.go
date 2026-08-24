package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/rbac/dto"
	"github.com/Yogdunana/StarByte/backend/internal/rbac/service"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RoleHandler 角色处理器
type RoleHandler struct {
	roleService service.RoleService
}

// NewRoleHandler 创建角色处理器
func NewRoleHandler(roleService service.RoleService) *RoleHandler {
	return &RoleHandler{roleService: roleService}
}

// List 角色列表
// @Summary 获取角色列表
// @Description 分页查询角色列表
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param keyword query string false "关键词"
// @Success 200 {object} response.Response{data=response.PageResponse{list=[]dto.RoleListResponse}}
// @Router /system/roles [get]
func (h *RoleHandler) List(c *gin.Context) {
	var req dto.ListRoleRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	list, total, err := h.roleService.List(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Page(c, list, total, req.Page, req.PageSize)
}

// GetByID 获取角色详情
// @Summary 获取角色详情
// @Description 根据ID获取角色详细信息（含权限列表）
// @Tags 角色管理
// @Produce json
// @Security Bearer
// @Param id path string true "角色ID"
// @Success 200 {object} response.Response{data=dto.RoleDetailResponse}
// @Router /system/roles/{id} [get]
func (h *RoleHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "无效的角色ID")
		return
	}

	result, err := h.roleService.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, result)
}

// Create 创建角色
// @Summary 创建角色
// @Description 创建新角色
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body dto.CreateRoleRequest true "角色信息"
// @Success 200 {object} response.Response{data=dto.RoleResponse}
// @Router /system/roles [post]
func (h *RoleHandler) Create(c *gin.Context) {
	var req dto.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.roleService.Create(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, result)
}

// Update 更新角色
// @Summary 更新角色
// @Description 更新角色信息
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "角色ID"
// @Param request body dto.UpdateRoleRequest true "角色信息"
// @Success 200 {object} response.Response{data=dto.RoleResponse}
// @Router /system/roles/{id} [put]
func (h *RoleHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "无效的角色ID")
		return
	}

	var req dto.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.roleService.Update(c.Request.Context(), id, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, result)
}

// Delete 删除角色
// @Summary 删除角色
// @Description 删除指定角色
// @Tags 角色管理
// @Produce json
// @Security Bearer
// @Param id path string true "角色ID"
// @Success 200 {object} response.Response
// @Router /system/roles/{id} [delete]
func (h *RoleHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "无效的角色ID")
		return
	}

	if err := h.roleService.Delete(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}

	response.OKWithoutData(c)
}

// AssignPermissions 分配角色权限
// @Summary 分配角色权限
// @Description 为指定角色分配权限集合
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "角色ID"
// @Param request body dto.AssignPermissionsRequest true "权限ID列表"
// @Success 200 {object} response.Response
// @Router /system/roles/{id}/permissions [put]
func (h *RoleHandler) AssignPermissions(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "无效的角色ID")
		return
	}

	var req dto.AssignPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := h.roleService.AssignPermissions(c.Request.Context(), id, &req); err != nil {
		response.Error(c, err)
		return
	}

	response.OKWithoutData(c)
}

// GetRoleUsers 获取角色下用户
// @Summary 获取角色下用户
// @Description 分页查询指定角色关联的用户列表
// @Tags 角色管理
// @Produce json
// @Security Bearer
// @Param id path string true "角色ID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} response.Response{data=response.PageResponse{list=[]dto.RoleUserResponse}}
// @Router /system/roles/{id}/users [get]
func (h *RoleHandler) GetRoleUsers(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "无效的角色ID")
		return
	}

	var req struct {
		Page     int `form:"page,default=1" binding:"min=1"`
		PageSize int `form:"page_size,default=10" binding:"min=1,max=100"`
	}
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	list, total, err := h.roleService.GetRoleUsers(c.Request.Context(), id, req.Page, req.PageSize)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Page(c, list, total, req.Page, req.PageSize)
}
