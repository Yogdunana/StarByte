import type { StatusMap } from '@/types/common';

export const ActivityStatusMap: StatusMap = {
  0: { color: 'default', text: '草稿' },
  1: { color: 'processing', text: '报名中' },
  2: { color: 'success', text: '进行中' },
  3: { color: 'default', text: '已结束' },
  4: { color: 'error', text: '已取消' },
};

export const RegistrationStatusMap: StatusMap = {
  0: { color: 'warning', text: '待审批' },
  1: { color: 'success', text: '已通过' },
  2: { color: 'error', text: '已拒绝' },
  3: { color: 'processing', text: '候补' },
  4: { color: 'default', text: '已取消' },
};

export const ActivityCategoryOptions = [
  { label: '竞赛', value: '竞赛' },
  { label: '讲座', value: '讲座' },
  { label: '技术沙龙', value: '技术沙龙' },
  { label: '培训', value: '培训' },
  { label: '团建', value: '团建' },
  { label: '其他', value: '其他' },
];
