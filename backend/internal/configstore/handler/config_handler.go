package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/configstore/dto"
	"github.com/Yogdunana/StarByte/backend/internal/configstore/service"
	"github.com/Yogdunana/StarByte/backend/pkg/middleware/auth"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ConfigHandler struct {
	svc service.ConfigService
}

func NewConfigHandler(svc service.ConfigService) *ConfigHandler {
	return &ConfigHandler{svc: svc}
}

// List 运行时配置列表
// @Summary 运行时配置列表
// @Description 按分类或关键词列出系统运行时配置
// @Tags 系统配置
// @Produce json
// @Param category query string false "分类"
// @Param keyword query string false "关键词"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /system/configs [get]
// @Security BearerAuth
func (h *ConfigHandler) List(c *gin.Context) {
	var q dto.ListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	list, err := h.svc.List(c.Request.Context(), q)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, list)
}

// GetByKey 按键读取配置
// @Summary 按键读取配置
// @Description 根据 config_key 获取一条运行时配置
// @Tags 系统配置
// @Produce json
// @Param key path string true "配置键"
// @Success 200 {object} response.Response{data=dto.ConfigResponse}
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /system/configs/key/{key} [get]
// @Security BearerAuth
func (h *ConfigHandler) GetByKey(c *gin.Context) {
	out, err := h.svc.GetByKey(c.Request.Context(), c.Param("key"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// Create 创建运行时配置
// @Summary 创建运行时配置
// @Description 新增一条系统运行时配置
// @Tags 系统配置
// @Accept json
// @Produce json
// @Param request body dto.CreateConfigRequest true "配置"
// @Success 200 {object} response.Response{data=dto.ConfigResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /system/configs [post]
// @Security BearerAuth
func (h *ConfigHandler) Create(c *gin.Context) {
	operator, err := currentUser(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req dto.CreateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.Create(c.Request.Context(), operator, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// Update 更新运行时配置
// @Summary 更新运行时配置
// @Description 按 ID 更新运行时配置值
// @Tags 系统配置
// @Accept json
// @Produce json
// @Param id path string true "配置 ID"
// @Param request body dto.UpdateConfigRequest true "更新内容"
// @Success 200 {object} response.Response{data=dto.ConfigResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /system/configs/{id} [put]
// @Security BearerAuth
func (h *ConfigHandler) Update(c *gin.Context) {
	operator, err := currentUser(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "无效的配置ID")
		return
	}
	var req dto.UpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	out, err := h.svc.Update(c.Request.Context(), operator, id, &req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}

// Delete 删除运行时配置
// @Summary 删除运行时配置
// @Description 按 ID 删除一条运行时配置
// @Tags 系统配置
// @Produce json
// @Param id path string true "配置 ID"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /system/configs/{id} [delete]
// @Security BearerAuth
func (h *ConfigHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "无效的配置ID")
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	response.OKWithoutData(c)
}

func currentUser(c *gin.Context) (uuid.UUID, error) {
	raw := auth.GetUserID(c)
	if raw == "" {
		return uuid.Nil, response.NewUnauthorizedError("用户未认证")
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, response.NewError(response.CodeBadRequest, "无效的用户ID")
	}
	return id, nil
}
