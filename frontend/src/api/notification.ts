import request from './request';
import type {
  Notification,
  NotificationTemplate,
  ListNotificationParams,
  CreateNotificationTemplateParams,
  UpdateNotificationTemplateParams,
  PageResponse,
} from '@/types/api';

// ============================================================
// 通知
// ============================================================

// 获取通知列表
export function getNotificationList(
  params: ListNotificationParams,
): Promise<PageResponse<Notification>> {
  return request.get('/notifications', { params });
}

// 获取未读数量
export function getUnreadCount(): Promise<{ count: number }> {
  return request.get('/notifications/unread-count');
}

// 获取通知详情
export function getNotificationDetail(id: string): Promise<Notification> {
  return request.get(`/notifications/${id}`);
}

// 标记为已读
export function markAsRead(id: string): Promise<void> {
  return request.post(`/notifications/${id}/read`);
}

// 全部标记为已读
export function markAllAsRead(): Promise<void> {
  return request.post('/notifications/read-all');
}

// 删除通知
export function deleteNotification(id: string): Promise<void> {
  return request.delete(`/notifications/${id}`);
}

// ============================================================
// 通知模板
// ============================================================

// 获取模板列表
export function getNotificationTemplateList(params: {
  page?: number;
  page_size?: number;
  type?: number;
  status?: number;
}): Promise<PageResponse<NotificationTemplate>> {
  return request.get('/notifications/templates', { params });
}

// 获取模板详情
export function getNotificationTemplateDetail(id: string): Promise<NotificationTemplate> {
  return request.get(`/notifications/templates/${id}`);
}

// 创建模板
export function createNotificationTemplate(
  data: CreateNotificationTemplateParams,
): Promise<NotificationTemplate> {
  return request.post('/notifications/templates', data);
}

// 更新模板
export function updateNotificationTemplate(
  id: string,
  data: UpdateNotificationTemplateParams,
): Promise<NotificationTemplate> {
  return request.put(`/notifications/templates/${id}`, data);
}

// 删除模板
export function deleteNotificationTemplate(id: string): Promise<void> {
  return request.delete(`/notifications/templates/${id}`);
}
