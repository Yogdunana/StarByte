import request from './request';
import type {
  Meeting,
  MeetingAgenda,
  MeetingAttendee,
  MeetingVote,
  MeetingQRCode,
  VoteResult,
  VoteWeightConfig,
  CreateMeetingParams,
  UpdateMeetingParams,
  CreateVoteParams,
  ListMeetingParams,
  PageResponse,
} from '@/types/api';

export function getMeetingList(params: ListMeetingParams): Promise<PageResponse<Meeting>> {
  return request.get('/meetings', { params });
}

export function getMeetingDetail(id: string): Promise<Meeting> {
  return request.get(`/meetings/${id}`);
}

export function createMeeting(data: CreateMeetingParams): Promise<Meeting> {
  return request.post('/meetings', data);
}

export function updateMeeting(id: string, data: UpdateMeetingParams): Promise<Meeting> {
  return request.put(`/meetings/${id}`, data);
}

export function deleteMeeting(id: string): Promise<void> {
  return request.delete(`/meetings/${id}`);
}

export function startMeeting(id: string): Promise<Meeting> {
  return request.post(`/meetings/${id}/start`);
}

export function endMeeting(id: string): Promise<Meeting> {
  return request.post(`/meetings/${id}/end`);
}

export function cancelMeeting(id: string, reason?: string): Promise<Meeting> {
  return request.post(`/meetings/${id}/cancel`, { reason });
}

export function updateMinutes(id: string, minutes: string): Promise<Meeting> {
  return request.put(`/meetings/${id}/minutes`, { minutes });
}

export function getMeetingQRCode(id: string): Promise<MeetingQRCode> {
  return request.get(`/meetings/${id}/qrcode`);
}

export function getMeetingAgendas(meetingId: string): Promise<MeetingAgenda[]> {
  return request.get(`/meetings/${meetingId}/agendas`);
}

export function addAgenda(meetingId: string, data: {
  title: string; content?: string; duration?: number; presenter?: string; sort_order?: number;
}): Promise<MeetingAgenda> {
  return request.post(`/meetings/${meetingId}/agendas`, data);
}

export function updateAgenda(meetingId: string, agendaId: string, data: {
  title?: string; content?: string; duration?: number; presenter?: string;
}): Promise<MeetingAgenda> {
  return request.put(`/meetings/${meetingId}/agendas/${agendaId}`, data);
}

export function deleteAgenda(meetingId: string, agendaId: string): Promise<void> {
  return request.delete(`/meetings/${meetingId}/agendas/${agendaId}`);
}

export function sortAgendas(meetingId: string, agenda_ids: string[]): Promise<MeetingAgenda[]> {
  return request.put(`/meetings/${meetingId}/agendas/sort`, { agenda_ids });
}

export function getAttendees(meetingId: string): Promise<MeetingAttendee[]> {
  return request.get(`/meetings/${meetingId}/attendees`);
}

export function addAttendees(meetingId: string, user_ids: string[]): Promise<MeetingAttendee[]> {
  return request.post(`/meetings/${meetingId}/attendees`, { user_ids });
}

export function removeAttendee(meetingId: string, userId: string): Promise<void> {
  return request.delete(`/meetings/${meetingId}/attendees/${userId}`);
}

export function checkinMeeting(meetingId: string, token?: string): Promise<MeetingAttendee> {
  return request.post(`/meetings/${meetingId}/checkin`, { token });
}

export function getMeetingVotes(meetingId: string): Promise<MeetingVote[]> {
  return request.get(`/meetings/${meetingId}/votes`);
}

export function createVote(meetingId: string, data: CreateVoteParams): Promise<MeetingVote> {
  return request.post(`/meetings/${meetingId}/votes`, data);
}

export function getVoteDetail(voteId: string): Promise<MeetingVote> {
  return request.get(`/votes/${voteId}`);
}

export function castVote(voteId: string, option_key: string): Promise<void> {
  return request.post(`/votes/${voteId}/cast`, { option_key });
}

export function getVoteResult(voteId: string): Promise<VoteResult> {
  return request.get(`/votes/${voteId}/result`);
}

export function closeVote(voteId: string): Promise<MeetingVote> {
  return request.post(`/votes/${voteId}/close`);
}

export function getMyVote(voteId: string): Promise<{ option_key: string; weight: number; voted_at: string }> {
  return request.get(`/votes/${voteId}/my`);
}

export function getVoteWeightConfig(): Promise<VoteWeightConfig> {
  return request.get('/system/vote-weight-config');
}

export function updateVoteWeightConfig(data: VoteWeightConfig): Promise<VoteWeightConfig> {
  return request.put('/system/vote-weight-config', data);
}
