export const categoryOptions = [
  { label: '全部', value: '' },
  { label: '系统', value: 'system' },
  { label: '任务', value: 'task' },
  { label: '会议', value: 'meeting' },
  { label: '审批', value: 'approval' },
  { label: '面试', value: 'interview' },
  { label: '其他', value: 'other' },
];

export const categoryColorMap: Record<string, string> = {
  system: 'blue',
  task: 'green',
  meeting: 'purple',
  approval: 'orange',
  interview: 'cyan',
  other: 'default',
};

export const categoryLabelMap: Record<string, string> = {
  system: '系统',
  task: '任务',
  meeting: '会议',
  approval: '审批',
  interview: '面试',
  other: '其他',
};

export const priorityColorMap: Record<string, string> = {
  urgent: 'red',
  high: 'orange',
  normal: 'blue',
  low: 'default',
};

export const priorityLabelMap: Record<string, string> = {
  urgent: '紧急',
  high: '高',
  normal: '普通',
  low: '低',
};

