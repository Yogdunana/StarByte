import request from './request';
import type {
  Notification,
  NotificationTemplate,
  ListNotificationParams,
  CreateNotificationTemplateParams,
  UpdateNotificationTemplateParams,
  TestTemplateParams,
  TestTemplateResult,
  SendNotificationParams,
  BroadcastNotificationParams,
  ListTemplateParams,
  PageResponse,
} from '@/types/api';

// ============================================================
// 用户通知
// ============================================================

/** 获取通知列表（分页） */
export function getNotificationList(
  params: ListNotificationParams,
): Promise<PageResponse<Notification>> {
  return request.get('/notifications', { params });
}

/** 获取未读数量 */
export function getUnreadCount(): Promise<{ count: number }> {
  return request.get('/notifications/unread/count');
}

/** 标记单条通知为已读 */
export function markAsRead(id: string): Promise<void> {
  return request.post(`/notifications/${id}/read`);
}

/** 全部标记为已读 */
export function markAllAsRead(category?: string): Promise<void> {
  return request.post('/notifications/read-all', { category });
}

/** 删除通知 */
export function deleteNotification(id: string): Promise<void> {
  return request.delete(`/notifications/${id}`);
}

// ============================================================
// 管理员通知操作
// ============================================================

/** 通过模板向指定用户发送通知 */
export function sendNotification(
  data: SendNotificationParams,
): Promise<void> {
  return request.post('/system/notifications/send', data);
}

/** 向所有在线用户广播通知 */
export function broadcastNotification(
  data: BroadcastNotificationParams,
): Promise<void> {
  return request.post('/system/notifications/broadcast', data);
}

// ============================================================
// 通知模板管理
// ============================================================

/** 获取模板列表（分页） */
export function getNotificationTemplateList(
  params: ListTemplateParams,
): Promise<PageResponse<NotificationTemplate>> {
  return request.get('/notification-templates', { params });
}

/** 获取模板详情 */
export function getNotificationTemplateDetail(
  id: string,
): Promise<NotificationTemplate> {
  return request.get(`/notification-templates/${id}`);
}

/** 创建模板 */
export function createNotificationTemplate(
  data: CreateNotificationTemplateParams,
): Promise<NotificationTemplate> {
  return request.post('/notification-templates', data);
}

/** 更新模板 */
export function updateNotificationTemplate(
  id: string,
  data: UpdateNotificationTemplateParams,
): Promise<NotificationTemplate> {
  return request.put(`/notification-templates/${id}`, data);
}

/** 删除模板 */
export function deleteNotificationTemplate(id: string): Promise<void> {
  return request.delete(`/notification-templates/${id}`);
}

/** 测试模板渲染 */
export function testNotificationTemplate(
  id: string,
  data: TestTemplateParams,
): Promise<TestTemplateResult> {
  return request.post(`/notification-templates/${id}/test`, data);
}
