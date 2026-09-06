import request from './request';
import type { AuthSessionList, UserAuthSessions } from '@/types/api';

export function getSessions(params?: {
  keyword?: string;
  user_id?: string;
}): Promise<AuthSessionList> {
  return request.get('/auth/sessions', { params });
}

export function getUserSessions(userId: string): Promise<UserAuthSessions> {
  return request.get(`/auth/sessions/${encodeURIComponent(userId)}`);
}

export function kickSession(tokenId: string): Promise<void> {
  return request.delete(`/auth/sessions/${encodeURIComponent(tokenId)}`);
}

export function kickUserSessions(userId: string): Promise<void> {
  return request.delete(`/auth/sessions/user/${encodeURIComponent(userId)}`);
}
