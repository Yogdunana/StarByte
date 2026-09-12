import React from 'react';
import { Skeleton, Space } from 'antd';
import type { BreadcrumbItem } from '@/types/common';
import styles from './PageContainer.module.css';

export interface PageContainerProps {
  title?: string;
  breadcrumb?: BreadcrumbItem[];
  extra?: React.ReactNode;
  loading?: boolean;
  children: React.ReactNode;
}

const PageContainer: React.FC<PageContainerProps> = ({
  title,
  breadcrumb,
  extra,
  loading = false,
  children,
}) => {
  return (
    <div className={styles.page}>
      {(title || breadcrumb || extra) && (
        <div className={styles.bar}>
          <div>
            {breadcrumb && breadcrumb.length > 0 && (
              <div className={styles.crumbs}>
                {breadcrumb.map((item, index) => (
                  <React.Fragment key={index}>
                    {index > 0 && <span className={styles.sep}>/</span>}
                    <span>{item.label}</span>
                  </React.Fragment>
                ))}
              </div>
            )}
            {title && <h2 className={styles.title}>{title}</h2>}
          </div>
          {extra && <Space>{extra}</Space>}
        </div>
      )}
      {loading ? <Skeleton active paragraph={{ rows: 6 }} /> : children}
    </div>
  );
};

export default PageContainer;
