package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/leave/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

func (h *Handler) Types(c *gin.Context) {
	result, err := h.svc.ListTypes(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *Handler) Submit(c *gin.Context) {
	v, err := viewerOf(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.SubmitLeaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, err := h.svc.Submit(c.Request.Context(), v, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
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

func (h *Handler) ListMine(c *gin.Context) {
	v, err := viewerOf(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.ListLeaveRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	req.Page, req.PageSize = defaultPage(req.Page, req.PageSize)
	list, total, err := h.svc.ListMine(c.Request.Context(), v, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Page(c, list, total, req.Page, req.PageSize)
}

func (h *Handler) List(c *gin.Context) {
	v, err := viewerOf(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.ListLeaveRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	req.Page, req.PageSize = defaultPage(req.Page, req.PageSize)
	list, total, err := h.svc.ListAll(c.Request.Context(), v, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Page(c, list, total, req.Page, req.PageSize)
}

func (h *Handler) Approve(c *gin.Context) {
	h.decide(c, true)
}

func (h *Handler) Reject(c *gin.Context) {
	h.decide(c, false)
}

func (h *Handler) decide(c *gin.Context, approve bool) {
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
	var req dto.DecisionRequest
	_ = c.ShouldBindJSON(&req)
	if approve {
		err = h.svc.Approve(c.Request.Context(), v, id, req.Remark)
	} else {
		err = h.svc.Reject(c.Request.Context(), v, id, req.Remark)
	}
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OKWithoutData(c)
}

func (h *Handler) Balances(c *gin.Context) {
	v, err := viewerOf(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.ListLeaveRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, err := h.svc.Balances(c.Request.Context(), v, req.UserID, req.Year)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *Handler) Stats(c *gin.Context) {
	v, err := viewerOf(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.ListLeaveRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, err := h.svc.Stats(c.Request.Context(), v, req.Year)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *Handler) Todos(c *gin.Context) {
	v, err := viewerOf(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.ListLeaveRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	req.Page, req.PageSize = defaultPage(req.Page, req.PageSize)
	list, total, err := h.svc.ListTodos(c.Request.Context(), v, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Page(c, list, total, req.Page, req.PageSize)
}

func (h *Handler) Calendar(c *gin.Context) {
	v, err := viewerOf(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.ListLeaveRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, err := h.svc.Calendar(c.Request.Context(), v, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *Handler) CreateType(c *gin.Context) {
	v, err := viewerOf(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.UpsertLeaveTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, err := h.svc.CreateType(c.Request.Context(), v, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *Handler) UpdateType(c *gin.Context) {
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
	var req dto.UpsertLeaveTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, err := h.svc.UpdateType(c.Request.Context(), v, id, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}
