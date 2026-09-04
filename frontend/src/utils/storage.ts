import type { UserInfo } from '@/types/api';

const TOKEN_KEY = 'starbyte_access_token';
const REFRESH_TOKEN_KEY = 'starbyte_refresh_token';
const USER_INFO_KEY = 'starbyte_user_info';

// Token 相关
export function getToken(): string {
  return localStorage.getItem(TOKEN_KEY) || '';
}

export function setToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token);
}

export function getRefreshToken(): string {
  return localStorage.getItem(REFRESH_TOKEN_KEY) || '';
}

export function setRefreshToken(token: string): void {
  localStorage.setItem(REFRESH_TOKEN_KEY, token);
}

export function removeToken(): void {
  localStorage.removeItem(TOKEN_KEY);
  localStorage.removeItem(REFRESH_TOKEN_KEY);
  localStorage.removeItem(USER_INFO_KEY);
}

// 用户信息缓存
export function getUserInfo<T = UserInfo>(): T | null {
  const data = localStorage.getItem(USER_INFO_KEY);
  return data ? (JSON.parse(data) as T) : null;
}

export function setUserInfo(user: UserInfo): void {
  localStorage.setItem(USER_INFO_KEY, JSON.stringify(user));
}
