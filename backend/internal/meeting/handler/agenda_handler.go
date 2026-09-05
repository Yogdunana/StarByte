package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/meeting/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ListAgendas 议程列表
// @Summary 议程列表
// @Tags 会议
// @Router /meetings/{id}/agendas [get]
func (h *MeetingHandler) ListAgendas(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	out, err := h.svc.ListAgendas(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// AddAgenda 添加议程
// @Summary 添加议程
// @Tags 会议
// @Router /meetings/{id}/agendas [post]
func (h *MeetingHandler) AddAgenda(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.CreateAgendaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.AddAgenda(c.Request.Context(), id, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// UpdateAgenda 更新议程
// @Summary 更新议程
// @Tags 会议
// @Router /meetings/{id}/agendas/{aid} [put]
func (h *MeetingHandler) UpdateAgenda(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	aid, err := parseNamedID(c, "aid")
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.UpdateAgendaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.UpdateAgenda(c.Request.Context(), id, aid, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// DeleteAgenda 删除议程
// @Summary 删除议程
// @Tags 会议
// @Router /meetings/{id}/agendas/{aid} [delete]
func (h *MeetingHandler) DeleteAgenda(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	aid, err := parseNamedID(c, "aid")
	if err != nil {
		response.Error(c, err)
		return
	}
	if err := h.svc.DeleteAgenda(c.Request.Context(), id, aid); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}

// SortAgendas 排序议程
// @Summary 排序议程
// @Tags 会议
// @Router /meetings/{id}/agendas/sort [put]
func (h *MeetingHandler) SortAgendas(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.SortAgendasRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	ids := make([]uuid.UUID, 0, len(req.AgendaIDs))
	for _, raw := range req.AgendaIDs {
		aid, err := uuid.Parse(raw)
		if err != nil {
			response.BadRequest(c, "无效的议程ID")
			return
		}
		ids = append(ids, aid)
	}
	out, err := h.svc.SortAgendas(c.Request.Context(), id, ids)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}
