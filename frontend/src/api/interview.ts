import request from './request';
import type {
  Interview,
  InterviewEvaluation,
  CreateInterviewParams,
  UpdateInterviewParams,
  ListInterviewParams,
  PageResponse,
} from '@/types/api';

// 获取面试列表
export function getInterviewList(
  params: ListInterviewParams,
): Promise<PageResponse<Interview>> {
  return request.get('/interviews', { params });
}

// 获取面试详情
export function getInterviewDetail(id: string): Promise<Interview> {
  return request.get(`/interviews/${id}`);
}

// 创建面试
export function createInterview(data: CreateInterviewParams): Promise<Interview> {
  return request.post('/interviews', data);
}

// 更新面试
export function updateInterview(id: string, data: UpdateInterviewParams): Promise<Interview> {
  return request.put(`/interviews/${id}`, data);
}

// 删除面试
export function deleteInterview(id: string): Promise<void> {
  return request.delete(`/interviews/${id}`);
}

// 获取面试评分列表
export function getInterviewEvaluations(interviewId: string): Promise<InterviewEvaluation[]> {
  return request.get(`/interviews/${interviewId}/evaluations`);
}

// 提交面试评分
export function submitInterviewEvaluation(
  interviewId: string,
  data: { score: number; comment: string; recommendation: number },
): Promise<InterviewEvaluation> {
  return request.post(`/interviews/${interviewId}/evaluations`, data);
}
