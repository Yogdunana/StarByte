import type { Edge, Node } from 'reactflow';
import type { DesignerNodeData, DesignerNodeType, FlowGraphData } from '@/types/workflow';
import { NODE_META } from '../constants';

export type DesignerRFNode = Node<DesignerNodeData, DesignerNodeType>;
export type DesignerRFEdge = Edge<{ condition?: string }>;

export function defaultNodeData(type: DesignerNodeType): DesignerNodeData {
  const name = NODE_META[type].label;
  if (type === 'approval') {
    return {
      name,
      label: name,
      config: { assigneeStrategy: 'static', assignees: [], approvalType: 'single' },
    };
  }
  if (type === 'condition') {
    return {
      name,
      label: name,
      config: {
        branches: [
          { id: 'branch_1', label: '条件1', expression: '', is_default: false },
          { id: 'branch_2', label: '默认', expression: '', is_default: true },
        ],
      },
    };
  }
  if (type === 'parallel') {
    return { name, label: name, config: { branchCount: 2, branchLabels: ['分支1', '分支2'] } };
  }
  if (type === 'merge') {
    return { name, label: name, config: { kind: 'join', branchCount: 2, branchLabels: [] } };
  }
  if (type === 'timer') {
    return { name, label: name, config: { duration: 1, unit: 'hours' } };
  }
  if (type === 'notify') {
    return { name, label: name, config: { notificationType: 'default', channels: ['in_app'] } };
  }
  return { name, label: name, config: {} };
}

export function toFlowGraph(nodes: DesignerRFNode[], edges: DesignerRFEdge[]): FlowGraphData {
  return {
    nodes: nodes.map((node) => ({
      id: node.id,
      type: (node.type ?? 'approval') as DesignerNodeType,
      position: { x: node.position.x, y: node.position.y },
      data: node.data,
    })),
    edges: edges.map((edge) => ({
      id: edge.id,
      source: edge.source,
      target: edge.target,
      sourceHandle: edge.sourceHandle ?? undefined,
      label: typeof edge.label === 'string' ? edge.label : undefined,
      data: edge.data,
    })),
  };
}

export function fromFlowGraph(graph: FlowGraphData): {
  nodes: DesignerRFNode[];
  edges: DesignerRFEdge[];
} {
  return {
    nodes: graph.nodes.map((node) => ({
      id: node.id,
      type: node.type,
      position: node.position,
      data: node.data,
    })),
    edges: graph.edges.map((edge) => ({
      id: edge.id,
      source: edge.source,
      target: edge.target,
      sourceHandle: edge.sourceHandle,
      label: edge.label,
      data: edge.data,
    })),
  };
}

export function createNodeId(type: DesignerNodeType): string {
  return `${type}_${Date.now().toString(36)}_${Math.random().toString(36).slice(2, 6)}`;
}
