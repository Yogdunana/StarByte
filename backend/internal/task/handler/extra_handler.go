package handler

import (
	"fmt"
	"io"
	"net/url"

	"github.com/Yogdunana/StarByte/backend/internal/task/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// ListComments 评论列表
// @Summary 任务评论列表
// @Description 列出任务下的评论
// @Tags 任务
// @Produce json
// @Param id path string true "任务 ID"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /tasks/{id}/comments [get]
// @Security BearerAuth
func (h *TaskHandler) ListComments(c *gin.Context) {
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
	out, err := h.svc.ListComments(c.Request.Context(), userID, id, dataScope(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// AddComment 添加评论
// @Summary 添加任务评论
// @Description 在任务下新增一条评论
// @Tags 任务
// @Accept json
// @Produce json
// @Param id path string true "任务 ID"
// @Param request body dto.CommentRequest true "评论"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /tasks/{id}/comments [post]
// @Security BearerAuth
func (h *TaskHandler) AddComment(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	userID, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.CommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.AddComment(c.Request.Context(), id, userID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// UpdateComment 更新评论
// @Summary 更新任务评论
// @Description 修改一条任务评论
// @Tags 任务
// @Accept json
// @Produce json
// @Param id path string true "任务 ID"
// @Param cid path string true "评论 ID"
// @Param request body dto.CommentRequest true "评论"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /tasks/{id}/comments/{cid} [put]
// @Security BearerAuth
func (h *TaskHandler) UpdateComment(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	cid, err := parseNamedID(c, "cid")
	if err != nil {
		response.Error(c, err)
		return
	}
	userID, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.CommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.UpdateComment(c.Request.Context(), id, cid, userID, req.Content)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// DeleteComment 删除评论
// @Summary 删除任务评论
// @Description 删除一条任务评论
// @Tags 任务
// @Produce json
// @Param id path string true "任务 ID"
// @Param cid path string true "评论 ID"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /tasks/{id}/comments/{cid} [delete]
// @Security BearerAuth
func (h *TaskHandler) DeleteComment(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	cid, err := parseNamedID(c, "cid")
	if err != nil {
		response.Error(c, err)
		return
	}
	userID, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	if err := h.svc.DeleteComment(c.Request.Context(), id, cid, userID); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}

// ListAttachments 附件列表
// @Summary 任务附件列表
// @Description 列出任务附件
// @Tags 任务
// @Produce json
// @Param id path string true "任务 ID"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /tasks/{id}/attachments [get]
// @Security BearerAuth
func (h *TaskHandler) ListAttachments(c *gin.Context) {
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
	out, err := h.svc.ListAttachments(c.Request.Context(), userID, id, dataScope(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// UploadAttachment 上传附件
// @Summary 上传任务附件
// @Description 为任务上传附件文件
// @Tags 任务
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "任务 ID"
// @Param file formData file true "文件"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /tasks/{id}/attachments [post]
// @Security BearerAuth
func (h *TaskHandler) UploadAttachment(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	userID, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	header, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请选择要上传的文件")
		return
	}
	out, err := h.svc.UploadAttachment(c.Request.Context(), id, userID, header)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// DownloadAttachment 下载附件
// @Summary 下载任务附件
// @Description 下载任务附件文件
// @Tags 任务
// @Produce octet-stream
// @Param id path string true "任务 ID"
// @Param aid path string true "附件 ID"
// @Success 200 {file} file
// @Failure 401 {object} response.Response
// @Router /tasks/{id}/attachments/{aid} [get]
// @Security BearerAuth
func (h *TaskHandler) DownloadAttachment(c *gin.Context) {
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
	userID, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	rc, filename, contentType, err := h.svc.DownloadAttachment(c.Request.Context(), userID, id, aid, dataScope(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	defer rc.Close()
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", url.QueryEscape(filename)))
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	c.Header("Content-Type", contentType)
	c.Status(200)
	_, _ = io.Copy(c.Writer, rc)
}

// DeleteAttachment 删除附件
// @Summary 删除任务附件
// @Description 删除一条任务附件
// @Tags 任务
// @Produce json
// @Param id path string true "任务 ID"
// @Param aid path string true "附件 ID"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /tasks/{id}/attachments/{aid} [delete]
// @Security BearerAuth
func (h *TaskHandler) DeleteAttachment(c *gin.Context) {
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
	userID, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	if err := h.svc.DeleteAttachment(c.Request.Context(), id, aid, userID); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}
