import request from './request';
import { getToken } from '@/utils/storage';
import { downloadBlob } from '@/utils/download';
import type { PageResponse } from '@/types/api';

// ========== 类型定义 ==========

export interface AuditLogListItem {
  id: string;
  user_id: string;
  username: string;
  operation: string;
  method: string;
  path: string;
  ip: string;
  response_status: number;
  duration_ms: number;
  request_id: string;
  created_at: string;
}

export interface AuditLogDetail {
  id: string;
  user_id: string;
  username: string;
  operation: string;
  method: string;
  path: string;
  ip: string;
  user_agent: string;
  request_params: string;
  response_status: number;
  response_body: string;
  duration_ms: number;
  request_id: string;
  is_archived: boolean;
  created_at: string;
}

export interface ListAuditLogParams {
  page?: number;
  page_size?: number;
  username?: string;
  operation?: string;
  method?: string;
  path?: string;
  ip?: string;
  request_id?: string;
  status_min?: number;
  status_max?: number;
  start_time?: string;
  end_time?: string;
}

export interface ExportAuditLogParams {
  format: 'csv' | 'json';
  username?: string;
  operation?: string;
  method?: string;
  path?: string;
  ip?: string;
  status_min?: number;
  status_max?: number;
  start_time?: string;
  end_time?: string;
}

export interface ArchiveResponse {
  archive_id: string;
  record_count: number;
  archive_date: string;
  minio_object: string;
  status: number;
  message: string;
}

// ========== API 请求函数 ==========

/** 获取审计日志列表 */
export function getAuditLogList(
  params: ListAuditLogParams,
): Promise<PageResponse<AuditLogListItem>> {
  return request.get('/audit-logs', { params });
}

/** 获取审计日志详情 */
export function getAuditLogDetail(id: string): Promise<AuditLogDetail> {
  return request.get(`/audit-logs/${id}`);
}

/** 导出审计日志 */
export async function exportAuditLogs(
  params: ExportAuditLogParams,
): Promise<void> {
  const baseURL = import.meta.env.VITE_API_BASE_URL || '/api/v1';
  const queryParams = new URLSearchParams();
  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined && value !== null && value !== '') {
      queryParams.append(key, String(value));
    }
  });

  const url = `${baseURL}/audit-logs/export?${queryParams.toString()}`;
  const token = getToken();
  const headers: Record<string, string> = {};
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const response = await fetch(url, { headers });
  if (!response.ok) {
    throw new Error(`导出失败: ${response.status}`);
  }

  const blob = await response.blob();
  const disposition = response.headers.get('Content-Disposition');
  let filename = `audit_logs.${params.format}`;
  if (disposition) {
    const match = disposition.match(/filename\*?=(?:UTF-8'')?["']?([^"';\s]+)["']?/i);
    if (match) {
      filename = decodeURIComponent(match[1]);
    }
  }
  downloadBlob(blob, filename);
}

/** 手动触发归档 */
export function triggerArchive(beforeDays?: number): Promise<ArchiveResponse> {
  return request.post('/audit-logs/archive', { before_days: beforeDays });
}
