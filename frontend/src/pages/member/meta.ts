import type { StatusMap } from '@/types/common';

export const ApplicationStatusMap: StatusMap = {
  0: { color: 'processing', text: '待审核' },
  1: { color: 'warning', text: '审核中' },
  2: { color: 'cyan', text: '面试中' },
  3: { color: 'success', text: '通过' },
  4: { color: 'error', text: '拒绝' },
  5: { color: 'orange', text: '补充材料' },
};

export const ApplicantTypeMap: StatusMap = {
  1: { color: 'blue', text: '会员' },
  2: { color: 'purple', text: '干事' },
};

export const MemberTypeMap: StatusMap = {
  1: { color: 'blue', text: '会员' },
  2: { color: 'purple', text: '干事' },
  3: { color: 'gold', text: '部长' },
  4: { color: 'red', text: '社长' },
};

export const ProfileStatusMap: StatusMap = {
  0: { color: 'success', text: '正常' },
  1: { color: 'error', text: '禁用' },
  2: { color: 'default', text: '已退出' },
};

export const requiredFieldOptions = [
  { label: '申请理由', value: 'reason' },
  { label: '技能', value: 'skills' },
  { label: '项目经历', value: 'experience' },
  { label: '联系电话', value: 'contact_phone' },
  { label: '邮箱', value: 'contact_email' },
];
