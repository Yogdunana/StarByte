import request from './request';
import type { PageResponse } from '@/types/api';

export interface FinanceCategory {
  id: string;
  name: string;
  code: string;
  direction: number;
  description?: string;
}

export interface FinanceRecord {
  id: string;
  category_id: string;
  category_name: string;
  department_id?: string;
  department_name?: string;
  amount: number;
  direction: number;
  occurred_at: string;
  title: string;
  remark?: string;
  created_at: string;
}

export interface FinanceSummary {
  income_total: number;
  expense_total: number;
  balance: number;
  income_count: number;
  expense_count: number;
  by_category: Array<{
    category_id: string;
    category_name: string;
    direction: number;
    total: number;
    count: number;
  }>;
}

export interface CreateFinanceRecord {
  category_id: string;
  amount: number;
  direction: number;
  occurred_at: string;
  title: string;
  remark?: string;
}

export function getFinanceRecords(params: Record<string, unknown>): Promise<PageResponse<FinanceRecord>> {
  return request.get('/finance/records', { params });
}

export function createFinanceRecord(data: CreateFinanceRecord): Promise<FinanceRecord> {
  return request.post('/finance/records', data);
}

export function updateFinanceRecord(id: string, data: Partial<CreateFinanceRecord>): Promise<FinanceRecord> {
  return request.put(`/finance/records/${id}`, data);
}

export function deleteFinanceRecord(id: string): Promise<void> {
  return request.delete(`/finance/records/${id}`);
}

export function getFinanceCategories(): Promise<FinanceCategory[]> {
  return request.get('/finance/categories');
}

export function getFinanceSummary(params?: Record<string, unknown>): Promise<FinanceSummary> {
  return request.get('/finance/summary', { params });
}

export function exportFinance(): Promise<void> {
  return request.get('/finance/export');
}
