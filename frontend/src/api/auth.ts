import request from './request';
import type {
  CASExchangeResponse,
  CASStatusResponse,
  LoginRequest,
  LoginResponse,
  RefreshResponse,
  RegisterRequest,
  UserInfo,
} from '@/types/api';

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
export function logout(refreshToken?: string): Promise<void> {
  const body = refreshToken ? { refresh_token: refreshToken } : {};
  return request.post('/auth/logout', body);
}

// 获取当前用户信息
export function getCurrentUser(): Promise<UserInfo> {
  return request.get('/auth/me');
}

// 修改密码
export function changePassword(params: { old_password: string; new_password: string }): Promise<void> {
  return request.put('/auth/password', params);
}

export function getCasStatus(): Promise<CASStatusResponse> {
  return request.get('/auth/cas/status');
}

export function getCasLoginURL(redirect?: string): string {
  const base = import.meta.env.VITE_API_BASE_URL || '/api/v1';
  const query = redirect ? `?redirect=${encodeURIComponent(redirect)}` : '';
  return `${base}/auth/cas/login${query}`;
}

export function exchangeCasCode(code: string): Promise<CASExchangeResponse> {
  return request.post('/auth/cas/exchange', { code });
}
