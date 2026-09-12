import request from './request';
import type { PageResponse } from '@/types/api';

export type AnnouncementStatus = 0 | 1 | 2;
export type AnnouncementCategory = 'association' | 'activity' | 'system' | 'personnel';
export type AnnouncementContentType = 'markdown' | 'html';

export interface Announcement {
  id: string;
  title: string;
  content: string;
  content_type: AnnouncementContentType;
  category: AnnouncementCategory;
  pinned: boolean;
  required: boolean;
  status: AnnouncementStatus;
  scheduled_at?: string;
  published_at?: string;
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
  pinned?: boolean;
  required?: boolean;
  scheduled_at?: string;
}

export interface UpdateAnnouncementParams {
  title?: string;
  content?: string;
  content_type?: AnnouncementContentType;
  category?: AnnouncementCategory;
  required?: boolean;
  scheduled_at?: string;
  clear_scheduled_at?: boolean;
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
  readers: { user: { id: string; name: string }; read_at: string }[];
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

export function pinAnnouncement(id: string, pinned?: boolean): Promise<Announcement> {
  return request.post(`/announcements/${id}/pin`, pinned === undefined ? {} : { pinned });
}

export function archiveAnnouncement(id: string): Promise<Announcement> {
  return request.post(`/announcements/${id}/archive`);
}

export function markAnnouncementRead(id: string): Promise<void> {
  return request.post(`/announcements/${id}/read`);
}

export function getAnnouncementUnreadCount(): Promise<AnnouncementUnreadCount> {
  return request.get('/announcements/unread-count');
}

export function getAnnouncementReadStatus(id: string): Promise<AnnouncementReadStatus> {
  return request.get(`/announcements/${id}/read-status`);
}
