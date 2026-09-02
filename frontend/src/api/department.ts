import request from './request';
import type {
  Department,
  CreateDepartmentParams,
  UpdateDepartmentParams,
  PageResponse,
  Option,
} from '@/types/api';

// 获取部门树
export function getDepartmentTree(): Promise<Department[]> {
  return request.get('/departments/tree');
}

// 获取部门列表（分页）
export function getDepartmentList(params: {
  page?: number;
  page_size?: number;
  parent_id?: string;
  keyword?: string;
}): Promise<PageResponse<Department>> {
  return request.get('/departments', { params });
}

// 获取部门详情
export function getDepartmentDetail(id: string): Promise<Department> {
  return request.get(`/departments/${id}`);
}

// 创建部门
export function createDepartment(data: CreateDepartmentParams): Promise<Department> {
  return request.post('/departments', data);
}

// 更新部门
export function updateDepartment(id: string, data: UpdateDepartmentParams): Promise<Department> {
  return request.put(`/departments/${id}`, data);
}

// 删除部门
export function deleteDepartment(id: string): Promise<void> {
  return request.delete(`/departments/${id}`);
}

// 获取部门选项（下拉选择用）
export function getDepartmentOptions(): Promise<Option[]> {
  return request.get('/departments/options');
}
