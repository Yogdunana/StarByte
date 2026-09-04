import type {
  DesignerNodeType,
  FlowGraphData,
  FlowNode,
  ParallelConfig,
  PublishGraphData,
} from '@/types/workflow';

const TO_BACKEND: Record<DesignerNodeType, string> = {
  start: 'start',
  end: 'end',
  approval: 'approval',
  condition: 'exclusive_gateway',
  parallel: 'parallel_gateway',
  merge: 'parallel_gateway',
  timer: 'timer',
  notify: 'notification_task',
};

const FROM_BACKEND: Record<string, DesignerNodeType> = {
  start: 'start',
  end: 'end',
  approval: 'approval',
  exclusive_gateway: 'condition',
  notification_task: 'notify',
  timer: 'timer',
  notify: 'notify',
  condition: 'condition',
  parallel: 'parallel',
  merge: 'merge',
};

function isDesignerType(value: string): value is DesignerNodeType {
  return value in TO_BACKEND;
}

function mapPublishNode(node: FlowNode): PublishGraphData['nodes'][number] {
  const config =
    node.type === 'merge'
      ? { ...(node.data.config as ParallelConfig), kind: 'join' as const }
      : node.data.config;
  return {
    id: node.id,
    type: TO_BACKEND[node.type],
    position: node.position,
    data: {
      ...node.data,
      label: node.data.name,
      config,
    },
  };
}

/** 发布前：短名 → 后端引擎类型 */
export function toPublishGraph(graph: FlowGraphData): PublishGraphData {
  return {
    nodes: graph.nodes.map(mapPublishNode),
    edges: graph.edges,
  };
}

function mapLoadedNode(node: FlowNode | PublishGraphData['nodes'][number]): FlowNode {
  let type: DesignerNodeType = isDesignerType(node.type) ? node.type : 'approval';
  if (node.type === 'exclusive_gateway') {
    type = 'condition';
  } else if (node.type === 'notification_task') {
    type = 'notify';
  } else if (node.type === 'parallel_gateway') {
    const kind = (node.data.config as ParallelConfig | undefined)?.kind;
    type = kind === 'join' ? 'merge' : 'parallel';
  } else if (FROM_BACKEND[node.type]) {
    type = FROM_BACKEND[node.type];
  }
  const name = node.data.name || node.data.label || NODE_FALLBACK[type];
  return {
    ...node,
    type,
    data: {
      ...node.data,
      name,
      label: name,
      config: node.data.config ?? {},
    },
  };
}

const NODE_FALLBACK: Record<DesignerNodeType, string> = {
  start: '开始',
  end: '结束',
  approval: '审批',
  condition: '条件',
  parallel: '并行',
  merge: '合并',
  timer: '定时器',
  notify: '通知',
};

/** 加载已发布版本：后端类型 → 短名 */
export function fromBackendGraph(graph: FlowGraphData | PublishGraphData): FlowGraphData {
  return {
    nodes: graph.nodes.map(mapLoadedNode),
    edges: graph.edges,
  };
}
