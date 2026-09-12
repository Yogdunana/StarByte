import { tx } from '@/i18n/text';
import dayjs from 'dayjs';
export const instanceLabels: Record<number, string> = {
  get 0() {
    return tx('进行中');
  },
  get 1() {
    return tx('已完成');
  },
  get 2() {
    return tx('已终止');
  },
  get 3() {
    return tx('已挂起');
  },
};
export const taskLabels: Record<number, string> = {
  get 0() {
    return tx('待处理');
  },
  get 1() {
    return tx('已同意');
  },
  get 2() {
    return tx('已拒绝');
  },
  get 3() {
    return tx('已转办');
  },
  get 4() {
    return tx('已撤回');
  },
  get 5() {
    return tx('已取消');
  },
};
export const actionLabels: Record<string, string> = {
  get start() {
    return tx('发起流程');
  },
  get approve() {
    return tx('同意');
  },
  get reject() {
    return tx('拒绝');
  },
  get transfer() {
    return tx('转办');
  },
  get rollback() {
    return tx('退回重审');
  },
  get withdraw() {
    return tx('撤回');
  },
  get cancel() {
    return tx('取消待办');
  },
};
export const dateLabel = (value?: string) =>
  value ? dayjs(value).format('YYYY-MM-DD HH:mm') : '—';
