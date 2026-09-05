import request from './request';
import type {
  Task,
  TaskComment,
  TaskLog,
  TaskAttachment,
  TaskStats,
  CreateTaskParams,
  UpdateTaskParams,
  ListTaskParams,
  PageResponse,
} from '@/types/api';

export function getTaskList(params: ListTaskParams): Promise<PageResponse<Task>> {
  return request.get('/tasks', { params });
}

export function getTaskDetail(id: string): Promise<Task> {
  return request.get(`/tasks/${id}`);
}

export function createTask(data: CreateTaskParams): Promise<Task> {
  return request.post('/tasks', data);
}

export function updateTask(id: string, data: UpdateTaskParams): Promise<Task> {
  return request.put(`/tasks/${id}`, data);
}

export function deleteTask(id: string): Promise<void> {
  return request.delete(`/tasks/${id}`);
}

export function updateTaskStatus(id: string, status: number, comment?: string): Promise<Task> {
  return request.post(`/tasks/${id}/status`, { status, comment });
}

export function assignTask(id: string, assigneeId: string): Promise<Task> {
  return request.post(`/tasks/${id}/assign`, { assignee_id: assigneeId });
}

export function transferTask(id: string, newAssigneeId: string, reason: string): Promise<Task> {
  return request.post(`/tasks/${id}/transfer`, { new_assignee_id: newAssigneeId, reason });
}

export function urgeTask(id: string, message?: string): Promise<void> {
  return request.post(`/tasks/${id}/urge`, { message });
}

export function getTaskLogs(id: string): Promise<TaskLog[]> {
  return request.get(`/tasks/${id}/logs`);
}

export function getTaskComments(taskId: string): Promise<TaskComment[]> {
  return request.get(`/tasks/${taskId}/comments`);
}

export function addTaskComment(taskId: string, content: string, mentions?: string[]): Promise<TaskComment> {
  return request.post(`/tasks/${taskId}/comments`, { content, mentions });
}

export function updateTaskComment(taskId: string, cid: string, content: string): Promise<TaskComment> {
  return request.put(`/tasks/${taskId}/comments/${cid}`, { content });
}

export function deleteTaskComment(taskId: string, cid: string): Promise<void> {
  return request.delete(`/tasks/${taskId}/comments/${cid}`);
}

export function getTaskAttachments(taskId: string): Promise<TaskAttachment[]> {
  return request.get(`/tasks/${taskId}/attachments`);
}

export function uploadTaskAttachment(taskId: string, file: File): Promise<TaskAttachment> {
  const form = new FormData();
  form.append('file', file);
  return request.post(`/tasks/${taskId}/attachments`, form, {
    headers: { 'Content-Type': 'multipart/form-data' },
  });
}

export async function downloadTaskAttachment(taskId: string, aid: string, filename: string): Promise<void> {
  const res = await request.get(`/tasks/${taskId}/attachments/${aid}`, { responseType: 'blob' });
  const blob = (res as { data: Blob }).data;
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename || 'attachment';
  a.click();
  URL.revokeObjectURL(url);
}

export function deleteTaskAttachment(taskId: string, aid: string): Promise<void> {
  return request.delete(`/tasks/${taskId}/attachments/${aid}`);
}

export function getMyTodo(params: ListTaskParams): Promise<PageResponse<Task>> {
  return request.get('/tasks/my/todo', { params });
}

export function getMyDone(params: ListTaskParams): Promise<PageResponse<Task>> {
  return request.get('/tasks/my/done', { params });
}

export function getMyCreated(params: ListTaskParams): Promise<PageResponse<Task>> {
  return request.get('/tasks/my/created', { params });
}

export function getMyOverdue(params: ListTaskParams): Promise<PageResponse<Task>> {
  return request.get('/tasks/my/overdue', { params });
}

export function getTaskStats(params?: { department_id?: string; start_date?: string; end_date?: string }): Promise<TaskStats> {
  return request.get('/tasks/stats', { params });
}
