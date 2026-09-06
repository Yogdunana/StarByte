package handler

import (
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

const maxMultipartMemory = 32 << 20 // 32MB 进内存，其余落盘

// Upload POST /api/v1/files/upload
// @Summary 上传文件
// @Description 单文件上传到对象存储
// @Tags 文件
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "文件"
// @Param category formData string false "分类"
// @Param is_public formData bool false "是否公开"
// @Success 200 {object} response.Response{data=dto.FileUploadResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /files/upload [post]
// @Security BearerAuth
func (h *FileHandler) Upload(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	_ = c.Request.ParseMultipartForm(maxMultipartMemory)
	header, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请选择要上传的文件")
		return
	}
	result, err := h.fileService.Upload(
		c.Request.Context(),
		userID,
		header,
		c.PostForm("category"),
		parseBoolForm(c, "is_public"),
	)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

// UploadBatch POST /api/v1/files/upload-batch
// @Summary 批量上传文件
// @Description 一次上传多个文件，表单字段名为 files 或 files[]
// @Tags 文件
// @Accept multipart/form-data
// @Produce json
// @Param files formData file true "文件列表"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /files/upload-batch [post]
// @Security BearerAuth
func (h *FileHandler) UploadBatch(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	form, err := c.MultipartForm()
	if err != nil {
		response.BadRequest(c, "请选择要上传的文件")
		return
	}
	headers := form.File["files"]
	if len(headers) == 0 {
		headers = form.File["files[]"]
	}
	result, err := h.fileService.UploadBatch(c.Request.Context(), userID, headers)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}
