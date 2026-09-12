package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/announcement/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

func (h *Handler) Create(c *gin.Context) {
	v, err := viewerOf(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.CreateAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, err := h.svc.Create(c.Request.Context(), v, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *Handler) Update(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	v, err := viewerOf(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.UpdateAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, err := h.svc.Update(c.Request.Context(), v, id, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *Handler) Delete(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	v, err := viewerOf(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	if err := h.svc.Delete(c.Request.Context(), v, id); err != nil {
		response.Error(c, err)
		return
	}
	response.OKWithoutData(c)
}

func (h *Handler) Get(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	v, err := viewerOf(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	result, err := h.svc.Get(c.Request.Context(), v, id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *Handler) List(c *gin.Context) {
	v, err := viewerOf(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.ListAnnouncementRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	req.Page, req.PageSize = defaultPage(req.Page, req.PageSize)
	list, total, err := h.svc.List(c.Request.Context(), v, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Page(c, list, total, req.Page, req.PageSize)
}

func (h *Handler) Publish(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	v, err := viewerOf(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	result, err := h.svc.Publish(c.Request.Context(), v, id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *Handler) Pin(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	v, err := viewerOf(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.PinRequest
	_ = c.ShouldBindJSON(&req)
	result, err := h.svc.Pin(c.Request.Context(), v, id, req.Pinned)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *Handler) Archive(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	v, err := viewerOf(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	result, err := h.svc.Archive(c.Request.Context(), v, id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *Handler) MarkRead(c *gin.Context) {
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
	if err := h.svc.MarkRead(c.Request.Context(), uid, id); err != nil {
		response.Error(c, err)
		return
	}
	response.OKWithoutData(c)
}

func (h *Handler) UnreadCount(c *gin.Context) {
	uid, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	result, err := h.svc.UnreadCount(c.Request.Context(), uid)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *Handler) ReadStatus(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	v, err := viewerOf(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	result, err := h.svc.ReadStatus(c.Request.Context(), v, id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}
