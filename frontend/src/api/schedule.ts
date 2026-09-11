import request from './request';
import type { PageResponse } from '@/types/api';

export interface SchedulePerson {
  id: string;
  name: string;
}

export interface CalendarItem {
  id: string;
  name: string;
  description: string;
  calendar_type: number;
  source?: 'personal' | 'timetable' | 'import' | 'google' | string;
  source_key?: string;
  color: string;
  owner: SchedulePerson;
  department_id?: string;
  department_name?: string;
  member_role: number;
  can_edit: boolean;
  created_at: string;
  updated_at: string;
}

export interface ImportResult {
  calendar_id: string;
  calendar?: CalendarItem;
  event_count: number;
  replaced: boolean;
  source: string;
}

export interface GoogleStatus {
  configured: boolean;
  connected: boolean;
  email?: string;
  calendar_id?: string;
}

export interface ScheduleEvent {
  id: string;
  calendar_id: string;
  calendar_name: string;
  calendar_color: string;
  title: string;
  description: string;
  location: string;
  start_at: string;
  end_at: string;
  all_day: boolean;
  color: string;
  status: number;
  recurrence: string;
  recurrence_until?: string;
  meeting_id?: string;
  creator: SchedulePerson;
  attendee_count: number;
  occurrence_start?: string;
  can_edit: boolean;
  reminders?: Array<{ id: string; minutes_before: number }>;
}

export interface CreateCalendarPayload {
  name: string;
  description?: string;
  calendar_type: number;
  color?: string;
}

export interface CreateEventPayload {
  calendar_id?: string;
  title: string;
  description?: string;
  location?: string;
  start_at: string;
  end_at: string;
  all_day?: boolean;
  color?: string;
  recurrence?: string;
  remind_minutes?: number[];
}

export function listCalendars(params?: Record<string, unknown>): Promise<PageResponse<CalendarItem>> {
  return request.get('/schedules/calendars', { params });
}

export function createCalendar(data: CreateCalendarPayload): Promise<CalendarItem> {
  return request.post('/schedules/calendars', data);
}

export function updateCalendar(id: string, data: Partial<CreateCalendarPayload>): Promise<CalendarItem> {
  return request.put(`/schedules/calendars/${id}`, data);
}

export function deleteCalendar(id: string): Promise<void> {
  return request.delete(`/schedules/calendars/${id}`);
}

export function listEvents(params?: Record<string, unknown>): Promise<PageResponse<ScheduleEvent>> {
  return request.get('/schedules/events', { params });
}

export function rangeEvents(params: { start: string; end: string; calendar_id?: string }): Promise<ScheduleEvent[]> {
  return request.get('/schedules/events/range', { params });
}

export function createEvent(data: CreateEventPayload): Promise<ScheduleEvent> {
  return request.post('/schedules/events', data);
}

export function updateEvent(id: string, data: Partial<CreateEventPayload>): Promise<ScheduleEvent> {
  return request.put(`/schedules/events/${id}`, data);
}

export function deleteEvent(id: string): Promise<void> {
  return request.delete(`/schedules/events/${id}`);
}

export function setEventReminders(id: string, minutes: number[]): Promise<unknown> {
  return request.post(`/schedules/events/${id}/remind`, { minutes });
}

function multipartHeaders(): { 'Content-Type': undefined } {
  return { 'Content-Type': undefined };
}

export function importTimetable(file: File, semesterStart: string): Promise<ImportResult> {
  const form = new FormData();
  form.append('file', file);
  form.append('semester_start', semesterStart);
  return request.post('/schedules/imports/timetable', form, { headers: multipartHeaders() });
}

export function importICS(file: File, calendarId?: string): Promise<ImportResult> {
  const form = new FormData();
  form.append('file', file);
  if (calendarId) form.append('calendar_id', calendarId);
  return request.post('/schedules/imports/ics', form, { headers: multipartHeaders() });
}

export function googleStatus(): Promise<GoogleStatus> {
  return request.get('/schedules/google/status');
}

export function googleConnect(): Promise<{ auth_url: string }> {
  return request.get('/schedules/google/connect');
}

export function googleDisconnect(): Promise<void> {
  return request.post('/schedules/google/disconnect');
}

export function googleSync(): Promise<ImportResult> {
  return request.post('/schedules/google/sync');
}
