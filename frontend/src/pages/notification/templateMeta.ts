import { tx } from '@/i18n/text';
export const channelOptions = [
  {
    get label() {
      return tx('站内消息');
    },
    value: 'in_app',
  },
  {
    get label() {
      return tx('邮件');
    },
    value: 'email',
  },
  { label: 'WebSocket', value: 'websocket' },
];

export const statusMap: Record<number, { color: string; text: string }> = {
  0: {
    color: 'default',
    get text() {
      return tx('禁用');
    },
  },
  1: {
    color: 'success',
    get text() {
      return tx('启用');
    },
  },
};

export const channelColorMap: Record<string, string> = {
  in_app: 'blue',
  email: 'green',
  websocket: 'purple',
};

export const categorySelectOptions = [
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
      return tx('其他');
    },
    value: 'other',
  },
];
