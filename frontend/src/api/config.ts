import request from './request';
import type {
  RuntimeConfig,
  CreateRuntimeConfigParams,
  UpdateRuntimeConfigParams,
  SMTPSettings,
  UpdateSMTPSettingsParams,
  TestSMTPParams,
  TestSMTPResult,
} from '@/types/api';

export function getRuntimeConfigs(params?: {
  category?: string;
  keyword?: string;
}): Promise<RuntimeConfig[]> {
  return request.get('/system/configs', { params });
}

export function getRuntimeConfigByKey(key: string): Promise<RuntimeConfig> {
  return request.get(`/system/configs/key/${encodeURIComponent(key)}`);
}

export function createRuntimeConfig(data: CreateRuntimeConfigParams): Promise<RuntimeConfig> {
  return request.post('/system/configs', data);
}

export function updateRuntimeConfig(id: string, data: UpdateRuntimeConfigParams): Promise<RuntimeConfig> {
  return request.put(`/system/configs/${id}`, data);
}

export function deleteRuntimeConfig(id: string): Promise<void> {
  return request.delete(`/system/configs/${id}`);
}

export function getSMTPSettings(): Promise<SMTPSettings> {
  return request.get('/system/smtp');
}

export function updateSMTPSettings(data: UpdateSMTPSettingsParams): Promise<SMTPSettings> {
  return request.put('/system/smtp', data);
}

export function testSMTPSettings(data: TestSMTPParams): Promise<TestSMTPResult> {
  return request.post('/system/smtp/test', data);
}
