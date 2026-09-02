import request from './request';
import type {
  MemberApplication,
  MemberProfile,
  CreateMemberApplicationParams,
  ListMemberApplicationParams,
  ListMemberParams,
  PageResponse,
  OperationResult,
} from '@/types/api';

// ============================================================
// 入会申请
// ============================================================

// 提交入会申请
export function submitApplication(data: CreateMemberApplicationParams): Promise<MemberApplication> {
  return request.post('/members/applications', data);
}

// 获取申请列表
export function getApplicationList(
  params: ListMemberApplicationParams,
): Promise<PageResponse<MemberApplication>> {
  return request.get('/members/applications', { params });
}

// 获取申请详情
export function getApplicationDetail(id: string): Promise<MemberApplication> {
  return request.get(`/members/applications/${id}`);
}

// 审核申请（通过/拒绝）
export function reviewApplication(
  id: string,
  data: { action: 'approve' | 'reject'; comment?: string },
): Promise<OperationResult> {
  return request.post(`/members/applications/${id}/review`, data);
}

// 取消申请
export function cancelApplication(id: string): Promise<void> {
  return request.post(`/members/applications/${id}/cancel`);
}

// ============================================================
// 会员档案
// ============================================================

// 获取会员列表
export function getMemberList(params: ListMemberParams): Promise<PageResponse<MemberProfile>> {
  return request.get('/members', { params });
}

// 获取会员详情
export function getMemberDetail(id: string): Promise<MemberProfile> {
  return request.get(`/members/${id}`);
}

// 更新会员信息
export function updateMember(
  id: string,
  data: Partial<{
    member_type: number;
    department_id: string;
    position_id: string;
    status: number;
  }>,
): Promise<MemberProfile> {
  return request.put(`/members/${id}`, data);
}
