package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/notification/dto"
	"github.com/Yogdunana/StarByte/backend/internal/notification/service"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type EmailHandler struct {
	svc service.EmailService
}

func NewEmailHandler(svc service.EmailService) *EmailHandler {
	return &EmailHandler{svc: svc}
}

// SendEmail POST /api/v1/notifications/email/send
// @Summary 发送邮件
// @Tags 通知
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body dto.SendEmailRequest true "发送参数"
// @Success 200 {object} response.Response{data=dto.SendEmailResponse}
// @Router /notifications/email/send [post]
func (h *EmailHandler) SendEmail(c *gin.Context) {
	var req dto.SendEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.NewError(response.CodeNotificationEmailInvalid, "邮件参数无效"))
		return
	}
	res, err := h.svc.Send(c.Request.Context(), &req, getUserID(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, res)
}

// SendEmailBatch POST /api/v1/notifications/email/batch
// @Summary 批量发送邮件
// @Tags 通知
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body dto.BatchSendEmailRequest true "批量发送"
// @Success 200 {object} response.Response{data=dto.BatchSendEmailResponse}
// @Router /notifications/email/batch [post]
func (h *EmailHandler) SendEmailBatch(c *gin.Context) {
	var req dto.BatchSendEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.NewError(response.CodeNotificationEmailInvalid, "邮件参数无效"))
		return
	}
	res, err := h.svc.SendBatch(c.Request.Context(), &req, getUserID(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, res)
}

// ListEmailLogs GET /api/v1/notifications/email/logs
// @Summary 邮件发送记录
// @Tags 通知
// @Produce json
// @Security Bearer
// @Param page query int false "页码"
// @Param page_size query int false "每页条数"
// @Param status query string false "queued/sent/failed/retrying"
// @Success 200 {object} response.Response
// @Router /notifications/email/logs [get]
func (h *EmailHandler) ListEmailLogs(c *gin.Context) {
	var req dto.ListEmailLogsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, response.NewError(response.CodeNotificationEmailInvalid, "邮件参数无效"))
		return
	}
	list, total, err := h.svc.ListLogs(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Page(c, list, total, req.Page, req.PageSize)
}
