import request from './request';
import { downloadBlob } from '@/utils/download';
import type { PageResponse } from '@/types/api';
import type { AxiosResponse } from 'axios';

export interface AuditUser {
  id: string;
  username: string;
  real_name: string;
}

export interface AuditFieldChange {
  path: string;
  before: unknown;
  after: unknown;
}

export interface AuditLogItem {
  id: string;
  user: AuditUser;
  method: 'POST' | 'PUT' | 'DELETE' | string;
  path: string;
  module: string;
  action: 'CREATE' | 'UPDATE' | 'DELETE' | 'LOGIN' | 'LOGOUT' | 'EXPORT' | string;
  request_body: string;
  response_code: number;
  ip_address: string;
  user_agent: string;
  duration_ms: number;
  timestamp: string;
  entity_type?: string;
  entity_id?: string;
  before_json?: string;
  after_json?: string;
  diff?: AuditFieldChange[];
  compliance_flags?: string[];
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

export interface AuditTraceItem {
  id: string;
  user: AuditUser;
  action: string;
  method: string;
  path: string;
  module: string;
  entity_type: string;
  entity_id: string;
  before_json: string;
  after_json: string;
  diff: AuditFieldChange[];
  compliance_flags: string[];
  timestamp: string;
}

export interface AuditCountItem {
  key: string;
  count: number;
}

export interface AuditReport {
  start_time: string;
  end_time: string;
  total: number;
  by_action: AuditCountItem[];
  by_module: AuditCountItem[];
  by_compliance: AuditCountItem[];
  top_operators: AuditCountItem[];
  note: string;
}

export interface AuditArchiveItem {
  id: string;
  archive_date: string;
  record_count: number;
  minio_object: string;
  status: number;
  created_at: string;
}

export interface AuditArchivePull {
  archive: AuditArchiveItem;
  list: AuditLogItem[];
  total: number;
  page: number;
  page_size: number;
  truncated: boolean;
}

function filenameFromDisposition(disposition: string | undefined, fallback: string): string {
  if (!disposition) return fallback;
  const match = disposition.match(/filename\*?=(?:UTF-8'')?["']?([^"';\s]+)["']?/i);
  if (!match) return fallback;
  return decodeURIComponent(match[1]);
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
  downloadBlob(
    response.data as Blob,
    filenameFromDisposition(
      response.headers['content-disposition'] as string | undefined,
      params.format === 'excel' ? 'audit_logs.xlsx' : 'audit_logs.csv',
    ),
  );
}

export function triggerArchive(beforeDays?: number): Promise<ArchiveResponse> {
  return request.post('/system/audit-logs/archive', { before_days: beforeDays });
}

export function getAuditTrace(
  entityType: string,
  entityId: string,
  params?: { page?: number; page_size?: number },
): Promise<PageResponse<AuditTraceItem>> {
  return request.get(`/system/audit-logs/traces/${entityType}/${entityId}`, { params });
}

export function getAuditReport(params: {
  start_time?: string;
  end_time?: string;
  module?: string;
  format?: string;
}): Promise<AuditReport> {
  return request.get('/system/audit-logs/reports', { params: { ...params, format: 'json' } });
}

export async function downloadAuditReport(params: {
  start_time?: string;
  end_time?: string;
  module?: string;
  format: 'csv' | 'pdf' | 'excel';
}): Promise<void> {
  const response: AxiosResponse = await request.get('/system/audit-logs/reports', {
    params,
    responseType: 'blob',
  });
  const fallback =
    params.format === 'excel'
      ? 'audit_compliance_report.xlsx'
      : params.format === 'pdf'
        ? 'audit_compliance_report.pdf'
        : 'audit_compliance_report.csv';
  downloadBlob(
    response.data as Blob,
    filenameFromDisposition(response.headers['content-disposition'] as string | undefined, fallback),
  );
}

export function getAuditArchives(params: {
  page?: number;
  page_size?: number;
}): Promise<PageResponse<AuditArchiveItem>> {
  return request.get('/system/audit-logs/archives', { params });
}

export function pullAuditArchive(params: {
  id: string;
  page?: number;
  page_size?: number;
  keyword?: string;
}): Promise<AuditArchivePull> {
  return request.get('/system/audit-logs/archives', { params });
}
