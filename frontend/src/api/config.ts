import request from './request';
import type {
  RuntimeConfig,
  CreateRuntimeConfigParams,
  UpdateRuntimeConfigParams,
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
