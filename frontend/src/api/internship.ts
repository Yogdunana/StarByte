import request from './request';
import type {
  Internship,
  InternshipConfig,
  InternshipDepartmentStats,
  InternshipDurationStats,
  InternshipRankingResponse,
  CreateInternshipParams,
  UpdateInternshipParams,
  ListInternshipParams,
  PageResponse,
} from '@/types/api';

export function getInternshipList(params: ListInternshipParams): Promise<PageResponse<Internship>> {
  return request.get('/internships', { params });
}

export function getMyInternships(params?: { status?: number }): Promise<Internship[]> {
  return request.get('/internships/my', { params });
}

export function getInternshipDetail(id: string): Promise<Internship> {
  return request.get(`/internships/${id}`);
}

export function createInternship(data: CreateInternshipParams): Promise<Internship> {
  return request.post('/internships', data);
}

export function updateInternship(id: string, data: UpdateInternshipParams): Promise<Internship> {
  return request.put(`/internships/${id}`, data);
}

export function deleteInternship(id: string): Promise<void> {
  return request.delete(`/internships/${id}`);
}

export function completeInternship(id: string, data: { report?: string; achievements?: string }): Promise<Internship> {
  return request.post(`/internships/${id}/complete`, data);
}

export function submitInternshipReport(id: string, report: string): Promise<Internship> {
  return request.post(`/internships/${id}/report`, { report });
}

export function commentInternship(id: string, mentorComment: string): Promise<Internship> {
  return request.post(`/internships/${id}/mentor-comment`, { mentor_comment: mentorComment });
}

export function getInternshipDurationStats(params?: {
  start_date?: string;
  end_date?: string;
  group_by?: 'user' | 'department' | 'month';
}): Promise<InternshipDurationStats> {
  return request.get('/internships/stats/duration', { params });
}

export function getInternshipRanking(params?: {
  department_id?: string;
  limit?: number;
  sort_by?: 'duration' | 'count';
}): Promise<InternshipRankingResponse> {
  return request.get('/internships/stats/ranking', { params });
}

export function getInternshipDepartmentStats(params?: {
  start_date?: string;
  end_date?: string;
}): Promise<InternshipDepartmentStats> {
  return request.get('/internships/stats/department', { params });
}

export function getInternshipConfig(): Promise<InternshipConfig> {
  return request.get('/system/internship-config');
}

export function updateInternshipConfig(data: Partial<InternshipConfig>): Promise<InternshipConfig> {
  return request.put('/system/internship-config', data);
}
