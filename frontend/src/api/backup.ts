import request from './request';
import type { PageResponse } from '@/types/api';

export interface BackupRecord {
  id: string;
  trigger_source: string;
  status: number;
  storage: string;
  object_key: string;
  filename: string;
  checksum_sha256: string;
  size_bytes: number;
  encrypted: boolean;
  started_at?: string | null;
  finished_at?: string | null;
  error_message: string;
  created_by?: string;
  created_at: string;
  updated_at: string;
}

export interface BackupPolicy {
  enabled: boolean;
  retention_days: number;
  cron_expr: string;
  timezone: string;
  updated_at?: string;
}

export interface BackupStorageStats {
  count: number;
  size_bytes: number;
  prefix: string;
  bucket: string;
  local_path?: string;
  compression?: string;
  encryption_enabled?: boolean;
  incremental_enabled?: boolean;
  pitr_enabled?: boolean;
}

export interface BackupPreview {
  id: string;
  filename: string;
  size_bytes: number;
  checksum_ok: boolean;
  encrypted: boolean;
  decrypt_ok: boolean;
  gzip_ok: boolean;
  toc_valid: boolean;
  toc?: string;
  ready: boolean;
  compression: string;
  encryption_configured: boolean;
  error?: string;
}

export function getBackups(params: {
  page?: number;
  page_size?: number;
  status?: number;
}): Promise<PageResponse<BackupRecord>> {
  return request.get('/system/backups', { params });
}

export function getBackup(id: string): Promise<BackupRecord> {
  return request.get(`/system/backups/${id}`);
}

export function createBackup(): Promise<BackupRecord> {
  return request.post('/system/backups', {});
}

export function deleteBackup(id: string): Promise<void> {
  return request.delete(`/system/backups/${id}`);
}

export function restoreBackup(id: string, confirmation = 'RESTORE'): Promise<BackupRecord> {
  return request.post(`/system/backups/${id}/restore`, {
    confirm: true,
    confirmation,
  });
}

export function getBackupPolicy(): Promise<BackupPolicy> {
  return request.get('/system/backups/policies');
}

export function updateBackupPolicy(data: Partial<BackupPolicy>): Promise<BackupPolicy> {
  return request.put('/system/backups/policies', data);
}

export function getBackupStorage(): Promise<BackupStorageStats> {
  return request.get('/system/backups/storage');
}

export function previewBackup(id: string, signal?: AbortSignal): Promise<BackupPreview> {
  return request.get(`/system/backups/${id}/preview`, { timeout: 180000, signal });
}

export interface BackupDrillResult {
  id: string;
  filename: string;
  queued?: boolean;
  ready: boolean;
  restored: boolean;
  status?: string;
  target_host: string;
  target_port: number;
  target_dbname: string;
  error?: string;
}

export function drillRestoreBackup(
  id: string,
  data: {
    target_dbname: string;
    target_dsn?: string;
    target_password?: string;
    confirmation?: string;
  },
): Promise<BackupDrillResult> {
  return request.post(`/system/backups/${id}/restore-drill`, {
    confirm: true,
    confirmation: data.confirmation || 'DRILL',
    target_dbname: data.target_dbname,
    target_dsn: data.target_dsn,
    target_password: data.target_password,
  });
}

export function getDrillRestore(id: string): Promise<BackupDrillResult> {
  return request.get(`/system/backups/${id}/restore-drill`);
}

export function drillIsPending(out: BackupDrillResult | null | undefined): boolean {
  if (!out) return false;
  if (out.restored || out.error || out.status === 'restored' || out.status === 'failed') {
    return false;
  }
  return !!out.queued || out.status === 'queued' || out.status === 'running';
}
