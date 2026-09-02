import request from './request';
import type {
  Meeting,
  MeetingAgenda,
  MeetingVote,
  VoteRecord,
  CreateMeetingParams,
  UpdateMeetingParams,
  CreateVoteParams,
  ListMeetingParams,
  PageResponse,
} from '@/types/api';

// ============================================================
// 会议
// ============================================================

// 获取会议列表
export function getMeetingList(params: ListMeetingParams): Promise<PageResponse<Meeting>> {
  return request.get('/meetings', { params });
}

// 获取会议详情
export function getMeetingDetail(id: string): Promise<Meeting> {
  return request.get(`/meetings/${id}`);
}

// 创建会议
export function createMeeting(data: CreateMeetingParams): Promise<Meeting> {
  return request.post('/meetings', data);
}

// 更新会议
export function updateMeeting(id: string, data: UpdateMeetingParams): Promise<Meeting> {
  return request.put(`/meetings/${id}`, data);
}

// 删除会议
export function deleteMeeting(id: string): Promise<void> {
  return request.delete(`/meetings/${id}`);
}

// 获取会议议程
export function getMeetingAgendas(meetingId: string): Promise<MeetingAgenda[]> {
  return request.get(`/meetings/${meetingId}/agendas`);
}

// ============================================================
// 投票
// ============================================================

// 获取会议投票列表
export function getMeetingVotes(meetingId: string): Promise<MeetingVote[]> {
  return request.get(`/meetings/${meetingId}/votes`);
}

// 获取投票详情
export function getVoteDetail(voteId: string): Promise<MeetingVote> {
  return request.get(`/meetings/votes/${voteId}`);
}

// 创建投票
export function createVote(data: CreateVoteParams): Promise<MeetingVote> {
  return request.post('/meetings/votes', data);
}

// 提交投票
export function submitVote(voteId: string, optionIds: string[]): Promise<void> {
  return request.post(`/meetings/votes/${voteId}/cast`, { option_ids: optionIds });
}

// 获取投票记录
export function getVoteRecords(voteId: string): Promise<VoteRecord[]> {
  return request.get(`/meetings/votes/${voteId}/records`);
}
