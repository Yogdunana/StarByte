import { tx } from '@/i18n/text';
import type { StatusMap } from '@/types/common';

export const InternshipStatusMap: StatusMap = {
  0: {
    color: 'processing',
    get text() {
      return tx('进行中');
    },
  },
  1: {
    color: 'success',
    get text() {
      return tx('已完成');
    },
  },
  2: {
    color: 'error',
    get text() {
      return tx('已中止');
    },
  },
};

export const InternshipTypeMap: StatusMap = {
  0: {
    color: 'blue',
    get text() {
      return tx('校内社团');
    },
  },
  1: {
    color: 'cyan',
    get text() {
      return tx('校内其他');
    },
  },
  2: {
    color: 'purple',
    get text() {
      return tx('校外实习');
    },
  },
};
