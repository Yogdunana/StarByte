import React from 'react';
import { Skeleton, Space } from 'antd';
import type { BreadcrumbItem } from '@/types/common';

export interface PageContainerProps {
  title?: string;
  breadcrumb?: BreadcrumbItem[];
  extra?: React.ReactNode;
  loading?: boolean;
  children: React.ReactNode;
}

/**
 * 页面容器 — 统一标题、面包屑、右侧操作区和 loading 状态
 */
const PageContainer: React.FC<PageContainerProps> = ({
  title,
  breadcrumb,
  extra,
  loading = false,
  children,
}) => {
  return (
    <div style={{ width: '100%' }}>
      {(title || breadcrumb || extra) && (
        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            marginBottom: 16,
          }}
        >
          <div>
            {breadcrumb && breadcrumb.length > 0 && (
              <div style={{ marginBottom: 4, color: '#999', fontSize: 13 }}>
                {breadcrumb.map((item, index) => (
                  <React.Fragment key={index}>
                    {index > 0 && <span style={{ margin: '0 4px' }}>/</span>}
                    <span>{item.label}</span>
                  </React.Fragment>
                ))}
              </div>
            )}
            {title && (
              <h2 style={{ margin: 0, fontSize: 20, fontWeight: 600 }}>{title}</h2>
            )}
          </div>
          {extra && <Space>{extra}</Space>}
        </div>
      )}
      {loading ? <Skeleton active paragraph={{ rows: 6 }} /> : children}
    </div>
  );
};

export default PageContainer;
