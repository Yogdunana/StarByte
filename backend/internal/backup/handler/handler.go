package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/backup/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// List 备份列表
// @Summary 备份列表
// @Tags 备份
// @Produce json
// @Param page query int false "页码"
// @Param page_size query int false "每页条数"
// @Param status query int false "状态"
// @Success 200 {object} response.Response
// @Router /system/backups [get]
// @Security BearerAuth
func (h *Handler) List(c *gin.Context) {
	var req dto.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	req.Page, req.PageSize = defaultPage(req.Page, req.PageSize)
	list, total, err := h.svc.List(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Page(c, list, total, req.Page, req.PageSize)
}

// Create 创建全量备份
// @Summary 创建全量备份
// @Tags 备份
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Router /system/backups [post]
// @Security BearerAuth
func (h *Handler) Create(c *gin.Context) {
	uid, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.CreateRequest
	_ = c.ShouldBindJSON(&req)
	out, err := h.svc.Create(c.Request.Context(), uid, &req)
	write(c, out, err)
}

// Get 备份详情
// @Summary 备份详情
// @Tags 备份
// @Produce json
// @Param id path string true "备份 ID"
// @Success 200 {object} response.Response
// @Router /system/backups/{id} [get]
// @Security BearerAuth
func (h *Handler) Get(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	out, err := h.svc.Get(c.Request.Context(), id)
	write(c, out, err)
}

// Delete 删除备份
// @Summary 删除备份产物与记录
// @Tags 备份
// @Param id path string true "备份 ID"
// @Success 200 {object} response.Response
// @Router /system/backups/{id} [delete]
// @Security BearerAuth
func (h *Handler) Delete(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	response.OKWithoutData(c)
}

// Restore 从备份恢复（危险）
// @Summary 恢复备份
// @Tags 备份
// @Accept json
// @Param id path string true "备份 ID"
// @Success 200 {object} response.Response
// @Router /system/backups/{id}/restore [post]
// @Security BearerAuth
func (h *Handler) Restore(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	uid, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.RestoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.Restore(c.Request.Context(), uid, id, &req)
	write(c, out, err)
}

// GetPolicy 备份策略
// @Summary 备份保留与调度策略
// @Tags 备份
// @Produce json
// @Success 200 {object} response.Response
// @Router /system/backups/policies [get]
// @Security BearerAuth
func (h *Handler) GetPolicy(c *gin.Context) {
	out, err := h.svc.GetPolicy(c.Request.Context())
	write(c, out, err)
}

// PutPolicy 更新策略
// @Summary 更新备份策略
// @Tags 备份
// @Accept json
// @Success 200 {object} response.Response
// @Router /system/backups/policies [put]
// @Security BearerAuth
func (h *Handler) PutPolicy(c *gin.Context) {
	uid, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.UpdatePolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.UpdatePolicy(c.Request.Context(), uid, &req)
	write(c, out, err)
}

// Storage 存储统计
// @Summary 备份存储统计
// @Tags 备份
// @Produce json
// @Success 200 {object} response.Response
// @Router /system/backups/storage [get]
// @Security BearerAuth
func (h *Handler) Storage(c *gin.Context) {
	out, err := h.svc.Storage(c.Request.Context())
	write(c, out, err)
}

// Preview 恢复预览 / 完整性检查（不执行恢复）
// @Summary 备份完整性检查
// @Tags 备份
// @Produce json
// @Param id path string true "备份 ID"
// @Success 200 {object} response.Response
// @Router /system/backups/{id}/preview [get]
// @Security BearerAuth
func (h *Handler) Preview(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	out, err := h.svc.Preview(c.Request.Context(), id)
	write(c, out, err)
}
