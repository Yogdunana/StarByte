package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/scheduler/dto"
	"github.com/Yogdunana/StarByte/backend/internal/scheduler/service"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware/auth"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SchedulerHandler struct {
	svc service.SchedulerService
}

func NewSchedulerHandler(svc service.SchedulerService) *SchedulerHandler {
	return &SchedulerHandler{svc: svc}
}

func parseID(c *gin.Context) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return uuid.Nil, response.NewError(response.CodeBadRequest, "无效的ID")
	}
	return id, nil
}

func currentUser(c *gin.Context) (uuid.UUID, error) {
	raw := auth.GetUserID(c)
	if raw == "" {
		return uuid.Nil, response.NewUnauthorizedError("用户未认证")
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, response.NewError(response.CodeBadRequest, "无效的用户ID")
	}
	return id, nil
}

// List 定时任务列表
// @Summary 定时任务列表
// @Description 分页查询系统定时任务
// @Tags 调度
// @Produce json
// @Param page query int false "页码"
// @Param page_size query int false "每页条数"
// @Param keyword query string false "关键词"
// @Param status query int false "状态"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /system/scheduler/tasks [get]
// @Security BearerAuth
func (h *SchedulerHandler) List(c *gin.Context) {
	var q dto.ListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	list, total, page, size, err := h.svc.List(c.Request.Context(), q)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Page(c, list, total, page, size)
}

// Get 定时任务详情
// @Summary 定时任务详情
// @Description 按 ID 获取定时任务配置
// @Tags 调度
// @Produce json
// @Param id path string true "任务 ID"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /system/scheduler/tasks/{id} [get]
// @Security BearerAuth
func (h *SchedulerHandler) Get(c *gin.Context) {
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

// Handlers 可注册的任务处理器
// @Summary 可注册的任务处理器
// @Description 列出调度器已注册的 handler_key
// @Tags 调度
// @Produce json
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /system/scheduler/handlers [get]
// @Security BearerAuth
func (h *SchedulerHandler) Handlers(c *gin.Context) {
	response.OK(c, h.svc.Handlers())
}

// Create 创建定时任务
// @Summary 创建定时任务
// @Description 创建 cron 或一次性调度任务
// @Tags 调度
// @Accept json
// @Produce json
// @Param request body object true "定时任务"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /system/scheduler/tasks [post]
// @Security BearerAuth
func (h *SchedulerHandler) Create(c *gin.Context) {
	userID, err := currentUser(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.Create(c.Request.Context(), userID, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// Update 更新定时任务
// @Summary 更新定时任务
// @Description 更新定时任务配置
// @Tags 调度
// @Accept json
// @Produce json
// @Param id path string true "任务 ID"
// @Param request body object true "更新内容"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /system/scheduler/tasks/{id} [put]
// @Security BearerAuth
func (h *SchedulerHandler) Update(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// Delete 删除定时任务
// @Summary 删除定时任务
// @Description 删除定时任务及其运行记录
// @Tags 调度
// @Produce json
// @Param id path string true "任务 ID"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /system/scheduler/tasks/{id} [delete]
// @Security BearerAuth
func (h *SchedulerHandler) Delete(c *gin.Context) {
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
