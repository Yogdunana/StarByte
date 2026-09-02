import request from './request';
import type {
  FlowDefinition,
  FlowInstance,
  FlowTask,
  FlowTaskHistory,
  ListFlowDefinitionParams,
  ListFlowInstanceParams,
  ListFlowTaskParams,
  StartFlowParams,
  ApproveTaskParams,
  PageResponse,
} from '@/types/api';

// ============================================================
// 流程定义
// ============================================================

// 获取流程定义列表
export function getFlowDefinitionList(
  params: ListFlowDefinitionParams,
): Promise<PageResponse<FlowDefinition>> {
  return request.get('/workflow/definitions', { params });
}

// 获取流程定义详情
export function getFlowDefinitionDetail(id: string): Promise<FlowDefinition> {
  return request.get(`/workflow/definitions/${id}`);
}

// 发布流程
export function publishFlowDefinition(id: string): Promise<FlowDefinition> {
  return request.post(`/workflow/definitions/${id}/publish`);
}

// 停用流程
export function deactivateFlowDefinition(id: string): Promise<FlowDefinition> {
  return request.post(`/workflow/definitions/${id}/deactivate`);
}

// ============================================================
// 流程实例
// ============================================================

// 发起流程
export function startFlow(data: StartFlowParams): Promise<FlowInstance> {
  return request.post('/workflow/instances/start', data);
}

// 获取流程实例列表
export function getFlowInstanceList(
  params: ListFlowInstanceParams,
): Promise<PageResponse<FlowInstance>> {
  return request.get('/workflow/instances', { params });
}

// 获取我发起的流程
export function getMyFlowInstances(params: ListFlowInstanceParams): Promise<PageResponse<FlowInstance>> {
  return request.get('/workflow/instances/my', { params });
}

// 获取流程实例详情
export function getFlowInstanceDetail(id: string): Promise<FlowInstance> {
  return request.get(`/workflow/instances/${id}`);
}

// 获取流程实例历史
export function getFlowInstanceHistory(instanceId: string): Promise<FlowTaskHistory[]> {
  return request.get(`/workflow/instances/${instanceId}/history`);
}

// ============================================================
// 流程任务
// ============================================================

// 获取待办任务
export function getMyTodoTasks(params: ListFlowTaskParams): Promise<PageResponse<FlowTask>> {
  return request.get('/workflow/tasks/todo', { params });
}

// 获取已办任务
export function getMyDoneTasks(params: ListFlowTaskParams): Promise<PageResponse<FlowTask>> {
  return request.get('/workflow/tasks/done', { params });
}

// 获取任务详情
export function getFlowTaskDetail(id: string): Promise<FlowTask> {
  return request.get(`/workflow/tasks/${id}`);
}

// 审批任务
export function approveFlowTask(data: ApproveTaskParams): Promise<void> {
  return request.post('/workflow/tasks/approve', data);
}
