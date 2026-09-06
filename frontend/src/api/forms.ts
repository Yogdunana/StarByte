import request from './request';
import type { PageResponse } from '@/types/api';
import type { FormField } from '@/components/FormEngine';

export interface FormListItem {
  id: string;
  name: string;
  description: string;
  status: 0 | 1 | 2;
  submission_count: number;
  created_at: string;
  updated_at: string;
}

export interface FormDetail extends FormListItem {
  fields: FormField[];
}

export interface FormSubmitResult {
  submission_id: string;
  submitted_at: string;
}

export interface FormSubmissionItem {
  id: string;
  data: Record<string, unknown>;
  submitted_by: { id: string; name: string } | null;
  submitted_at: string;
}

export function listForms(params: { page?: number; page_size?: number; keyword?: string; status?: number }): Promise<PageResponse<FormListItem>> {
  return request.get('/forms', { params });
}

export function getForm(id: string): Promise<FormDetail> {
  return request.get(`/forms/${id}`);
}

export function createForm(body: { name: string; description?: string; fields: FormField[]; status?: number }): Promise<FormDetail> {
  return request.post('/forms', body);
}

export function updateForm(id: string, body: { name?: string; description?: string; fields?: FormField[]; status?: number }): Promise<FormDetail> {
  return request.put(`/forms/${id}`, body);
}

export function submitForm(id: string, data: Record<string, unknown>): Promise<FormSubmitResult> {
  return request.post(`/forms/${id}/submit`, data);
}

export function listFormSubmissions(id: string, params: { page?: number; page_size?: number }): Promise<PageResponse<FormSubmissionItem>> {
  return request.get(`/forms/${id}/submissions`, { params });
}
