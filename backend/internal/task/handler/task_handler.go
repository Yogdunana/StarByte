package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/task/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// CreateTask 创建任务
// @Summary 创建任务
// @Tags 任务
// @Accept json
// @Produce json
// @Param request body dto.CreateTaskRequest true "任务"
// @Success 200 {object} response.Response
// @Router /tasks [post]
func (h *TaskHandler) CreateTask(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.Create(c.Request.Context(), userID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// ListTasks 任务列表
// @Summary 任务列表
// @Tags 任务
// @Produce json
// @Router /tasks [get]
func (h *TaskHandler) ListTasks(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.ListTaskRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	list, total, err := h.svc.List(c.Request.Context(), userID, &req, dataScope(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	page, size := defaultPage(req.Page, req.PageSize)
	response.Page(c, list, total, page, size)
}

// GetTask 任务详情
// @Summary 任务详情
// @Tags 任务
// @Router /tasks/{id} [get]
func (h *TaskHandler) GetTask(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	out, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// UpdateTask 更新任务
// @Summary 更新任务
// @Tags 任务
// @Router /tasks/{id} [put]
func (h *TaskHandler) UpdateTask(c *gin.Context) {
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
	var req dto.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.Update(c.Request.Context(), id, userID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// DeleteTask 删除任务
// @Summary 删除任务
// @Tags 任务
// @Router /tasks/{id} [delete]
func (h *TaskHandler) DeleteTask(c *gin.Context) {
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
	if err := h.svc.Delete(c.Request.Context(), id, userID); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}
