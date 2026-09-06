import request from './request';
import type {
  Role,
  Permission,
  CreateRoleParams,
  UpdateRoleParams,
  ListRoleParams,
  PageResponse,
} from '@/types/api';

// 获取角色列表
export function getRoleList(params: ListRoleParams): Promise<PageResponse<Role>> {
  return request.get('/system/roles', { params });
}

// 获取角色详情
export function getRoleDetail(id: string): Promise<Role> {
  return request.get(`/system/roles/${id}`);
}

// 创建角色
export function createRole(data: CreateRoleParams): Promise<Role> {
  return request.post('/system/roles', data);
}

// 更新角色
export function updateRole(id: string, data: UpdateRoleParams): Promise<Role> {
  return request.put(`/system/roles/${id}`, data);
}

// 删除角色
export function deleteRole(id: string): Promise<void> {
  return request.delete(`/system/roles/${id}`);
}

// 分配权限
export function assignRolePermissions(roleId: string, permissionIds: string[]): Promise<void> {
  return request.put(`/system/roles/${roleId}/permissions`, { permission_ids: permissionIds });
}

// 获取全部权限树
export function getPermissionTree(): Promise<Permission[]> {
  return request.get('/system/permissions');
}
