import request from './request';
import type { PageResponse } from '@/types/api';

export interface WorkflowInstance {
  id: string; definition_id: string; definition_version_id: string; definition_name: string;
  initiator_id: string; initiator_name: string; business_type: string; business_key: string;
  status: number; current_node_ids: string[]; started_at: string; ended_at?: string;
  terminate_reason: string;
}
export interface WorkflowTask {
  id: string; instance_id: string; node_id: string; node_name: string; task_type: string;
  definition_name: string; assignee_id?: string; assignee_name: string; status: number;
  instance_status: number; business_type: string; business_key: string;
  action: string; comment: string; due_date?: string; created_at: string; completed_at?: string;
  form_data?: Record<string, unknown>;
}
export interface WorkflowHistory {
  id: string; task_id?: string; node_id: string; node_name: string; operator_name: string;
  action: string; comment: string; from_node_id: string; to_node_id: string; created_at: string;
}
export const listWorkflowTasks = (kind: 'todo' | 'done', page = 1): Promise<PageResponse<WorkflowTask>> => request.get(`/workflow/tasks/${kind}`, { params: { page, page_size: 12 } });
export const getWorkflowTask = (id: string): Promise<WorkflowTask> => request.get(`/workflow/tasks/${id}`);
export const listWorkflowInstances = (page = 1, status?: number): Promise<PageResponse<WorkflowInstance>> => request.get('/workflow/instances', { params: { page, page_size: 12, status } });
export const getWorkflowInstance = (id: string): Promise<WorkflowInstance> => request.get(`/workflow/instances/${id}`);
export const getWorkflowHistory = (id: string): Promise<WorkflowHistory[]> => request.get(`/workflow/instances/${id}/history`);
export const decideWorkflowTask = (id: string, action: 'approve' | 'reject', comment: string): Promise<void> => request.post(`/workflow/tasks/${id}/${action}`, { action, comment });
export const transferWorkflowTask = (id: string, target_user_id: string, comment: string): Promise<void> => request.post(`/workflow/tasks/${id}/transfer`, { target_user_id, comment });
export const rollbackWorkflowTask = (id: string, target_node_id: string, comment: string): Promise<void> => request.post(`/workflow/tasks/${id}/rollback`, { target_node_id, comment });
export const changeWorkflowInstance = (id: string, action: 'suspend' | 'resume' | 'terminate', reason: string): Promise<void> => request.post(`/workflow/instances/${id}/${action}`, { reason });
export interface WorkflowPerson { id: string; name: string; department_name: string }
export const getWorkflowCandidates = (id: string, keyword: string): Promise<WorkflowPerson[]> => request.get(`/workflow/tasks/${id}/assignees`, { params: { keyword } });
