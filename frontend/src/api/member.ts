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

export interface AdmissionSignature {
 id: string; stage: string; revision: number; signer_id: string; signer_name?: string;
 signer_role: string; decision: string; comment: string; delegated: boolean;
 delegation_reason: string; created_at: string;
}
export interface AdmissionObjection {
 id: string; reason: string; status: string; center_comment: string; final_comment: string; created_at: string;
}
export interface AdmissionState {
 objections: AdmissionObjection[]; allowed_objection_actions: string[];
 application_id: string; stage: string; revision: number; historical_review_required: boolean;
 signatures: AdmissionSignature[]; allowed_roles: string[]; interview_completed: boolean;
}
export interface SignAdmissionParams {
 stage: string; revision: number; role: string; decision: 'approve' | 'reject' | 'supplement';
 comment: string; delegation_reason?: string; required_fields?: string[];
}
export function getAdmission(id: string): Promise<AdmissionState> { return request.get(`/member/applications/${id}/admission`); }
export function signAdmission(id: string, data: SignAdmissionParams): Promise<AdmissionState> { return request.post(`/member/applications/${id}/admission/sign`, data); }

export function handleAdmissionObjection(id: string, action: string, comment: string): Promise<void> { return request.post(`/member/applications/${id}/admission/objection`, { action, comment }); }
