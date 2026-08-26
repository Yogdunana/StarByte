import React from 'react';
import { Empty } from 'antd';
import { InboxOutlined } from '@ant-design/icons';

export interface EmptyStateProps {
  image?: React.ReactNode;
  title: string;
  description?: string;
  action?: React.ReactNode;
}

/**
 * 空状态组件 — 自定义图片、标题、描述和操作
 */
const EmptyState: React.FC<EmptyStateProps> = ({
  image,
  title,
  description,
  action,
}) => {
  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        padding: '40px 0',
      }}
    >
      {image || <InboxOutlined style={{ fontSize: 48, color: '#bfbfbf', marginBottom: 16 }} />}
      <Empty
        description={
          <div>
            <div style={{ fontSize: 16, color: '#333', fontWeight: 500 }}>{title}</div>
            {description && (
              <div style={{ color: '#999', fontSize: 13, marginTop: 4 }}>{description}</div>
            )}
          </div>
        }
      >
        {action && <div style={{ marginTop: 8 }}>{action}</div>}
      </Empty>
    </div>
  );
};

export default EmptyState;
