package handler

import (
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Pause 暂停定时任务
// @Summary 暂停定时任务
// @Description 暂停后不再调度执行
// @Tags 调度
// @Produce json
// @Param id path string true "任务 ID"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /system/scheduler/tasks/{id}/pause [post]
// @Security BearerAuth
func (h *SchedulerHandler) Pause(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	out, err := h.svc.Pause(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// Resume 恢复定时任务
// @Summary 恢复定时任务
// @Description 恢复已暂停的定时任务
// @Tags 调度
// @Produce json
// @Param id path string true "任务 ID"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /system/scheduler/tasks/{id}/resume [post]
// @Security BearerAuth
func (h *SchedulerHandler) Resume(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	out, err := h.svc.Resume(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// Run 立即执行定时任务
// @Summary 立即执行定时任务
// @Description 手动触发一次调度执行
// @Tags 调度
// @Produce json
// @Param id path string true "任务 ID"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /system/scheduler/tasks/{id}/run [post]
// @Security BearerAuth
func (h *SchedulerHandler) Run(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	if err := h.svc.RunNow(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	response.OKWithoutData(c)
}

// Logs 定时任务运行日志
// @Summary 定时任务运行日志
// @Description 查询任务运行记录，可按 run_id 过滤
// @Tags 调度
// @Produce json
// @Param id path string true "任务 ID"
// @Param run_id query string false "运行记录 ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /system/scheduler/tasks/{id}/logs [get]
// @Security BearerAuth
func (h *SchedulerHandler) Logs(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var runID *uuid.UUID
	if raw := c.Query("run_id"); raw != "" {
		parsed, perr := uuid.Parse(raw)
		if perr != nil {
			response.BadRequest(c, "无效的 run_id")
			return
		}
		runID = &parsed
	}
	out, err := h.svc.Logs(c.Request.Context(), id, runID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}
