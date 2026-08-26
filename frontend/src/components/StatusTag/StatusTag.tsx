import React from 'react';
import { Tag } from 'antd';
import type { StatusMap } from '@/types/common';

export interface StatusTagProps {
  status: number;
  mapping: StatusMap;
  fallbackText?: string;
}

/**
 * 状态标签 — 通过数字状态码映射到颜色和文本
 */
const StatusTag: React.FC<StatusTagProps> = ({
  status,
  mapping,
  fallbackText = '未知',
}) => {
  const item = mapping[status];
  const color = item?.color || 'default';
  const text = item?.text || fallbackText;

  return <Tag color={color}>{text}</Tag>;
};

export default StatusTag;
