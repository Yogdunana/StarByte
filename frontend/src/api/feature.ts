import request from './request';
import type { PageResponse } from '@/types/api';

export type FeatureType = 'boolean' | 'user_allowlist' | 'role_dept' | 'percentage';

export interface FeatureRules {
  user_ids?: string[];
  role_codes?: string[];
  department_ids?: string[];
  percent?: number;
  salt?: string;
}

export interface FeatureFlag {
  id: string;
  flag_key: string;
  name: string;
  description: string;
  flag_type: FeatureType;
  enabled: boolean;
  group_name: string;
  priority: number;
  rules: FeatureRules;
  is_system: boolean;
  created_at: string;
  updated_at: string;
}

export interface FeatureEvaluate {
  key: string;
  enabled: boolean;
  reason: string;
  flag_type?: string;
}

export interface FeatureAudit {
  id: string;
  flag_id?: string;
  flag_key: string;
  action: string;
  actor_id?: string;
  before_json?: unknown;
  after_json?: unknown;
  reason: string;
  created_at: string;
}

export interface CreateFeaturePayload {
  flag_key: string;
  name: string;
  description?: string;
  flag_type: FeatureType;
  enabled?: boolean;
  group_name?: string;
  priority?: number;
  rules?: FeatureRules;
}

export function listFeatures(params?: {
  page?: number;
  page_size?: number;
  keyword?: string;
  group?: string;
  enabled?: boolean;
}): Promise<PageResponse<FeatureFlag>> {
  return request.get('/system/features', { params });
}

export function getFeature(id: string): Promise<FeatureFlag> {
  return request.get(`/system/features/${id}`);
}

export function createFeature(data: CreateFeaturePayload): Promise<FeatureFlag> {
  return request.post('/system/features', data);
}

export function updateFeature(id: string, data: Partial<CreateFeaturePayload>): Promise<FeatureFlag> {
  return request.put(`/system/features/${id}`, data);
}

export function toggleFeature(id: string, enabled?: boolean, reason?: string): Promise<FeatureFlag> {
  return request.post(`/system/features/${id}/toggle`, { enabled, reason });
}

export function evaluateFeature(id: string, userId?: string): Promise<FeatureEvaluate> {
  return request.get(`/system/features/${id}/evaluate`, { params: { user_id: userId } });
}

export function listFeatureAudits(params?: {
  page?: number; page_size?: number; flag_key?: string;
}): Promise<PageResponse<FeatureAudit>> {
  return request.get('/system/features/audit', { params });
}

export function evaluateMyFeatures(keys?: string[]): Promise<Record<string, FeatureEvaluate>> {
  return request.get('/features/me', { params: keys?.length ? { keys: keys.join(',') } : undefined });
}

export interface CmsPage {
  id: string;
  name: string;
  description: string;
  updated_at: string;
}

export function listCmsPages(): Promise<CmsPage[]> {
  return request.get('/cms/pages');
}

export interface MembershipPortal {
  applications: Array<{ id: string; real_name: string; status: number; submitted_at: string }>;
  hint: string;
}

export function getMembershipPortal(): Promise<MembershipPortal> {
  return request.get('/member/portal');
}

export const DEFAULT_FEATURE_KEYS = ['cms.public', 'announcement.feed', 'membership.portal'] as const;
