import type { StatusMap } from '@/types/common';

export const InternshipStatusMap: StatusMap = {
  0: { color: 'processing', text: '进行中' },
  1: { color: 'success', text: '已完成' },
  2: { color: 'error', text: '已中止' },
};

export const InternshipTypeMap: StatusMap = {
  0: { color: 'blue', text: '校内社团' },
  1: { color: 'cyan', text: '校内其他' },
  2: { color: 'purple', text: '校外实习' },
};
