package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/Yogdunana/StarByte/backend/internal/task/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

// GetWorkflow godoc
// @Summary 查看指定参与人的任务审核验收进度
// @Tags tasks
// @Produce json
// @Param id path string true "任务 ID"
// @Router /tasks/{id}/workflow [get]
func (h *TaskHandler) GetWorkflow(c *gin.Context) {
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
	out, err := h.svc.GetWorkflow(c.Request.Context(), id, actor)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// ActWorkflow godoc
// @Summary 提交交付、签字审核、拒绝或认领
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "任务 ID"
// @Param body body dto.WorkflowActionRequest true "操作及说明"
// @Router /tasks/{id}/workflow/actions [post]
func (h *TaskHandler) ActWorkflow(c *gin.Context) {
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
	var req dto.WorkflowActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	out, err := h.svc.ActWorkflow(c.Request.Context(), id, actor, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}
