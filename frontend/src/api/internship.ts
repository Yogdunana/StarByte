import request from './request';
import type {
  InternshipRecord,
  InternshipStats,
  CreateInternshipParams,
  UpdateInternshipParams,
  ListInternshipParams,
  PageResponse,
} from '@/types/api';

// 提交实习记录
export function createInternship(data: CreateInternshipParams): Promise<InternshipRecord> {
  return request.post('/internships', data);
}

// 获取实习记录列表
export function getInternshipList(
  params: ListInternshipParams,
): Promise<PageResponse<InternshipRecord>> {
  return request.get('/internships', { params });
}

// 获取我的实习记录
export function getMyInternships(params: ListInternshipParams): Promise<PageResponse<InternshipRecord>> {
  return request.get('/internships/my', { params });
}

// 获取实习记录详情
export function getInternshipDetail(id: string): Promise<InternshipRecord> {
  return request.get(`/internships/${id}`);
}

// 更新实习记录
export function updateInternship(id: string, data: Partial<CreateInternshipParams>): Promise<InternshipRecord> {
  return request.put(`/internships/${id}`, data);
}

// 删除实习记录
export function deleteInternship(id: string): Promise<void> {
  return request.delete(`/internships/${id}`);
}

// 审核实习记录
export function reviewInternship(id: string, data: UpdateInternshipParams): Promise<InternshipRecord> {
  return request.post(`/internships/${id}/review`, data);
}

// 获取实习统计排名
export function getInternshipStats(params: {
  page?: number;
  page_size?: number;
  department_id?: string;
  start_date?: string;
  end_date?: string;
}): Promise<PageResponse<InternshipStats>> {
  return request.get('/internships/stats', { params });
}

// 获取个人实习统计
export function getMyInternshipStats(): Promise<InternshipStats> {
  return request.get('/internships/stats/my');
}
