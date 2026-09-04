import type { StatusMap } from '@/types/common';

export const SessionStatusMap: StatusMap = {
  0: { color: 'default', text: '待开始' },
  1: { color: 'processing', text: '进行中' },
  2: { color: 'success', text: '已结束' },
  3: { color: 'error', text: '已取消' },
};

export const InterviewStatusMap: StatusMap = {
  0: { color: 'default', text: '待面试' },
  1: { color: 'cyan', text: '已签到' },
  2: { color: 'processing', text: '面试中' },
  3: { color: 'success', text: '已完成' },
  4: { color: 'warning', text: '缺席' },
  5: { color: 'error', text: '已取消' },
};

export const ResultMap: StatusMap = {
  0: { color: 'default', text: '未出结果' },
  1: { color: 'success', text: '通过' },
  2: { color: 'error', text: '不通过' },
  3: { color: 'warning', text: '待定' },
};
