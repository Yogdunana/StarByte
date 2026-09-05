import type { StatusMap } from '@/types/common';

export const TaskStatusMap: StatusMap = {
  0: { color: 'default', text: '待处理' },
  1: { color: 'processing', text: '进行中' },
  2: { color: 'success', text: '已完成' },
  3: { color: 'error', text: '已取消' },
  4: { color: 'warning', text: '已挂起' },
};

export const TaskPriorityMap: StatusMap = {
  0: { color: 'default', text: '低' },
  1: { color: 'blue', text: '中' },
  2: { color: 'orange', text: '高' },
  3: { color: 'red', text: '紧急' },
};

export const BOARD_COLUMNS: Array<{ status: 0 | 1 | 4 | 2 | 3; title: string }> = [
  { status: 0, title: '待处理' },
  { status: 1, title: '进行中' },
  { status: 4, title: '已挂起' },
  { status: 2, title: '已完成' },
  { status: 3, title: '已取消' },
];
