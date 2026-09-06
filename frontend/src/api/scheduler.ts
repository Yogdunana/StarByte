import request from './request';
import type {
  CreateSchedulerTaskParams,
  PageResponse,
  SchedulerHandlerInfo,
  SchedulerLogs,
  SchedulerTask,
  UpdateSchedulerTaskParams,
} from '@/types/api';

export function getSchedulerTasks(params: {
  page?: number;
  page_size?: number;
  keyword?: string;
  status?: number;
}): Promise<PageResponse<SchedulerTask>> {
  return request.get('/system/scheduler/tasks', { params });
}

export function getSchedulerHandlers(): Promise<SchedulerHandlerInfo[]> {
  return request.get('/system/scheduler/handlers');
}

export function getSchedulerTask(id: string): Promise<SchedulerTask> {
  return request.get(`/system/scheduler/tasks/${id}`);
}

export function createSchedulerTask(data: CreateSchedulerTaskParams): Promise<SchedulerTask> {
  return request.post('/system/scheduler/tasks', data);
}

export function updateSchedulerTask(id: string, data: UpdateSchedulerTaskParams): Promise<SchedulerTask> {
  return request.put(`/system/scheduler/tasks/${id}`, data);
}

export function deleteSchedulerTask(id: string): Promise<void> {
  return request.delete(`/system/scheduler/tasks/${id}`);
}

export function runSchedulerTask(id: string): Promise<void> {
  return request.post(`/system/scheduler/tasks/${id}/run`);
}

export function pauseSchedulerTask(id: string): Promise<SchedulerTask> {
  return request.post(`/system/scheduler/tasks/${id}/pause`);
}

export function resumeSchedulerTask(id: string): Promise<SchedulerTask> {
  return request.post(`/system/scheduler/tasks/${id}/resume`);
}

export function getSchedulerLogs(id: string, runId?: string): Promise<SchedulerLogs> {
  return request.get(`/system/scheduler/tasks/${id}/logs`, {
    params: runId ? { run_id: runId } : undefined,
  });
}
