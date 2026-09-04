import request from './request';
import type {
  MemberApplication,
  MemberProfile,
  CreateMemberApplicationParams,
  ResubmitMemberApplicationParams,
  ListMemberApplicationParams,
  ListMemberParams,
  MemberApplicationHistory,
  MemberProfileHistory,
  MemberStatsResponse,
  MemberDepartmentOption,
  PageResponse,
} from '@/types/api';

export function submitApplication(data: CreateMemberApplicationParams): Promise<MemberApplication> {
  return request.post('/member/applications', data);
}

export function resubmitApplication(
  id: string,
  data: ResubmitMemberApplicationParams,
): Promise<MemberApplication> {
  return request.post(`/member/applications/${id}/resubmit`, data);
}

export function getMyApplications(): Promise<MemberApplication[]> {
  return request.get('/member/applications/my');
}

export function getApplicationList(
  params: ListMemberApplicationParams,
): Promise<PageResponse<MemberApplication>> {
  return request.get('/member/applications', { params });
}

export function getApplicationDetail(id: string): Promise<MemberApplication> {
  return request.get(`/member/applications/${id}`);
}

export function getApplicationHistory(id: string): Promise<MemberApplicationHistory[]> {
  return request.get(`/member/applications/${id}/history`);
}

export function approveApplication(id: string, comment: string): Promise<MemberApplication> {
  return request.post(`/member/applications/${id}/approve`, { comment });
}

export function rejectApplication(id: string, comment: string): Promise<MemberApplication> {
  return request.post(`/member/applications/${id}/reject`, { comment });
}

export function supplementApplication(
  id: string,
  comment: string,
  required_fields: string[],
): Promise<MemberApplication> {
  return request.post(`/member/applications/${id}/supplement`, { comment, required_fields });
}

export function getMemberDepartments(): Promise<MemberDepartmentOption[]> {
  return request.get('/member/departments');
}

export function getMemberList(params: ListMemberParams): Promise<PageResponse<MemberProfile>> {
  return request.get('/member/profiles', { params });
}

export function getMemberDetail(id: string): Promise<MemberProfile> {
  return request.get(`/member/profiles/${id}`);
}

export function updateMember(
  id: string,
  data: Partial<{
    real_name: string;
    gender: number;
    grade: string;
    major: string;
    skills: string[];
    projects: Array<{ name: string; role: string; period: string }>;
    bio: string;
    contact_phone: string;
    contact_email: string;
  }>,
): Promise<MemberProfile> {
  return request.put(`/member/profiles/${id}`, data);
}

export function updateMemberStatus(
  id: string,
  status: number,
  reason: string,
): Promise<MemberProfile> {
  return request.put(`/member/profiles/${id}/status`, { status, reason });
}

export function getMemberHistory(id: string): Promise<MemberProfileHistory[]> {
  return request.get(`/member/profiles/${id}/history`);
}

export function exportMemberProfiles(params: ListMemberParams): Promise<{ data: Blob }> {
  return request.get('/member/profiles/export', { params, responseType: 'blob' });
}

export function getApplicationStats(params: {
  start_date?: string;
  end_date?: string;
  group_by?: string;
}): Promise<MemberStatsResponse> {
  return request.get('/member/stats/applications', { params });
}

export function getMemberStats(params: { group_by?: string }): Promise<MemberStatsResponse> {
  return request.get('/member/stats/members', { params });
}
