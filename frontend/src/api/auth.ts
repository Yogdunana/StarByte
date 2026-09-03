import request from './request';
import type { LoginRequest, LoginResponse, RefreshResponse, RegisterRequest, UserInfo } from '@/types/api';

// 登录
export function login(params: LoginRequest): Promise<LoginResponse> {
  return request.post('/auth/login', params);
}

// 注册
export function register(params: RegisterRequest): Promise<{ id: string; username: string }> {
  return request.post('/auth/register', params);
}

// 刷新 Token
export function refreshToken(refreshToken: string): Promise<RefreshResponse> {
  return request.post('/auth/refresh', { refresh_token: refreshToken });
}

// 登出
export function logout(): Promise<void> {
  return request.post('/auth/logout');
}

// 获取当前用户信息
export function getCurrentUser(): Promise<UserInfo> {
  return request.get('/auth/me');
}

// 修改密码
export function changePassword(params: { old_password: string; new_password: string }): Promise<void> {
  return request.put('/auth/password', params);
}
