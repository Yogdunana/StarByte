package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/configstore/dto"
	"github.com/Yogdunana/StarByte/backend/pkg/locale"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// GetSMTP 读取 SMTP 设置（不含密码）
// @Summary 读取 SMTP 设置
// @Description 返回可编辑的 SMTP 字段与密码是否已通过环境变量配置；不回传密码
// @Tags 系统配置
// @Produce json
// @Success 200 {object} response.Response{data=dto.SMTPSettingsResponse}
// @Failure 401 {object} response.Response
// @Router /system/smtp [get]
// @Security BearerAuth
func (h *ConfigHandler) GetSMTP(c *gin.Context) {
	out, err := h.svc.GetSMTP(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// UpdateSMTP 更新 SMTP 非密钥字段
// @Summary 更新 SMTP 设置
// @Description 保存 host/port/SSL/发件人等；密码只能通过 STARBYTE_SMTP_PASSWORD
// @Tags 系统配置
// @Accept json
// @Produce json
// @Param request body dto.UpdateSMTPRequest true "SMTP 设置"
// @Success 200 {object} response.Response{data=dto.SMTPSettingsResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /system/smtp [put]
// @Security BearerAuth
func (h *ConfigHandler) UpdateSMTP(c *gin.Context) {
	operator, err := currentUser(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.UpdateSMTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.UpdateSMTP(c.Request.Context(), operator, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// TestSMTP 管理员发送一封测试邮件
// @Summary 发送 SMTP 测试邮件
// @Description 使用当前设置立即发一封测试信，用于验证部署侧密码与连通性
// @Tags 系统配置
// @Accept json
// @Produce json
// @Param request body dto.TestSMTPRequest true "收件人"
// @Success 200 {object} response.Response{data=dto.TestSMTPResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /system/smtp/test [post]
// @Security BearerAuth
func (h *ConfigHandler) TestSMTP(c *gin.Context) {
	var req dto.TestSMTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.TestSMTP(locale.WithContext(c.Request.Context(), locale.FromHeader(c.GetHeader("Accept-Language"))), &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}
