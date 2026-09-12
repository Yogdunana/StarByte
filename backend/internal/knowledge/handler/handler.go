package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/knowledge/dto"
	"github.com/Yogdunana/StarByte/backend/internal/knowledge/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *Handler) Create(c *gin.Context) {
	v, err := viewerOf(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.CreateDocRequest
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
	var req dto.UpdateDocRequest
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
	var req dto.ListDocRequest
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

func (h *Handler) History(c *gin.Context) {
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
	result, err := h.svc.History(c.Request.Context(), v, id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *Handler) GetVersion(c *gin.Context) {
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
	ver, convErr := parseVersion(c)
	if convErr != nil {
		response.Error(c, convErr)
		return
	}
	result, err := h.svc.GetVersion(c.Request.Context(), v, id, ver)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *Handler) Rollback(c *gin.Context) {
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
	var req dto.RollbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, err := h.svc.Rollback(c.Request.Context(), v, id, req.Version)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *Handler) Search(c *gin.Context) {
	v, err := viewerOf(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.SearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	req.Page, req.PageSize = defaultPage(req.Page, req.PageSize)
	list, total, err := h.svc.Search(c.Request.Context(), v, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Page(c, list, total, req.Page, req.PageSize)
}

func (h *Handler) Tree(c *gin.Context) {
	v, err := viewerOf(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	result, err := h.svc.Tree(c.Request.Context(), v)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *Handler) CreateCategory(c *gin.Context) {
	v, err := viewerOf(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, err := h.svc.CreateCategory(c.Request.Context(), v, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *Handler) UpdateCategory(c *gin.Context) {
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
	var req dto.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	result, err := h.svc.UpdateCategory(c.Request.Context(), v, id, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *Handler) DeleteCategory(c *gin.Context) {
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
	if err := h.svc.DeleteCategory(c.Request.Context(), v, id); err != nil {
		response.Error(c, err)
		return
	}
	response.OKWithoutData(c)
}

func (h *Handler) ListCategories(c *gin.Context) {
	if _, err := viewerOf(c); err != nil {
		response.Error(c, err)
		return
	}
	result, err := h.svc.ListCategories(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *Handler) Attach(c *gin.Context) {
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
	var req dto.AttachRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	fileID, err := uuid.Parse(req.FileID)
	if err != nil {
		response.BadRequest(c, "无效的文件ID")
		return
	}
	result, err := h.svc.Attach(c.Request.Context(), v, id, fileID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *Handler) Detach(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	fileID, err := uuid.Parse(c.Param("fileId"))
	if err != nil {
		response.BadRequest(c, "无效的文件ID")
		return
	}
	v, err := viewerOf(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	if err := h.svc.Detach(c.Request.Context(), v, id, fileID); err != nil {
		response.Error(c, err)
		return
	}
	response.OKWithoutData(c)
}

func (h *Handler) PublicPage(c *gin.Context) {
	h.publicBySlug(c, model.KindPage, c.Param("slug"))
}

func (h *Handler) PublicDoc(c *gin.Context) {
	h.publicBySlug(c, model.KindDoc, c.Param("slug"))
}

func (h *Handler) publicBySlug(c *gin.Context, kind, slug string) {
	v := optionalViewer(c)
	result, err := h.svc.GetBySlug(c.Request.Context(), v, kind, slug)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *Handler) PublicDocs(c *gin.Context) {
	v := optionalViewer(c)
	pub := model.VisibilityPublic
	status := model.StatusPublished
	if v.Authenticated() {
		pub = ""
	}
	kind := model.KindDoc
	req := &dto.ListDocRequest{Kind: kind, Visibility: pub, Status: &status, IncludeBody: false}
	req.Page, req.PageSize = defaultPage(0, 50)
	if v.Authenticated() {
		req.Visibility = ""
	}
	list, total, err := h.svc.List(c.Request.Context(), v, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Page(c, list, total, req.Page, req.PageSize)
}

func (h *Handler) PublicTree(c *gin.Context) {
	v := optionalViewer(c)
	result, err := h.svc.Tree(c.Request.Context(), v)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *Handler) PublicSearch(c *gin.Context) {
	v := optionalViewer(c)
	var req dto.SearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	req.Page, req.PageSize = defaultPage(req.Page, req.PageSize)
	list, total, err := h.svc.Search(c.Request.Context(), v, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Page(c, list, total, req.Page, req.PageSize)
}

func parseVersion(c *gin.Context) (int, error) {
	raw := c.Param("version")
	var n int
	for _, ch := range raw {
		if ch < '0' || ch > '9' {
			return 0, response.NewError(response.CodeBadRequest, "无效的版本号")
		}
		n = n*10 + int(ch-'0')
	}
	if n <= 0 {
		return 0, response.NewError(response.CodeBadRequest, "无效的版本号")
	}
	return n, nil
}
