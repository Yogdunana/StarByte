import request from './request';
import type {
  Task,
  TaskComment,
  CreateTaskParams,
  UpdateTaskParams,
  ListTaskParams,
  PageResponse,
} from '@/types/api';

// 获取任务列表
export function getTaskList(params: ListTaskParams): Promise<PageResponse<Task>> {
  return request.get('/tasks', { params });
}

// 获取我的任务
export function getMyTasks(params: ListTaskParams): Promise<PageResponse<Task>> {
  return request.get('/tasks/my', { params });
}

// 获取我创建的任务
export function getCreatedTasks(params: ListTaskParams): Promise<PageResponse<Task>> {
  return request.get('/tasks/created', { params });
}

// 获取任务详情
export function getTaskDetail(id: string): Promise<Task> {
  return request.get(`/tasks/${id}`);
}

// 创建任务
export function createTask(data: CreateTaskParams): Promise<Task> {
  return request.post('/tasks', data);
}

// 更新任务
export function updateTask(id: string, data: UpdateTaskParams): Promise<Task> {
  return request.put(`/tasks/${id}`, data);
}

// 删除任务
export function deleteTask(id: string): Promise<void> {
  return request.delete(`/tasks/${id}`);
}

// 更新任务状态
export function updateTaskStatus(id: string, status: number): Promise<Task> {
  return request.patch(`/tasks/${id}/status`, { status });
}

// 分配任务
export function assignTask(id: string, assigneeId: string): Promise<Task> {
  return request.post(`/tasks/${id}/assign`, { assignee_id: assigneeId });
}

// 获取任务评论
export function getTaskComments(taskId: string): Promise<TaskComment[]> {
  return request.get(`/tasks/${taskId}/comments`);
}

// 添加任务评论
export function addTaskComment(taskId: string, content: string): Promise<TaskComment> {
  return request.post(`/tasks/${taskId}/comments`, { content });
}
