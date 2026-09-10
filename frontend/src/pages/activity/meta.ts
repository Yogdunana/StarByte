import type { ActivityStatus, RegistrationStatus } from '@/api/activity';

export const ActivityStatusMap: Record<ActivityStatus, { text: string; color: string }> = {
  0: { text: '草稿', color: 'default' },
  1: { text: '报名中', color: 'processing' },
  2: { text: '进行中', color: 'success' },
  3: { text: '已结束', color: 'default' },
  4: { text: '已取消', color: 'error' },
};

export const RegistrationStatusMap: Record<RegistrationStatus, { text: string; color: string }> = {
  0: { text: '待审批', color: 'warning' },
  1: { text: '已通过', color: 'success' },
  2: { text: '已拒绝', color: 'error' },
  3: { text: '候补', color: 'processing' },
  4: { text: '已取消', color: 'default' },
};

export const ActivityCategoryOptions = [
  { label: '竞赛', value: '竞赛' },
  { label: '讲座', value: '讲座' },
  { label: '技术沙龙', value: '技术沙龙' },
  { label: '培训', value: '培训' },
  { label: '团建', value: '团建' },
  { label: '其他', value: '其他' },
];
