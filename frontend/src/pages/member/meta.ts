import { tx } from '@/i18n/text';
import type { StatusMap } from '@/types/common';

export const ApplicationStatusMap: StatusMap = {
  0: {
    color: 'processing',
    get text() {
      return tx('待审核');
    },
  },
  1: {
    color: 'warning',
    get text() {
      return tx('审核中');
    },
  },
  2: {
    color: 'cyan',
    get text() {
      return tx('面试中');
    },
  },
  3: {
    color: 'success',
    get text() {
      return tx('通过');
    },
  },
  4: {
    color: 'error',
    get text() {
      return tx('拒绝');
    },
  },
  5: {
    color: 'orange',
    get text() {
      return tx('补充材料');
    },
  },
};

export const ApplicantTypeMap: StatusMap = {
  1: {
    color: 'blue',
    get text() {
      return tx('会员');
    },
  },
  2: {
    color: 'purple',
    get text() {
      return tx('干事');
    },
  },
};

export const MemberTypeMap: StatusMap = {
  1: {
    color: 'blue',
    get text() {
      return tx('会员');
    },
  },
  2: {
    color: 'purple',
    get text() {
      return tx('干事');
    },
  },
  3: {
    color: 'gold',
    get text() {
      return tx('部长');
    },
  },
  4: {
    color: 'red',
    get text() {
      return tx('社长');
    },
  },
};

export const ProfileStatusMap: StatusMap = {
  3: {
    color: 'processing',
    get text() {
      return tx('候补期');
    },
  },
  0: {
    color: 'success',
    get text() {
      return tx('正常');
    },
  },
  1: {
    color: 'error',
    get text() {
      return tx('禁用');
    },
  },
  2: {
    color: 'default',
    get text() {
      return tx('已退出');
    },
  },
};

export const requiredFieldOptions = [
  {
    get label() {
      return tx('申请理由');
    },
    value: 'reason',
  },
  {
    get label() {
      return tx('技能');
    },
    value: 'skills',
  },
  {
    get label() {
      return tx('项目经历');
    },
    value: 'experience',
  },
  {
    get label() {
      return tx('联系电话');
    },
    value: 'contact_phone',
  },
  {
    get label() {
      return tx('邮箱');
    },
    value: 'contact_email',
  },
];
