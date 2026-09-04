/** 设计器 8 种节点（短名） */
export type DesignerNodeType =
  | 'start'
  | 'end'
  | 'approval'
  | 'condition'
  | 'parallel'
  | 'merge'
  | 'timer'
  | 'notify';

export type AssigneeStrategy = 'static' | 'role' | 'dept_leader' | 'initiator';
export type ApprovalType = 'single' | 'all' | 'any' | 'ratio';
export type TimerUnit = 'minutes' | 'hours' | 'days';
export type NotifyChannel = 'in_app' | 'email';
export type ConditionOperator = '==' | '!=' | '>' | '<' | '>=' | '<=';

export interface ApprovalConfig {
  assigneeStrategy: AssigneeStrategy;
  assignees?: string[];
  roleId?: string;
  approvalType?: ApprovalType;
  passRatio?: number;
}

export interface ConditionBranch {
  id: string;
  label: string;
  expression: string;
  is_default: boolean;
}

export interface ConditionConfig {
  branches: ConditionBranch[];
}

export interface ParallelConfig {
  branchCount: number;
  branchLabels: string[];
  kind?: 'fork' | 'join';
}

export interface TimerConfig {
  duration: number;
  unit: TimerUnit;
}

export interface NotifyConfig {
  notificationType: string;
  channels: NotifyChannel[];
}

export type NodeConfig =
  | ApprovalConfig
  | ConditionConfig
  | ParallelConfig
  | TimerConfig
  | NotifyConfig
  | Record<string, never>;

export interface DesignerNodeData {
  name: string;
  label: string;
  description?: string;
  config: NodeConfig;
}

export interface FlowGraphPosition {
  x: number;
  y: number;
}

export interface FlowNode {
  id: string;
  type: DesignerNodeType;
  position: FlowGraphPosition;
  data: DesignerNodeData;
}

export interface FlowEdge {
  id: string;
  source: string;
  target: string;
  sourceHandle?: string;
  label?: string;
  data?: { condition?: string };
}

export interface FlowGraphData {
  nodes: FlowNode[];
  edges: FlowEdge[];
}

/** 发布给后端的图（节点 type 可能已映射） */
export interface PublishGraphData {
  nodes: Array<{
    id: string;
    type: string;
    position: FlowGraphPosition;
    data: DesignerNodeData;
  }>;
  edges: FlowEdge[];
}

export interface FlowDefinitionDTO {
  id: string;
  key: string;
  name: string;
  description: string;
  category: string;
  status: 0 | 1 | 2;
  draft_graph?: FlowGraphData | null;
  created_by?: string | null;
  updated_by?: string | null;
  created_at: string;
  updated_at: string;
}

export interface FlowVersionDTO {
  id: string;
  definition_id: string;
  version: number;
  bpmn_data: FlowGraphData;
  status: number;
  published_by?: string | null;
  published_at?: string | null;
  created_at: string;
}

export interface CreateDefinitionPayload {
  key: string;
  name: string;
  description?: string;
  category?: string;
}

export interface ValidationIssue {
  level: 'error' | 'warning';
  message: string;
}
