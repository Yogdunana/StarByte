import request from './request';
import type { PageResponse } from '@/types/api';

export interface Activity {
  id: string;
  title: string;
  description: string;
  cover_image_id?: string;
  category: string;
  tags: string[];
  start_time: string;
  end_time: string;
  location: string;
  latitude?: number;
  longitude?: number;
  checkin_radius_m?: number;
  gps_enabled: boolean;
  max_participants: number;
  status: ActivityStatus;
  organizer: { id: string; name: string };
  registered_count: number;
  checked_in_count: number;
  created_at: string;
  updated_at: string;
}

export type ActivityStatus = 0 | 1 | 2 | 3 | 4;

export interface Registration {
  id: string;
  activity_id: string;
  user: { id: string; name: string };
  status: RegistrationStatus;
  checkin_status: CheckinStatus;
  checked_in_at?: string;
  checkin_method?: number;
  created_at: string;
}

export type RegistrationStatus = 0 | 1 | 2 | 3 | 4;
export type CheckinStatus = 0 | 1;

export interface ActivityStats {
  activity_id: string;
  max_participants: number;
  registered_count: number;
  approved_count: number;
  waitlist_count: number;
  checked_in_count: number;
  register_rate: number;
  attend_rate: number;
  survey_count: number;
  avg_rating: number;
  rating_distribution: Record<string, number>;
}

export interface ActivityQRCode {
  activity_id: string;
  token: string;
  expires_at: string;
  checkin_path: string;
  png_base64: string;
}

export interface CreateActivityParams {
  title: string;
  description?: string;
  cover_image_id?: string;
  category?: string;
  tags?: string[];
  start_time: string;
  end_time: string;
  location?: string;
  latitude?: number;
  longitude?: number;
  checkin_radius_m?: number;
  max_participants?: number;
}

export interface UpdateActivityParams {
  title?: string;
  description?: string;
  cover_image_id?: string;
  category?: string;
  tags?: string[];
  start_time?: string;
  end_time?: string;
  location?: string;
  latitude?: number;
  longitude?: number;
  checkin_radius_m?: number;
  clear_geo?: boolean;
  max_participants?: number;
}

export interface ListActivityParams {
  page?: number;
  page_size?: number;
  status?: number;
  category?: string;
  keyword?: string;
  start_date?: string;
  end_date?: string;
}

export interface CheckinParams {
  method: 1 | 2;
  token?: string;
  latitude?: number;
  longitude?: number;
}

export interface SurveyParams {
  rating: number;
  comment?: string;
}

export function getActivityList(params: ListActivityParams): Promise<PageResponse<Activity>> {
  return request.get('/activities', { params });
}

export function getActivityDetail(id: string): Promise<Activity> {
  return request.get(`/activities/${id}`);
}

export function createActivity(data: CreateActivityParams): Promise<Activity> {
  return request.post('/activities', data);
}

export function updateActivity(id: string, data: UpdateActivityParams): Promise<Activity> {
  return request.put(`/activities/${id}`, data);
}

export function deleteActivity(id: string): Promise<void> {
  return request.delete(`/activities/${id}`);
}

export function startActivity(id: string): Promise<Activity> {
  return request.post(`/activities/${id}/start`);
}

export function endActivity(id: string): Promise<Activity> {
  return request.post(`/activities/${id}/end`);
}

export function cancelActivity(id: string, reason?: string): Promise<Activity> {
  return request.post(`/activities/${id}/cancel`, { reason });
}

export function registerActivity(id: string): Promise<Registration> {
  return request.post(`/activities/${id}/register`);
}

export function getMyRegistration(id: string): Promise<Registration | null> {
  return request.get(`/activities/${id}/register`);
}

export function cancelRegistration(id: string): Promise<void> {
  return request.delete(`/activities/${id}/register`);
}

export function getRegistrations(id: string): Promise<Registration[]> {
  return request.get(`/activities/${id}/registrations`);
}

export function approveRegistration(id: string, uid: string, approve: boolean, reason?: string): Promise<Registration> {
  return request.post(`/activities/${id}/registrations/${uid}/approve`, { approve, reason });
}

export function checkinActivity(id: string, data: CheckinParams): Promise<Registration> {
  return request.post(`/activities/${id}/checkin`, data);
}

export function getActivityQRCode(id: string): Promise<ActivityQRCode> {
  return request.get(`/activities/${id}/qrcode`);
}

export function getActivityStats(id: string): Promise<ActivityStats> {
  return request.get(`/activities/${id}/stats`);
}

export function submitSurvey(id: string, data: SurveyParams): Promise<void> {
  return request.post(`/activities/${id}/survey`, data);
}
