import type { NodeTypes } from 'reactflow';
import FlowNodeCard from './FlowNodeCard';

export const designerNodeTypes: NodeTypes = {
  start: FlowNodeCard,
  end: FlowNodeCard,
  approval: FlowNodeCard,
  condition: FlowNodeCard,
  parallel: FlowNodeCard,
  merge: FlowNodeCard,
  timer: FlowNodeCard,
  notify: FlowNodeCard,
};
