import request from './request';
import type { PageResponse } from '@/types/api';
import type {
  CreateDefinitionPayload,
  FlowDefinitionDTO,
  FlowGraphData,
  FlowVersionDTO,
  PublishGraphData,
} from '@/types/workflow';

export function getFlowDefinitionList(params: {
  page?: number;
  page_size?: number;
  keyword?: string;
  category?: string;
  status?: number;
}): Promise<PageResponse<FlowDefinitionDTO>> {
  return request.get('/workflow/definitions', { params });
}

export function getFlowDefinitionDetail(id: string): Promise<FlowDefinitionDTO> {
  return request.get(`/workflow/definitions/${id}`);
}

export function createFlowDefinition(data: CreateDefinitionPayload): Promise<FlowDefinitionDTO> {
  return request.post('/workflow/definitions', data);
}

export function saveFlowDraft(id: string, graphData: FlowGraphData): Promise<FlowDefinitionDTO> {
  return request.put(`/workflow/definitions/${id}/draft`, { graph_data: graphData });
}

export function publishFlowDefinition(
  id: string,
  graphData: PublishGraphData,
): Promise<FlowVersionDTO> {
  return request.post(`/workflow/definitions/${id}/publish`, { graph_data: graphData });
}

export function listFlowVersions(id: string): Promise<FlowVersionDTO[]> {
  return request.get(`/workflow/definitions/${id}/versions`);
}

export function getFlowVersion(id: string, versionId: string): Promise<FlowVersionDTO> {
  return request.get(`/workflow/definitions/${id}/versions/${versionId}`);
}
