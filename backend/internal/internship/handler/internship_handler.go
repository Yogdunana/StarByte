package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/internship/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// CreateInternship 创建实习记录
// @Summary 创建实习记录
// @Tags 实习
// @Accept json
// @Produce json
// @Param request body dto.CreateInternshipRequest true "实习"
// @Success 200 {object} response.Response
// @Router /internships [post]
func (h *InternshipHandler) CreateInternship(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.CreateInternshipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.Create(c.Request.Context(), userID, &req, dataScope(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// ListInternships 实习列表
// @Summary 实习列表
// @Tags 实习
// @Produce json
// @Router /internships [get]
func (h *InternshipHandler) ListInternships(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.ListInternshipRequest
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

// GetInternship 实习详情
// @Summary 实习详情
// @Tags 实习
// @Produce json
// @Param id path string true "实习ID"
// @Success 200 {object} response.Response
// @Router /internships/{id} [get]
func (h *InternshipHandler) GetInternship(c *gin.Context) {
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
	out, err := h.svc.Get(c.Request.Context(), userID, id, dataScope(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// UpdateInternship 更新实习
// @Summary 更新实习
// @Tags 实习
// @Accept json
// @Produce json
// @Param id path string true "实习ID"
// @Param request body dto.UpdateInternshipRequest true "更新"
// @Success 200 {object} response.Response
// @Router /internships/{id} [put]
func (h *InternshipHandler) UpdateInternship(c *gin.Context) {
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
	var req dto.UpdateInternshipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.Update(c.Request.Context(), userID, id, &req, dataScope(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// DeleteInternship 删除实习
// @Summary 删除实习
// @Tags 实习
// @Produce json
// @Param id path string true "实习ID"
// @Success 200 {object} response.Response
// @Router /internships/{id} [delete]
func (h *InternshipHandler) DeleteInternship(c *gin.Context) {
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
	if err := h.svc.Delete(c.Request.Context(), userID, id, dataScope(c)); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}

// ListMine 我的实习
// @Summary 我的实习
// @Tags 实习
// @Produce json
// @Router /internships/my [get]
func (h *InternshipHandler) ListMine(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.MyInternshipRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	list, err := h.svc.ListMine(c.Request.Context(), userID, req.Status)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, list)
}
