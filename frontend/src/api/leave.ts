import request from './request';
import type { PageResponse } from '@/types/api';

export type LeaveStatus = 'pending' | 'approved' | 'rejected';

export interface LeaveType {
  id: string;
  name: string;
  code: string;
  deductible: boolean;
  default_days: number;
  description: string;
}

export interface LeavePerson {
  id: string;
  name: string;
}

export interface LeaveApplication {
  id: string;
  applicant: LeavePerson;
  leave_type: LeaveType;
  start_time: string;
  end_time: string;
  duration_days: number;
  reason: string;
  status: LeaveStatus;
  approver?: LeavePerson;
  approve_remark?: string;
  approved_at?: string;
  created_at: string;
  updated_at: string;
}

export interface LeaveBalance {
  id: string;
  user_id: string;
  year: number;
  total_days: number;
  used_days: number;
  remaining_days: number;
  leave_type: LeaveType;
}

export interface SubmitLeaveParams {
  leave_type_id: string;
  start_time: string;
  end_time: string;
  reason: string;
}

export interface ListLeaveParams {
  page?: number;
  page_size?: number;
  status?: LeaveStatus | '';
  user_id?: string;
  year?: number;
}

export interface LeaveStats {
  total: number;
  by_status: Record<string, number>;
  by_type: { leave_type_id: string; code: string; name: string; count: number; days: number }[];
}

export function getLeaveTypes(): Promise<LeaveType[]> {
  return request.get('/leave/types');
}

export function getMyLeaveList(params?: ListLeaveParams): Promise<PageResponse<LeaveApplication>> {
  return request.get('/leave/my', { params });
}

export function getLeaveList(params?: ListLeaveParams): Promise<PageResponse<LeaveApplication>> {
  return request.get('/leave', { params });
}

export function getLeaveDetail(id: string): Promise<LeaveApplication> {
  return request.get(`/leave/${id}`);
}

export function submitLeave(data: SubmitLeaveParams): Promise<LeaveApplication> {
  return request.post('/leave', data);
}

export function approveLeave(id: string, remark?: string): Promise<void> {
  return request.put(`/leave/${id}/approve`, { remark });
}

export function rejectLeave(id: string, remark?: string): Promise<void> {
  return request.put(`/leave/${id}/reject`, { remark });
}

export function getLeaveBalances(params?: { user_id?: string; year?: number }): Promise<LeaveBalance[]> {
  return request.get('/leave/balance', { params });
}

export function getLeaveStats(): Promise<LeaveStats> {
  return request.get('/leave/stats');
}
