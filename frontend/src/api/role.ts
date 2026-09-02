import request from './request';
import type {
  Role,
  Permission,
  CreateRoleParams,
  UpdateRoleParams,
  ListRoleParams,
  PageResponse,
  Option,
} from '@/types/api';

// 获取角色列表
export function getRoleList(params: ListRoleParams): Promise<PageResponse<Role>> {
  return request.get('/roles', { params });
}

// 获取角色详情
export function getRoleDetail(id: string): Promise<Role> {
  return request.get(`/roles/${id}`);
}

// 创建角色
export function createRole(data: CreateRoleParams): Promise<Role> {
  return request.post('/roles', data);
}

// 更新角色
export function updateRole(id: string, data: UpdateRoleParams): Promise<Role> {
  return request.put(`/roles/${id}`, data);
}

// 删除角色
export function deleteRole(id: string): Promise<void> {
  return request.delete(`/roles/${id}`);
}

// 分配权限
export function assignRolePermissions(roleId: string, permissionIds: string[]): Promise<void> {
  return request.post(`/roles/${roleId}/permissions`, { permission_ids: permissionIds });
}

// 获取角色权限列表
export function getRolePermissions(roleId: string): Promise<Permission[]> {
  return request.get(`/roles/${roleId}/permissions`);
}

// 获取角色选项（下拉选择用）
export function getRoleOptions(): Promise<Option[]> {
  return request.get('/roles/options');
}

// 获取全部权限树
export function getPermissionTree(): Promise<Permission[]> {
  return request.get('/permissions/tree');
}
