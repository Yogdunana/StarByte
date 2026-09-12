import { tx } from '@/i18n/text';
import type { StatusMap } from '@/types/common';

export const SessionStatusMap: StatusMap = {
  0: {
    color: 'default',
    get text() {
      return tx('待开始');
    },
  },
  1: {
    color: 'processing',
    get text() {
      return tx('进行中');
    },
  },
  2: {
    color: 'success',
    get text() {
      return tx('已结束');
    },
  },
  3: {
    color: 'error',
    get text() {
      return tx('已取消');
    },
  },
};

export const InterviewStatusMap: StatusMap = {
  0: {
    color: 'default',
    get text() {
      return tx('待面试');
    },
  },
  1: {
    color: 'cyan',
    get text() {
      return tx('已签到');
    },
  },
  2: {
    color: 'processing',
    get text() {
      return tx('面试中');
    },
  },
  3: {
    color: 'success',
    get text() {
      return tx('已完成');
    },
  },
  4: {
    color: 'warning',
    get text() {
      return tx('缺席');
    },
  },
  5: {
    color: 'error',
    get text() {
      return tx('已取消');
    },
  },
};

export const ResultMap: StatusMap = {
  0: {
    color: 'default',
    get text() {
      return tx('未出结果');
    },
  },
  1: {
    color: 'success',
    get text() {
      return tx('通过');
    },
  },
  2: {
    color: 'error',
    get text() {
      return tx('不通过');
    },
  },
  3: {
    color: 'warning',
    get text() {
      return tx('待定');
    },
  },
};
