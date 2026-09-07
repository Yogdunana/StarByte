import request from './request';
import type { PageResponse } from '@/types/api';

export interface DisciplinePerson {
  id: string;
  name: string;
}

export interface DisciplineAppeal {
  id: string;
  reason: string;
  status: number;
  created_at: string;
}

export interface DisciplineRecord {
  id: string;
  user: DisciplinePerson;
  title: string;
  description: string;
  level: number;
  status: number;
  issued_at: string;
  flow_instance_id?: string;
  appeals?: DisciplineAppeal[];
}

export interface CreateDisciplineRecord {
  user_id: string;
  title: string;
  description?: string;
  level: number;
}

export function getDisciplineRecords(params: Record<string, unknown>): Promise<PageResponse<DisciplineRecord>> {
  return request.get('/discipline/records', { params });
}

export function getDisciplineRecord(id: string): Promise<DisciplineRecord> {
  return request.get(`/discipline/records/${id}`);
}

export function createDisciplineRecord(data: CreateDisciplineRecord): Promise<DisciplineRecord> {
  return request.post('/discipline/records', data);
}

export function updateDisciplineRecord(id: string, data: Partial<CreateDisciplineRecord>): Promise<DisciplineRecord> {
  return request.put(`/discipline/records/${id}`, data);
}

export function approveDiscipline(id: string, comment?: string): Promise<DisciplineRecord> {
  return request.post(`/discipline/records/${id}/approve`, { comment });
}

export function revokeDiscipline(id: string, reason: string): Promise<DisciplineRecord> {
  return request.post(`/discipline/records/${id}/revoke`, { reason });
}

export function appealDiscipline(id: string, reason: string): Promise<DisciplineRecord> {
  return request.post(`/discipline/records/${id}/appeal`, { reason });
}
