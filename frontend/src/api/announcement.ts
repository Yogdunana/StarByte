import request from './request';
import type { PageResponse } from '@/types/api';

export type AnnouncementStatus = 0 | 1 | 2;
export type AnnouncementCategory = 'association' | 'activity' | 'system' | 'personnel';
export type AnnouncementContentType = 'markdown' | 'html';
export type AnnouncementAudience = 'all' | 'role' | 'department' | 'users';

export interface AnnouncementAttachment {
  file_id: string;
  name: string;
  size: number;
}

export interface Announcement {
  id: string;
  title: string;
  content: string;
  content_type: AnnouncementContentType;
  category: AnnouncementCategory;
  pinned: boolean;
  required: boolean;
  sort_order: number;
  status: AnnouncementStatus;
  scheduled_at?: string;
  expires_at?: string;
  published_at?: string;
  audience_type: AnnouncementAudience;
  audience_ids: string[];
  attachments: AnnouncementAttachment[];
  author: { id: string; name: string };
  is_read: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreateAnnouncementParams {
  title: string;
  content?: string;
  content_type?: AnnouncementContentType;
  category: AnnouncementCategory;
  required?: boolean;
  sort_order?: number;
  scheduled_at?: string;
  expires_at?: string;
  audience_type?: AnnouncementAudience;
  audience_ids?: string[];
  attachments?: AnnouncementAttachment[];
}

export interface UpdateAnnouncementParams {
  title?: string;
  content?: string;
  content_type?: AnnouncementContentType;
  category?: AnnouncementCategory;
  required?: boolean;
  sort_order?: number;
  scheduled_at?: string;
  clear_scheduled_at?: boolean;
  expires_at?: string;
  clear_expires_at?: boolean;
  audience_type?: AnnouncementAudience;
  audience_ids?: string[];
  attachments?: AnnouncementAttachment[];
}

export interface ListAnnouncementParams {
  page?: number;
  page_size?: number;
  status?: number;
  category?: string;
  keyword?: string;
  unread_only?: boolean;
  pinned_only?: boolean;
}

export interface AnnouncementUnreadCount {
  count: number;
}

export interface AnnouncementReadStatus {
  announcement_id: string;
  read_count: number;
  unread_count: number;
  read_rate: number;
  avg_duration_seconds: number;
  readers: { user: { id: string; name: string }; read_at: string; duration_seconds: number }[];
  unread_users: { id: string; name: string }[];
}

export function getAnnouncementList(params: ListAnnouncementParams): Promise<PageResponse<Announcement>> {
  return request.get('/announcements', { params });
}

export function getAnnouncementDetail(id: string): Promise<Announcement> {
  return request.get(`/announcements/${id}`);
}

export function createAnnouncement(data: CreateAnnouncementParams): Promise<Announcement> {
  return request.post('/announcements', data);
}

export function updateAnnouncement(id: string, data: UpdateAnnouncementParams): Promise<Announcement> {
  return request.put(`/announcements/${id}`, data);
}

export function deleteAnnouncement(id: string): Promise<void> {
  return request.delete(`/announcements/${id}`);
}

export function publishAnnouncement(id: string): Promise<Announcement> {
  return request.post(`/announcements/${id}/publish`);
}

export function pinAnnouncement(id: string, pinned?: boolean, sortOrder?: number): Promise<Announcement> {
  const body: { pinned?: boolean; sort_order?: number } = {};
  if (pinned !== undefined) body.pinned = pinned;
  if (sortOrder !== undefined) body.sort_order = sortOrder;
  return request.post(`/announcements/${id}/pin`, body);
}

export function archiveAnnouncement(id: string): Promise<Announcement> {
  return request.post(`/announcements/${id}/archive`);
}

export function markAnnouncementRead(id: string, durationSeconds?: number): Promise<void> {
  return request.post(`/announcements/${id}/read`, durationSeconds ? { duration_seconds: durationSeconds } : {});
}

export function getAnnouncementUnreadCount(): Promise<AnnouncementUnreadCount> {
  return request.get('/announcements/unread-count');
}

export function getAnnouncementReadStatus(id: string): Promise<AnnouncementReadStatus> {
  return request.get(`/announcements/${id}/read-status`);
}
