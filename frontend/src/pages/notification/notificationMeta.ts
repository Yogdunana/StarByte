import { tx } from '@/i18n/text';
export const categoryOptions = [
  {
    get label() {
      return tx('入会申请');
    },
    value: 'member',
  },
  {
    get label() {
      return tx('全部');
    },
    value: '',
  },
  {
    get label() {
      return tx('系统');
    },
    value: 'system',
  },
  {
    get label() {
      return tx('任务');
    },
    value: 'task',
  },
  {
    get label() {
      return tx('会议');
    },
    value: 'meeting',
  },
  {
    get label() {
      return tx('审批');
    },
    value: 'approval',
  },
  {
    get label() {
      return tx('面试');
    },
    value: 'interview',
  },
  {
    get label() {
      return tx('公告');
    },
    value: 'announcement',
  },
  {
    get label() {
      return tx('其他');
    },
    value: 'other',
  },
];

export const categoryColorMap: Record<string, string> = {
  member: 'cyan',
  system: 'blue',
  task: 'green',
  meeting: 'purple',
  approval: 'orange',
  interview: 'cyan',
  announcement: 'gold',
  other: 'default',
};

export const categoryLabelMap: Record<string, string> = {
  get member() {
    return tx('入会申请');
  },
  get system() {
    return tx('系统');
  },
  get task() {
    return tx('任务');
  },
  get meeting() {
    return tx('会议');
  },
  get approval() {
    return tx('审批');
  },
  get interview() {
    return tx('面试');
  },
  get announcement() {
    return tx('公告');
  },
  get other() {
    return tx('其他');
  },
};

export const priorityColorMap: Record<string, string> = {
  urgent: 'red',
  high: 'orange',
  normal: 'blue',
  low: 'default',
};

export const priorityLabelMap: Record<string, string> = {
  get urgent() {
    return tx('紧急');
  },
  get high() {
    return tx('高');
  },
  get normal() {
    return tx('普通');
  },
  get low() {
    return tx('低');
  },
};
