import request from './request';
import type { TaskLog, TaskPerson } from '@/types/api';

export interface TaskHandoverSign {
  id: string;
  requirement: string;
  signer: TaskPerson;
  signer_role: string;
  waived: boolean;
  decision: string;
  comment: string;
  created_at: string;
}

export interface TaskHandover {
  id: string;
  task_id: string;
  task_title: string;
  kind: string;
  status: string;
  revision: number;
  from: TaskPerson;
  to: TaskPerson;
  reason: string;
  source_department: string;
  target_department: string;
  requirements: string[];
  can_sign: string[];
  signatures: TaskHandoverSign[];
  created_at: string;
}

export interface TaskWorkflow {
  assignment_mode: string;
  revision: number;
  task_id: string;
  title: string;
  instance_id: string;
  stage: string;
  submission: string;
  creator: TaskPerson;
  assignee?: TaskPerson;
  reviewer: TaskPerson;
  acceptor: TaskPerson;
  can_start: boolean;
  can_pause: boolean;
  can_resume: boolean;
  can_submit: boolean;
  can_approve: boolean;
  can_return: boolean;
  can_reject: boolean;
  can_claim: boolean;
  can_delegate: boolean;
  handover?: TaskHandover;
  updated_at: string;
  history: TaskLog[];
}

export function getTaskWorkflow(id: string): Promise<TaskWorkflow> {
  return request.get(`/tasks/${id}/workflow`);
}

export function actTaskWorkflow(
  id: string,
  action: 'start' | 'pause' | 'resume' | 'submit' | 'approve' | 'return' | 'reject' | 'claim',
  comment: string,
  revision: number,
): Promise<TaskWorkflow> {
  return request.post(`/tasks/${id}/workflow/actions`, { action, comment, revision });
}

export function requestTaskHandover(id: string, targetId: string, reason: string, revision: number): Promise<TaskHandover> {
  return request.post(`/tasks/${id}/handover`, { target_id: targetId, reason, revision });
}

export function getTaskHandover(id: string): Promise<TaskHandover> {
  return request.get(`/tasks/${id}/handover`);
}

export function decideTaskHandover(id: string, requirement: string, decision: 'approve' | 'reject', comment: string, revision: number): Promise<TaskHandover> {
  return request.post(`/tasks/${id}/handover/decisions`, { requirement, decision, comment, revision });
}

export function getTaskTransfer(id: string): Promise<TaskHandover> {
  return request.get(`/tasks/transfers/${id}`);
}

export function decideTaskTransfer(id: string, requirement: string, decision: 'approve' | 'reject', comment: string, revision: number): Promise<TaskHandover> {
  return request.post(`/tasks/transfers/${id}/decisions`, { requirement, decision, comment, revision });
}
