package handler

import (
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

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
