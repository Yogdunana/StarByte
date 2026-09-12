import { tx } from '@/i18n/text';
import type { StatusMap } from '@/types/common';

export const TaskStatusMap: StatusMap = {
  0: {
    color: 'default',
    get text() {
      return tx('待处理');
    },
  },
  1: {
    color: 'processing',
    get text() {
      return tx('进行中');
    },
  },
  2: {
    color: 'success',
    get text() {
      return tx('已完成');
    },
  },
  3: {
    color: 'error',
    get text() {
      return tx('已取消');
    },
  },
  4: {
    color: 'warning',
    get text() {
      return tx('已挂起');
    },
  },
};

export const TaskPriorityMap: StatusMap = {
  0: {
    color: 'default',
    get text() {
      return tx('低');
    },
  },
  1: {
    color: 'blue',
    get text() {
      return tx('中');
    },
  },
  2: {
    color: 'orange',
    get text() {
      return tx('高');
    },
  },
  3: {
    color: 'red',
    get text() {
      return tx('紧急');
    },
  },
};

export const BOARD_COLUMNS: Array<{ status: 0 | 1 | 4 | 2 | 3; title: string }> = [
  {
    status: 0,
    get title() {
      return tx('待处理');
    },
  },
  {
    status: 1,
    get title() {
      return tx('进行中');
    },
  },
  {
    status: 4,
    get title() {
      return tx('已挂起');
    },
  },
  {
    status: 2,
    get title() {
      return tx('已完成');
    },
  },
  {
    status: 3,
    get title() {
      return tx('已取消');
    },
  },
];

export const TaskWorkflowStageMap: Record<string, string> = {
  get assignment() {
    return tx('待分配');
  },
  get execution() {
    return tx('执行中');
  },
  get review() {
    return tx('待审核');
  },
  get acceptance() {
    return tx('待验收');
  },
  get completed() {
    return tx('已验收');
  },
  get cancelled() {
    return tx('已取消');
  },
  get rejected() {
    return tx('已拒绝');
  },
};
