package handler

import (
	"github.com/Yogdunana/StarByte/backend/internal/notification/dto"
	"github.com/Yogdunana/StarByte/backend/internal/notification/model"
	"github.com/Yogdunana/StarByte/backend/internal/notification/service"
	authmiddleware "github.com/Yogdunana/StarByte/backend/pkg/middleware/auth"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// NotificationHandler 通知相关 HTTP 处理器
type NotificationHandler struct {
	notificationService service.NotificationService
	hubManager          service.HubManager
}

// NewNotificationHandler 创建通知处理器
func NewNotificationHandler(notificationService service.NotificationService, hubManager service.HubManager) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
		hubManager:          hubManager,
	}
}

// List GET /api/v1/notifications
// @Summary 通知列表（分页）
// @Description 获取当前用户的通知列表，支持分类筛选和未读筛选
// @Tags 通知
// @Produce json
// @Security Bearer
// @Param page query int false "页码"
// @Param page_size query int false "每页条数"
// @Param category query string false "分类"
// @Param unread_only query bool false "仅未读"
// @Success 200 {object} response.Response
// @Router /notifications [get]
func (h *NotificationHandler) List(c *gin.Context) {
	userID := getUserID(c)

	var req dto.ListNotificationsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	list, total, err := h.notificationService.ListByUser(c.Request.Context(), userID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	result := make([]*dto.NotificationResponse, 0, len(list))
	for _, n := range list {
		result = append(result, notificationToResponse(n))
	}

	response.Page(c, result, total, req.Page, req.PageSize)
}

// UnreadCount GET /api/v1/notifications/unread/count
// @Summary 未读通知计数
// @Description 获取当前用户的未读通知数量
// @Tags 通知
// @Produce json
// @Security Bearer
// @Success 200 {object} response.Response
// @Router /notifications/unread/count [get]
func (h *NotificationHandler) UnreadCount(c *gin.Context) {
	userID := getUserID(c)

	count, err := h.notificationService.GetUnreadCount(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, &dto.UnreadCountResponse{Count: count})
}

// MarkAsRead POST /api/v1/notifications/:id/read
// @Summary 标记通知已读
// @Description 标记指定通知为已读
// @Tags 通知
// @Produce json
// @Security Bearer
// @Param id path string true "通知 ID"
// @Success 200 {object} response.Response
// @Router /notifications/{id}/read [post]
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userID := getUserID(c)

	idStr := c.Param("id")
	notificationID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "无效的通知 ID")
		return
	}

	if err := h.notificationService.MarkAsRead(c.Request.Context(), userID, []uuid.UUID{notificationID}); err != nil {
		response.Error(c, err)
		return
	}

	response.OKWithoutData(c)
}

// MarkAllAsRead POST /api/v1/notifications/read-all
// @Summary 全部已读
// @Description 标记当前用户的所有通知为已读
// @Tags 通知
// @Produce json
// @Security Bearer
// @Param category body dto.MarkAllReadRequest false "按分类标记已读（可选）"
// @Success 200 {object} response.Response
// @Router /notifications/read-all [post]
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID := getUserID(c)

	var req dto.MarkAllReadRequest
	_ = c.ShouldBindJSON(&req)

	if err := h.notificationService.MarkAllAsRead(c.Request.Context(), userID, req.Category); err != nil {
		response.Error(c, err)
		return
	}

	response.OKWithoutData(c)
}

// Delete DELETE /api/v1/notifications/:id
// @Summary 删除通知
// @Description 删除指定通知
// @Tags 通知
// @Produce json
// @Security Bearer
// @Param id path string true "通知 ID"
// @Success 200 {object} response.Response
// @Router /notifications/{id} [delete]
func (h *NotificationHandler) Delete(c *gin.Context) {
	userID := getUserID(c)

	idStr := c.Param("id")
	notificationID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "无效的通知 ID")
		return
	}

	if err := h.notificationService.Delete(c.Request.Context(), userID, notificationID); err != nil {
		response.Error(c, err)
		return
	}

	response.OKWithoutData(c)
}

// Send POST /api/v1/system/notifications/send
// @Summary 发送通知（管理员）
// @Description 通过模板向指定用户发送通知
// @Tags 系统管理-通知
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body dto.SendNotificationRequest true "发送参数"
// @Success 200 {object} response.Response
// @Router /system/notifications/send [post]
func (h *NotificationHandler) Send(c *gin.Context) {
	var req dto.SendNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := h.notificationService.Send(c.Request.Context(), &req); err != nil {
		response.Error(c, err)
		return
	}

	response.OKWithoutData(c)
}

// Broadcast POST /api/v1/system/notifications/broadcast
// @Summary 广播通知（管理员）
// @Description 向所有在线用户广播通知
// @Tags 系统管理-通知
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body dto.BroadcastNotificationRequest true "广播参数"
// @Success 200 {object} response.Response
// @Router /system/notifications/broadcast [post]
func (h *NotificationHandler) Broadcast(c *gin.Context) {
	var req dto.BroadcastNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	onlineUserIDs := h.hubManager.GetOnlineUsers()
	if len(onlineUserIDs) == 0 {
		response.OKWithoutData(c)
		return
	}

	if err := h.notificationService.Broadcast(c.Request.Context(), &req, onlineUserIDs); err != nil {
		response.Error(c, err)
		return
	}

	response.OKWithoutData(c)
}

// getUserID 从上下文获取用户 ID
func getUserID(c *gin.Context) uuid.UUID {
	userIDStr := authmiddleware.GetUserID(c)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil
	}
	return userID
}

// notificationToResponse 将通知模型转为响应 DTO
func notificationToResponse(n *model.Notification) *dto.NotificationResponse {
	resp := &dto.NotificationResponse{
		ID:        n.ID.String(),
		Title:     n.Title,
		Content:   n.Content,
		Category:  n.Category,
		Priority:  n.Priority,
		IsRead:    n.IsRead,
		ActionURL: n.ActionURL,
		CreatedAt: n.CreatedAt,
	}

	if n.SenderID != nil {
		resp.Sender = dto.SenderInfo{
			ID:   n.SenderID.String(),
			Name: n.SenderName,
		}
	}

	return resp
}
