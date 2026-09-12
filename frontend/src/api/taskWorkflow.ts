import request from './request';
import type { TaskLog, TaskPerson } from '@/types/api';
export interface TaskWorkflow {
  assignment_mode: string; revision: number; task_id: string; title: string; instance_id: string; stage: string; submission: string;
  creator: TaskPerson; assignee?: TaskPerson; reviewer: TaskPerson; acceptor: TaskPerson;
  can_start: boolean; can_pause: boolean; can_resume: boolean; can_submit: boolean; can_approve: boolean; can_return: boolean; can_reject: boolean; can_claim: boolean; updated_at: string; history: TaskLog[];
}
export function getTaskWorkflow(id: string): Promise<TaskWorkflow> { return request.get(`/tasks/${id}/workflow`); }
export function actTaskWorkflow(id: string, action: 'start' | 'pause' | 'resume' | 'submit' | 'approve' | 'return' | 'reject' | 'claim', comment: string, revision: number): Promise<TaskWorkflow> { return request.post(`/tasks/${id}/workflow/actions`, { action, comment, revision }); }
