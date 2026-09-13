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

export type PermissionInput = {
  name: string;
  code?: string;
  type?: Permission['type'];
  parent_id?: string;
  description?: string;
  path?: string;
  icon?: string;
  resource?: string;
  action?: string;
  api_method?: string;
  api_path?: string;
  status?: number;
  sort_order?: number;
};
export const createPermission = (data: PermissionInput): Promise<Permission> =>
  request.post('/system/permissions', data);
export const updatePermission = (id: string, data: PermissionInput): Promise<Permission> =>
  request.put(`/system/permissions/${id}`, data);
export const deletePermission = (id: string): Promise<void> =>
  request.delete(`/system/permissions/${id}`);

export interface RoleMember {
  department_ids?: string[];
  id: string;
  username: string;
  real_name: string;
  status: number;
}
export const getRoleMembers = (id: string, page: number): Promise<PageResponse<RoleMember>> =>
  request.get(`/system/roles/${id}/users`, { params: { page, page_size: 20 } });
export const addRoleMember = (
  id: string,
  userId: string,
  departmentIds: string[] = [],
): Promise<void> =>
  request.put(`/system/roles/${id}/users/${userId}`, { department_ids: departmentIds });
export const removeRoleMember = (id: string, userId: string): Promise<void> =>
  request.delete(`/system/roles/${id}/users/${userId}`);
