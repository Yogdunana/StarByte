package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/task/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// Assign 分配任务
// @Summary 分配任务
// @Tags 任务
// @Router /tasks/{id}/assign [post]
func (h *TaskHandler) Assign(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	userID, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.AssignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.Assign(c.Request.Context(), id, userID, req.AssigneeID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// Transfer 转办任务
// @Summary 转办任务
// @Tags 任务
// @Router /tasks/{id}/transfer [post]
func (h *TaskHandler) Transfer(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	userID, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.Transfer(c.Request.Context(), id, userID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// ChangeStatus 更新状态
// @Summary 更新任务状态
// @Tags 任务
// @Router /tasks/{id}/status [post]
func (h *TaskHandler) ChangeStatus(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	userID, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.StatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.ChangeStatus(c.Request.Context(), id, userID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// Urge 催办
// @Summary 催办任务
// @Tags 任务
// @Router /tasks/{id}/urge [post]
func (h *TaskHandler) Urge(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	userID, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.UrgeRequest
	_ = c.ShouldBindJSON(&req)
	if err := h.svc.Urge(c.Request.Context(), id, userID, req.Message); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}

// ListLogs 流转记录
// @Summary 任务流转记录
// @Tags 任务
// @Router /tasks/{id}/logs [get]
func (h *TaskHandler) ListLogs(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	out, err := h.svc.ListLogs(c.Request.Context(), userID, id, dataScope(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// Stats 任务统计
// @Summary 任务统计
// @Tags 任务
// @Router /tasks/stats [get]
func (h *TaskHandler) Stats(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.StatsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.Stats(c.Request.Context(), userID, &req, dataScope(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}
