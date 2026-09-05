import type { StatusMap } from '@/types/common';

export const MeetingStatusMap: StatusMap = {
  0: { color: 'default', text: '待开始' },
  1: { color: 'processing', text: '进行中' },
  2: { color: 'success', text: '已结束' },
  3: { color: 'error', text: '已取消' },
};

export const MeetingTypeMap: Record<number, string> = {
  1: '例会',
  2: '临时会议',
  3: '线上会议',
};

export const VoteStatusMap: StatusMap = {
  0: { color: 'default', text: '未开始' },
  1: { color: 'processing', text: '投票中' },
  2: { color: 'success', text: '已结束' },
  3: { color: 'error', text: '已取消' },
};

export const VoteTypeMap: Record<number, string> = {
  1: '等权投票',
  2: '加权投票',
};
