import request from './request';
import type {
  Department,
  CreateDepartmentParams,
  UpdateDepartmentParams,
} from '@/types/api';

export function getDepartmentTree(): Promise<Department[]> {
  return request.get('/system/departments');
}

export function getDepartmentDetail(id: string): Promise<Department> {
  return request.get(`/system/departments/${id}`);
}

export function createDepartment(data: CreateDepartmentParams): Promise<Department> {
  return request.post('/system/departments', data);
}

export function updateDepartment(id: string, data: UpdateDepartmentParams): Promise<Department> {
  return request.put(`/system/departments/${id}`, data);
}

export function deleteDepartment(id: string): Promise<void> {
  return request.delete(`/system/departments/${id}`);
}
