import request from './request';

export interface Schedule {
  id: string;
  user_id: string;
  user_name: string;
  department_id: string;
  dept_name: string;
  duty_date: string;
  time_slot: string;
  location: string;
  remark: string;
  status: number;
  status_text: string;
  created_at: string;
  updated_at: string;
}

export interface SwapRequest {
  id: string;
  requester_id: string;
  requester_name: string;
  target_user_id: string;
  target_user_name: string;
  requester_schedule_id: string;
  target_schedule_id: string;
  reason: string;
  status: number;
  status_text: string;
  approver_id: string;
  approver_name: string;
  approved_at: string;
  approved_remark: string;
  created_at: string;
  updated_at: string;
}

export interface DutyStats {
  total_duties: number;
  completed_duties: number;
  absent_duties: number;
  pending_duties: number;
  completion_rate: number;
  user_stats: Array<{
    user_id: string;
    user_name: string;
    dept_name: string;
    total_count: number;
    completed: number;
    absent: number;
    rate: number;
  }>;
}

export interface ListScheduleParams {
  page?: number;
  page_size?: number;
  department_id?: string;
  user_id?: string;
  start_date?: string;
  end_date?: string;
  view?: 'day' | 'week' | 'month';
}

export function getScheduleList(params: ListScheduleParams) {
  return request.get('/duty/schedule', { params });
}

export function createSchedule(data: {
  user_id: string;
  department_id?: string;
  duty_date: string;
  time_slot: string;
  location?: string;
  remark?: string;
}) {
  return request.post('/duty/schedule', data);
}

export function batchCreateSchedules(data: {
  user_ids: string[];
  department_id?: string;
  start_date: string;
  end_date: string;
  time_slot: string;
  rotation?: boolean;
  location?: string;
}) {
  return request.post('/duty/schedule/batch', data);
}

export function updateSchedule(id: string, data: Partial<{
  user_id: string;
  duty_date: string;
  time_slot: string;
  location: string;
  remark: string;
  status: number;
}>) {
  return request.put(`/duty/schedule/${id}`, data);
}

export function createSwapRequest(data: {
  target_user_id: string;
  requester_schedule_id: string;
  target_schedule_id?: string;
  reason: string;
}) {
  return request.post('/duty/swap', data);
}

export function actionSwap(id: string, data: { status: number; remark?: string }) {
  return request.put(`/duty/swap/${id}`, data);
}

export function getSwapList(params: {
  page?: number;
  page_size?: number;
  requester_id?: string;
  status?: number;
}) {
  return request.get('/duty/swap', { params });
}

export function getDutyStats(params: {
  department_id?: string;
  user_id?: string;
  start_date?: string;
  end_date?: string;
}) {
  return request.get('/duty/stats', { params });
}
