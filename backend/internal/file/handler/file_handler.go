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
// @Summary 文件列表
// @Description 分页查询已上传文件
// @Tags 文件
// @Produce json
// @Param page query int false "页码"
// @Param page_size query int false "每页条数"
// @Param category query string false "分类"
// @Param keyword query string false "关键词"
// @Param uploader_id query string false "上传者 ID"
// @Param mime_type query string false "MIME 类型"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /files [get]
// @Security BearerAuth
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
// @Summary 文件详情
// @Description 按 ID 获取文件元数据
// @Tags 文件
// @Produce json
// @Param id path string true "文件 ID"
// @Success 200 {object} response.Response{data=dto.FileDetailResponse}
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /files/{id} [get]
// @Security BearerAuth
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
// @Summary 下载文件
// @Description 签发预签名地址并 302 跳转
// @Tags 文件
// @Produce json
// @Param id path string true "文件 ID"
// @Success 302 {string} string "跳转到对象存储预签名 URL"
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /files/{id}/download [get]
// @Security BearerAuth
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
// @Summary 删除文件
// @Description 删除文件记录及对象存储中的对象；上传者或持 file:delete 可删
// @Tags 文件
// @Produce json
// @Param id path string true "文件 ID"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /files/{id} [delete]
// @Security BearerAuth
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
