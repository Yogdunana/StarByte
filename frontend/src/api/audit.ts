import request from './request';
import { downloadBlob } from '@/utils/download';
import type { PageResponse } from '@/types/api';
import type { AxiosResponse } from 'axios';

export interface AuditUser {
  id: string;
  username: string;
  real_name: string;
}

export interface AuditLogItem {
  id: string;
  user: AuditUser;
  method: 'POST' | 'PUT' | 'DELETE' | string;
  path: string;
  module: string;
  action: 'CREATE' | 'UPDATE' | 'DELETE' | 'LOGIN' | 'LOGOUT' | string;
  request_body: string;
  response_code: number;
  ip_address: string;
  user_agent: string;
  duration_ms: number;
  timestamp: string;
}

export interface AuditQueryParams {
  page?: number;
  page_size?: number;
  start_time?: string;
  end_time?: string;
  user_id?: string;
  username?: string;
  action?: string;
  module?: string;
  keyword?: string;
  ip_address?: string;
  method?: string;
}

export interface ExportAuditLogParams extends AuditQueryParams {
  format: 'csv' | 'excel';
}

export interface ArchiveResponse {
  archive_id: string;
  record_count: number;
  archive_date: string;
  minio_object: string;
  status: number;
  message: string;
}

export function getAuditLogList(
  params: AuditQueryParams,
): Promise<PageResponse<AuditLogItem>> {
  return request.get('/system/audit-logs', { params });
}

export function getAuditLogDetail(id: string): Promise<AuditLogItem> {
  return request.get(`/system/audit-logs/${id}`);
}

export async function exportAuditLogs(params: ExportAuditLogParams): Promise<void> {
  const response: AxiosResponse = await request.get('/system/audit-logs/export', {
    params,
    responseType: 'blob',
  });
  const blob = response.data as Blob;
  const disposition = response.headers['content-disposition'];
  let filename = params.format === 'excel' ? 'audit_logs.xlsx' : 'audit_logs.csv';
  if (disposition) {
    const match = disposition.match(/filename\*?=(?:UTF-8'')?["']?([^"';\s]+)["']?/i);
    if (match) {
      filename = decodeURIComponent(match[1]);
    }
  }
  downloadBlob(blob, filename);
}

export function triggerArchive(beforeDays?: number): Promise<ArchiveResponse> {
  return request.post('/system/audit-logs/archive', { before_days: beforeDays });
}
