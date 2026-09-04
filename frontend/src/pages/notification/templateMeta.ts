export const channelOptions = [
  { label: '站内消息', value: 'in_app' },
  { label: '邮件', value: 'email' },
  { label: 'WebSocket', value: 'websocket' },
];

export const statusMap: Record<number, { color: string; text: string }> = {
  0: { color: 'default', text: '禁用' },
  1: { color: 'success', text: '启用' },
};

export const channelColorMap: Record<string, string> = {
  in_app: 'blue',
  email: 'green',
  websocket: 'purple',
};

export const categorySelectOptions = [
  { label: '系统', value: 'system' },
  { label: '任务', value: 'task' },
  { label: '会议', value: 'meeting' },
  { label: '审批', value: 'approval' },
  { label: '面试', value: 'interview' },
  { label: '其他', value: 'other' },
];
