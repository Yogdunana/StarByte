import request from './request';
import type { UserInfo, PageResponse, PageParams } from '@/types/api';

export interface ListUserParams extends PageParams {
  status?: number;
  department_id?: string;
}

export interface CreateUserParams {
  username: string;
  password: string;
  real_name?: string;
  email?: string;
  phone?: string;
  gender?: number;
  department_id?: string;
  position_id?: string;
  role_ids?: string[];
}

export interface UpdateUserParams {
  real_name?: string;
  email?: string;
  phone?: string;
  gender?: number;
  status?: number;
  department_id?: string;
  position_id?: string;
  role_ids?: string[];
}

export interface UserListItem {
  id: string;
  username: string;
  real_name: string;
  avatar_url: string;
  email: string;
  phone: string;
  gender: number;
  status: number;
  department_id: string;
  department_name: string;
  position_id: string;
  position_name: string;
  last_login_at: string;
  created_at: string;
}

// 获取用户列表
export function getUserList(params: ListUserParams): Promise<PageResponse<UserListItem>> {
  return request.get('/users', { params });
}

// 获取用户详情
export function getUserDetail(id: string): Promise<UserInfo> {
  return request.get(`/users/${id}`);
}

// 创建用户
export function createUser(params: CreateUserParams): Promise<UserInfo> {
  return request.post('/users', params);
}

// 更新用户
export function updateUser(id: string, params: UpdateUserParams): Promise<UserInfo> {
  return request.put(`/users/${id}`, params);
}

// 删除用户
export function deleteUser(id: string): Promise<void> {
  return request.delete(`/users/${id}`);
}
