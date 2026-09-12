package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/Yogdunana/StarByte/backend/internal/task/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

// RequestHandover godoc
// @Summary 委托或申请转办流程任务
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "任务 ID"
// @Param body body dto.HandoverRequest true "接收人与原因"
// @Router /tasks/{id}/handover [post]
func (h *TaskHandler) RequestHandover(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	actor, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.HandoverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	out, err := h.svc.RequestHandover(c.Request.Context(), id, actor, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// GetHandover godoc
// @Summary 查看任务当前转办
// @Tags tasks
// @Produce json
// @Param id path string true "任务 ID"
// @Router /tasks/{id}/handover [get]
func (h *TaskHandler) GetHandover(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	actor, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	out, err := h.svc.GetHandover(c.Request.Context(), id, actor)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// DecideHandover godoc
// @Summary 负责人签署任务转办
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "任务 ID"
// @Param body body dto.HandoverDecision true "签字决定"
// @Router /tasks/{id}/handover/decisions [post]
func (h *TaskHandler) DecideHandover(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	actor, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.HandoverDecision
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	out, err := h.svc.DecideHandover(c.Request.Context(), id, actor, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// GetTransfer godoc
// @Summary 按转办单查看签字进度
// @Tags tasks
// @Produce json
// @Param id path string true "转办 ID"
// @Router /tasks/transfers/{id} [get]
func (h *TaskHandler) GetTransfer(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	actor, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	out, err := h.svc.GetTransfer(c.Request.Context(), id, actor)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// DecideTransfer godoc
// @Summary 按转办单签署
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "转办 ID"
// @Param body body dto.HandoverDecision true "签字决定"
// @Router /tasks/transfers/{id}/decisions [post]
func (h *TaskHandler) DecideTransfer(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	actor, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.HandoverDecision
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	out, err := h.svc.DecideTransfer(c.Request.Context(), id, actor, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}
