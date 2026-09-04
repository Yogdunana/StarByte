package handler

import (
	"net/http"
	"time"

	"github.com/Yogdunana/StarByte/backend/internal/file/dto"
	"github.com/Yogdunana/StarByte/backend/internal/file/service"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// FileHandler 文件管理处理器
type FileHandler struct {
	fileService service.FileService
}

// NewFileHandler 创建文件处理器
func NewFileHandler(fileService service.FileService) *FileHandler {
	return &FileHandler{fileService: fileService}
}

// List GET /api/v1/files
func (h *FileHandler) List(c *gin.Context) {
	var req dto.ListFilesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	list, total, err := h.fileService.List(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	response.Page(c, list, total, req.Page, req.PageSize)
}

// GetByID GET /api/v1/files/:id
func (h *FileHandler) GetByID(c *gin.Context) {
	id, err := parseFileID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	result, err := h.fileService.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

// Download GET /api/v1/files/:id/download
func (h *FileHandler) Download(c *gin.Context) {
	id, err := parseFileID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	url, err := h.fileService.PresignDownload(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	c.Redirect(http.StatusFound, url)
}

// Delete DELETE /api/v1/files/:id
func (h *FileHandler) Delete(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	id, err := parseFileID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	if err := h.fileService.Delete(c.Request.Context(), id, userID); err != nil {
		response.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Response{
		Code:      response.CodeSuccess,
		Message:   "删除成功",
		Data:      nil,
		RequestID: c.GetString("request_id"),
		Timestamp: time.Now().Unix(),
	})
}
