import axios from 'axios';
import request from './request';
import type { PageResponse } from '@/types/api';

export type KnowledgeKind = 'page' | 'doc';
export type KnowledgeVisibility = 'public' | 'authenticated' | 'permission' | 'role';
export type KnowledgeStatus = 0 | 1;

export interface KnowledgePerson {
  id: string;
  name: string;
}

export interface KnowledgeAttachment {
  id: string;
  file_id: string;
  name: string;
  size: number;
  created_at: string;
}

export interface KnowledgeDoc {
  id: string;
  kind: KnowledgeKind;
  slug: string;
  path: string;
  title: string;
  summary: string;
  content?: string;
  category_id?: string;
  category_name?: string;
  visibility: KnowledgeVisibility;
  allowed_roles?: string[];
  permission_code?: string;
  status: KnowledgeStatus;
  version: number;
  author: KnowledgePerson;
  published_at?: string;
  created_at: string;
  updated_at: string;
  attachments?: KnowledgeAttachment[];
}

export interface KnowledgeCategory {
  id: string;
  parent_id?: string;
  name: string;
  slug: string;
  sort_order: number;
  children?: KnowledgeCategory[];
}

export interface KnowledgeTreeNode {
  id: string;
  type: 'category' | 'doc';
  title: string;
  slug?: string;
  path?: string;
  kind?: KnowledgeKind;
  children?: KnowledgeTreeNode[];
}

export interface KnowledgeVersion {
  version: number;
  title: string;
  summary: string;
  content?: string;
  editor: KnowledgePerson;
  created_at: string;
  is_current: boolean;
}

const publicOpts = { skipAuthRedirect: true, silent: true };

export function isKnowledgeLoginRequired(err: unknown): boolean {
  if (!axios.isAxiosError(err)) return false;
  const status = err.response?.status;
  const code = (err.response?.data as { code?: number } | undefined)?.code;
  return status === 401 || code === 33004;
}

export function getPublicPage(slug: string): Promise<KnowledgeDoc> {
  return request.get(`/knowledge/public/pages/${slug}`, publicOpts);
}

export function getPublicDoc(slug: string): Promise<KnowledgeDoc> {
  return request.get(`/knowledge/public/docs/${slug}`, publicOpts);
}

export function listPublicDocs(): Promise<PageResponse<KnowledgeDoc>> {
  return request.get('/knowledge/public/docs', publicOpts);
}

export function getPublicTree(): Promise<KnowledgeTreeNode[]> {
  return request.get('/knowledge/public/tree', publicOpts);
}

export function searchPublicDocs(q: string): Promise<PageResponse<KnowledgeDoc>> {
  return request.get('/knowledge/public/search', { ...publicOpts, params: { q } });
}

export function listKnowledgeDocs(params?: Record<string, unknown>): Promise<PageResponse<KnowledgeDoc>> {
  return request.get('/knowledge/docs', { params });
}

export function getKnowledgeDoc(id: string): Promise<KnowledgeDoc> {
  return request.get(`/knowledge/docs/${id}`);
}

export function createKnowledgeDoc(body: Record<string, unknown>): Promise<KnowledgeDoc> {
  return request.post('/knowledge/docs', body);
}

export function updateKnowledgeDoc(id: string, body: Record<string, unknown>): Promise<KnowledgeDoc> {
  return request.put(`/knowledge/docs/${id}`, body);
}

export function deleteKnowledgeDoc(id: string): Promise<void> {
  return request.delete(`/knowledge/docs/${id}`);
}

export function publishKnowledgeDoc(id: string): Promise<KnowledgeDoc> {
  return request.post(`/knowledge/docs/${id}/publish`);
}

export function getKnowledgeHistory(id: string): Promise<KnowledgeVersion[]> {
  return request.get(`/knowledge/docs/${id}/history`);
}

export function getKnowledgeVersion(id: string, version: number): Promise<KnowledgeVersion> {
  return request.get(`/knowledge/docs/${id}/versions/${version}`);
}

export function rollbackKnowledgeDoc(id: string, version: number): Promise<KnowledgeDoc> {
  return request.post(`/knowledge/docs/${id}/rollback`, { version });
}

export function searchKnowledge(q: string): Promise<PageResponse<KnowledgeDoc>> {
  return request.get('/knowledge/search', { params: { q } });
}

export function getKnowledgeTree(): Promise<KnowledgeTreeNode[]> {
  return request.get('/knowledge/tree');
}

export function listKnowledgeCategories(): Promise<KnowledgeCategory[]> {
  return request.get('/knowledge/categories');
}

export function createKnowledgeCategory(body: Record<string, unknown>): Promise<KnowledgeCategory> {
  return request.post('/knowledge/categories', body);
}

export function attachKnowledgeFile(id: string, fileId: string): Promise<KnowledgeDoc> {
  return request.post(`/knowledge/docs/${id}/attachments`, { file_id: fileId });
}
