import { tx } from '@/i18n/text';
import type { StatusMap } from '@/types/common';

export const MeetingStatusMap: StatusMap = {
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

export const MeetingTypeMap: Record<number, string> = {
  get 1() {
    return tx('例会');
  },
  get 2() {
    return tx('临时会议');
  },
  get 3() {
    return tx('线上会议');
  },
};

export const VoteStatusMap: StatusMap = {
  0: {
    color: 'default',
    get text() {
      return tx('未开始');
    },
  },
  1: {
    color: 'processing',
    get text() {
      return tx('投票中');
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

export const VoteTypeMap: Record<number, string> = {
  get 1() {
    return tx('等权投票');
  },
  get 2() {
    return tx('加权投票');
  },
};
