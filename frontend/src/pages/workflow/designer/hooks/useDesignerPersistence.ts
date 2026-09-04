import { message } from 'antd';
import {
  createFlowDefinition,
  getFlowDefinitionDetail,
  getFlowVersion,
  listFlowVersions,
  publishFlowDefinition,
  saveFlowDraft,
} from '@/api/workflow';
import type { CreateDefinitionPayload, FlowGraphData } from '@/types/workflow';
import { fromBackendGraph, toPublishGraph } from '../utils/graphMapper';
import { validateDraftGraph, validatePublishGraph } from '../utils/graphValidate';

export async function loadDefinitionGraph(id: string): Promise<{
  name: string;
  key: string;
  status: number;
  graph: FlowGraphData;
}> {
  const detail = await getFlowDefinitionDetail(id);
  if (detail.draft_graph && detail.draft_graph.nodes) {
    return {
      name: detail.name,
      key: detail.key,
      status: detail.status,
      graph: fromBackendGraph(detail.draft_graph),
    };
  }
  const versions = await listFlowVersions(id);
  const current = versions.find((item) => item.status === 1) ?? versions[0];
  if (!current) {
    return { name: detail.name, key: detail.key, status: detail.status, graph: { nodes: [], edges: [] } };
  }
  const version = current.bpmn_data?.nodes
    ? current
    : await getFlowVersion(id, current.id);
  return {
    name: detail.name,
    key: detail.key,
    status: detail.status,
    graph: fromBackendGraph(version.bpmn_data ?? { nodes: [], edges: [] }),
  };
}

export async function persistDraft(
  definitionId: string | null,
  graph: FlowGraphData,
  createPayload?: CreateDefinitionPayload,
): Promise<string> {
  const issues = validateDraftGraph(graph);
  if (issues.length > 0) {
    throw new Error(issues[0].message);
  }
  let id = definitionId;
  if (!id) {
    if (!createPayload) {
      throw new Error('NEED_CREATE');
    }
    const created = await createFlowDefinition(createPayload);
    id = created.id;
  }
  await saveFlowDraft(id, graph);
  message.success('草稿已保存');
  return id;
}

export async function persistPublish(definitionId: string, graph: FlowGraphData): Promise<void> {
  const issues = validatePublishGraph(graph);
  if (issues.length > 0) {
    throw new Error(issues.map((item) => item.message).join('；'));
  }
  await publishFlowDefinition(definitionId, toPublishGraph(graph));
  message.success('流程已发布');
}
