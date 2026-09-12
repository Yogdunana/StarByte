import { tx } from '@/i18n/text';
import type { StatusMap } from '@/types/common';

export const ActivityStatusMap: StatusMap = {
  0: {
    color: 'default',
    get text() {
      return tx('草稿');
    },
  },
  1: {
    color: 'processing',
    get text() {
      return tx('报名中');
    },
  },
  2: {
    color: 'success',
    get text() {
      return tx('进行中');
    },
  },
  3: {
    color: 'default',
    get text() {
      return tx('已结束');
    },
  },
  4: {
    color: 'error',
    get text() {
      return tx('已取消');
    },
  },
};

export function registerSuccessText(status: number): string {
  return status === 3 ? tx('已加入候补') : tx('报名成功');
}

export const RegistrationStatusMap: StatusMap = {
  0: {
    color: 'warning',
    get text() {
      return tx('待审批');
    },
  },
  1: {
    color: 'success',
    get text() {
      return tx('已通过');
    },
  },
  2: {
    color: 'error',
    get text() {
      return tx('已拒绝');
    },
  },
  3: {
    color: 'processing',
    get text() {
      return tx('候补');
    },
  },
  4: {
    color: 'default',
    get text() {
      return tx('已取消');
    },
  },
};

export const ActivityCategoryOptions = [
  {
    get label() {
      return tx('竞赛');
    },
    get value() {
      return tx('竞赛');
    },
  },
  {
    get label() {
      return tx('讲座');
    },
    get value() {
      return tx('讲座');
    },
  },
  {
    get label() {
      return tx('技术沙龙');
    },
    get value() {
      return tx('技术沙龙');
    },
  },
  {
    get label() {
      return tx('培训');
    },
    get value() {
      return tx('培训');
    },
  },
  {
    get label() {
      return tx('团建');
    },
    get value() {
      return tx('团建');
    },
  },
  {
    get label() {
      return tx('其他');
    },
    get value() {
      return tx('其他');
    },
  },
];
