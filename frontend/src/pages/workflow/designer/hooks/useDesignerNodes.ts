import type { Dispatch, SetStateAction } from 'react';
import type { DesignerNodeType } from '@/types/workflow';
import {
  appendNode,
  createNodeId,
  defaultNodeData,
  type DesignerRFEdge,
  type DesignerRFNode,
} from '../utils/flowConvert';

export function addNodeOfType(
  setNodes: Dispatch<SetStateAction<DesignerRFNode[]>>,
  takeSnapshot: () => void,
  type: DesignerNodeType,
): void {
  takeSnapshot();
  setNodes((current) =>
    appendNode(current, {
      id: createNodeId(type),
      type,
      position: { x: 180, y: 120 },
      data: defaultNodeData(type),
    }),
  );
}

export function applyImportedGraph(
  setNodes: (nodes: DesignerRFNode[]) => void,
  setEdges: (edges: DesignerRFEdge[]) => void,
  takeSnapshot: () => void,
  next: { nodes: DesignerRFNode[]; edges: DesignerRFEdge[] },
): void {
  takeSnapshot();
  setNodes(next.nodes);
  setEdges(next.edges);
}
