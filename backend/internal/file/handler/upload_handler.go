package handler

import (
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

const maxMultipartMemory = 32 << 20 // 32MB 进内存，其余落盘

// Upload POST /api/v1/files/upload
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
