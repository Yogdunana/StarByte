import request from './request';
import type {
  Interview,
  InterviewSession,
  InterviewQRCode,
  EvaluationDimension,
  EvaluationSummary,
  InterviewStats,
  CreateSessionParams,
  CreateInterviewParams,
  ListSessionParams,
  ListInterviewParams,
  PageResponse,
  InterviewRecordStatus,
} from '@/types/api';

export function getSessionList(params: ListSessionParams): Promise<PageResponse<InterviewSession>> {
  return request.get('/interviews/sessions', { params });
}

export function getSessionDetail(id: string): Promise<InterviewSession> {
  return request.get(`/interviews/sessions/${id}`);
}

export function createSession(data: CreateSessionParams): Promise<InterviewSession> {
  return request.post('/interviews/sessions', data);
}

export function updateSession(id: string, data: Partial<CreateSessionParams>): Promise<InterviewSession> {
  return request.put(`/interviews/sessions/${id}`, data);
}

export function deleteSession(id: string): Promise<void> {
  return request.delete(`/interviews/sessions/${id}`);
}

export function startSession(id: string): Promise<InterviewSession> {
  return request.post(`/interviews/sessions/${id}/start`);
}

export function endSession(id: string): Promise<InterviewSession> {
  return request.post(`/interviews/sessions/${id}/end`);
}

export function getSessionQRCode(id: string): Promise<InterviewQRCode> {
  return request.get(`/interviews/sessions/${id}/qrcode`);
}

export function getInterviewList(params: ListInterviewParams): Promise<PageResponse<Interview>> {
  return request.get('/interviews', { params });
}

export function getInterviewDetail(id: string): Promise<Interview> {
  return request.get(`/interviews/${id}`);
}

export function createInterview(data: CreateInterviewParams): Promise<Interview> {
  return request.post('/interviews', data);
}

export function assignEvaluators(id: string, evaluator_ids: string[]): Promise<Interview> {
  return request.post(`/interviews/${id}/assign`, { evaluator_ids });
}

export function checkinInterview(id: string, token?: string): Promise<Interview> {
  return request.post(`/interviews/${id}/checkin`, { token });
}

export function startInterview(id: string): Promise<Interview> {
  return request.post(`/interviews/${id}/start`);
}

export function endInterview(id: string): Promise<Interview> {
  return request.post(`/interviews/${id}/end`);
}

export function submitInterviewResult(id: string, result: 1 | 2 | 3, comment: string): Promise<Interview> {
  return request.post(`/interviews/${id}/result`, { result, comment });
}

export function getMyInterviews(status?: InterviewRecordStatus): Promise<Interview[]> {
  return request.get('/interviews/my', { params: { status } });
}

export function getDimensions(): Promise<EvaluationDimension[]> {
  return request.get('/interviews/dimensions');
}

export function createDimension(data: {
  name: string;
  weight: number;
  max_score: number;
  sort_order?: number;
}): Promise<EvaluationDimension> {
  return request.post('/interviews/dimensions', data);
}

export function updateDimension(
  id: string,
  data: Partial<{ name: string; weight: number; max_score: number; sort_order: number }>,
): Promise<EvaluationDimension> {
  return request.put(`/interviews/dimensions/${id}`, data);
}

export function deleteDimension(id: string): Promise<void> {
  return request.delete(`/interviews/dimensions/${id}`);
}

export function getEvaluationSummary(id: string): Promise<EvaluationSummary> {
  return request.get(`/interviews/${id}/evaluations`);
}

export function submitEvaluations(
  id: string,
  evaluations: Array<{ dimension: string; score: number; comment: string }>,
): Promise<EvaluationSummary> {
  return request.post(`/interviews/${id}/evaluations`, { evaluations });
}

export function getInterviewStats(params?: {
  start_date?: string;
  end_date?: string;
  department_id?: string;
  round?: number;
}): Promise<InterviewStats> {
  return request.get('/interviews/stats', { params });
}
