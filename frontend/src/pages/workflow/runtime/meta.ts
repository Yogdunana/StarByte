import dayjs from 'dayjs';
export const instanceLabels: Record<number, string> = { 0: '进行中', 1: '已完成', 2: '已终止', 3: '已挂起' };
export const taskLabels: Record<number, string> = { 0: '待处理', 1: '已同意', 2: '已拒绝', 3: '已转办', 4: '已撤回', 5: '已取消' };
export const actionLabels: Record<string, string> = { start: '发起流程', approve: '同意', reject: '拒绝', transfer: '转办', rollback: '退回重审', withdraw: '撤回', cancel: '取消待办' };
export const dateLabel = (value?: string) => value ? dayjs(value).format('YYYY-MM-DD HH:mm') : '—';
