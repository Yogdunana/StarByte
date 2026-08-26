package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/rbac/dto"
	"github.com/Yogdunana/StarByte/backend/internal/rbac/service"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// DepartmentHandler 部门处理器
type DepartmentHandler struct {
	deptService service.DepartmentService
}

// NewDepartmentHandler 创建部门处理器
func NewDepartmentHandler(deptService service.DepartmentService) *DepartmentHandler {
	return &DepartmentHandler{deptService: deptService}
}

// GetTree 部门树
// @Summary 获取部门树
// @Description 获取完整的部门树形结构
// @Tags 部门管理
// @Produce json
// @Security Bearer
// @Success 200 {object} response.Response{data=[]dto.DepartmentTreeResponse}
// @Router /system/departments [get]
func (h *DepartmentHandler) GetTree(c *gin.Context) {
	tree, err := h.deptService.GetTree(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, tree)
}

// GetByID 获取部门详情
// @Summary 获取部门详情
// @Description 根据ID获取部门详细信息
// @Tags 部门管理
// @Produce json
// @Security Bearer
// @Param id path string true "部门ID"
// @Success 200 {object} response.Response{data=dto.DepartmentResponse}
// @Router /system/departments/{id} [get]
func (h *DepartmentHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "无效的部门ID")
		return
	}

	result, err := h.deptService.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, result)
}

// Create 创建部门
// @Summary 创建部门
// @Description 创建新部门
// @Tags 部门管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body dto.CreateDepartmentRequest true "部门信息"
// @Success 200 {object} response.Response{data=dto.DepartmentResponse}
// @Router /system/departments [post]
func (h *DepartmentHandler) Create(c *gin.Context) {
	var req dto.CreateDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, parseBindingError(err))
		return
	}

	if req.ParentID != "" {
		if _, err := uuid.Parse(req.ParentID); err != nil {
			response.BadRequest(c, "无效的父部门ID")
			return
		}
	}
	if req.LeaderID != "" {
		if _, err := uuid.Parse(req.LeaderID); err != nil {
			response.BadRequest(c, "无效的负责人ID")
			return
		}
	}

	result, err := h.deptService.Create(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, result)
}

// Update 更新部门
// @Summary 更新部门
// @Description 更新部门信息
// @Tags 部门管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "部门ID"
// @Param request body dto.UpdateDepartmentRequest true "部门信息"
// @Success 200 {object} response.Response{data=dto.DepartmentResponse}
// @Router /system/departments/{id} [put]
func (h *DepartmentHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "无效的部门ID")
		return
	}

	var req dto.UpdateDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, parseBindingError(err))
		return
	}

	if req.LeaderID != nil && *req.LeaderID != "" {
		if _, err := uuid.Parse(*req.LeaderID); err != nil {
			response.BadRequest(c, "无效的负责人ID")
			return
		}
	}

	result, err := h.deptService.Update(c.Request.Context(), id, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, result)
}

// Delete 删除部门
// @Summary 删除部门
// @Description 删除指定部门
// @Tags 部门管理
// @Produce json
// @Security Bearer
// @Param id path string true "部门ID"
// @Success 200 {object} response.Response
// @Router /system/departments/{id} [delete]
func (h *DepartmentHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "无效的部门ID")
		return
	}

	if err := h.deptService.Delete(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}

	response.OKWithoutData(c)
}
