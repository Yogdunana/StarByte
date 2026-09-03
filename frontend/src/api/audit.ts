import request from './request';
import { downloadBlob } from '@/utils/download';
import type { PageResponse } from '@/types/api';
import type { AxiosResponse } from 'axios';

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

/** 导出审计日志（通过 axios 实例，享受拦截器：token 刷新、错误处理） */
export async function exportAuditLogs(
  params: ExportAuditLogParams,
): Promise<void> {
  const response: AxiosResponse = await request.get('/audit-logs/export', {
    params,
    responseType: 'blob',
  });

  const blob = response.data as Blob;
  const disposition = response.headers['content-disposition'];
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
