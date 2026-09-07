import request from './request';
import type { PageResponse } from '@/types/api';

export interface ContractTemplate {
  id: string;
  name: string;
  code: string;
  content: string;
}

export interface ContractItem {
  id: string;
  title: string;
  status: number;
  contract_type: number;
  party_name: string;
  amount?: number;
  template_id?: string;
  template_name?: string;
  file_id?: string;
  file_name?: string;
  owner: { id: string; name: string };
  start_at?: string;
  expired_at?: string;
  created_at: string;
}

export interface CreateContract {
  title: string;
  contract_type: number;
  party_name: string;
  amount?: number;
  template_id?: string;
  file_id?: string;
  start_at?: string;
  expired_at?: string;
  status?: number;
}

export function getContracts(params: Record<string, unknown>): Promise<PageResponse<ContractItem>> {
  return request.get('/contracts', { params });
}

export function getContract(id: string): Promise<ContractItem> {
  return request.get(`/contracts/${id}`);
}

export function createContract(data: CreateContract): Promise<ContractItem> {
  return request.post('/contracts', data);
}

export function updateContract(id: string, data: Partial<CreateContract>): Promise<ContractItem> {
  return request.put(`/contracts/${id}`, data);
}

export function deleteContract(id: string): Promise<void> {
  return request.delete(`/contracts/${id}`);
}

export function getContractTemplates(): Promise<ContractTemplate[]> {
  return request.get('/contracts/templates');
}

export function getExpiringContracts(days = 30): Promise<ContractItem[]> {
  return request.get('/contracts/expiring', { params: { days } });
}
