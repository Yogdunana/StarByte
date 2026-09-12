import { tx } from '@/i18n/text';
import type { DesignerNodeType } from '@/types/workflow';

export interface NodePaletteItem {
  type: DesignerNodeType;
  label: string;
  color: string;
  shape: 'circle' | 'rect' | 'diamond';
}

export const NODE_PALETTE: NodePaletteItem[] = [
  {
    type: 'start',
    get label() {
      return tx('开始');
    },
    color: '#52c41a',
    shape: 'circle',
  },
  {
    type: 'end',
    get label() {
      return tx('结束');
    },
    color: '#ff4d4f',
    shape: 'circle',
  },
  {
    type: 'approval',
    get label() {
      return tx('审批');
    },
    color: '#1677ff',
    shape: 'rect',
  },
  {
    type: 'condition',
    get label() {
      return tx('条件');
    },
    color: '#faad14',
    shape: 'diamond',
  },
  {
    type: 'parallel',
    get label() {
      return tx('并行');
    },
    color: '#722ed1',
    shape: 'rect',
  },
  {
    type: 'merge',
    get label() {
      return tx('合并');
    },
    color: '#8c8c8c',
    shape: 'rect',
  },
  {
    type: 'timer',
    get label() {
      return tx('定时器');
    },
    color: '#fa8c16',
    shape: 'rect',
  },
  {
    type: 'notify',
    get label() {
      return tx('通知');
    },
    color: '#13c2c2',
    shape: 'rect',
  },
];

export const NODE_META: Record<DesignerNodeType, NodePaletteItem> = NODE_PALETTE.reduce(
  (acc, item) => {
    acc[item.type] = item;
    return acc;
  },
  {} as Record<DesignerNodeType, NodePaletteItem>,
);

export const CONDITION_VARIABLES = [
  'applicant.name',
  'applicant.dept_id',
  'interview.score',
  'interview.result',
  'initiator.id',
];

export const CONDITION_OPERATORS = ['==', '!=', '>', '<', '>=', '<='] as const;

export const NOTIFY_TEMPLATES = [
  {
    get label() {
      return tx('默认通知');
    },
    value: 'default',
  },
  {
    get label() {
      return tx('审批待办');
    },
    value: 'approval_todo',
  },
  {
    get label() {
      return tx('审批结果');
    },
    value: 'approval_result',
  },
];

export const DND_TYPE = 'application/starbyte-flow-node';

export const HISTORY_LIMIT = 50;
