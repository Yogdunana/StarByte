import React from 'react';
import { InboxOutlined } from '@ant-design/icons';
import styles from './EmptyState.module.css';

export interface EmptyStateProps {
  image?: React.ReactNode;
  icon?: React.ReactNode;
  title: string;
  description?: string;
  action?: React.ReactNode;
}

const EmptyState: React.FC<EmptyStateProps> = ({
  image,
  icon,
  title,
  description,
  action,
}) => {
  const mark = icon ?? image;
  return (
    <div className={styles.wrap}>
      <div className={styles.icon} aria-hidden="true">
        {mark || <InboxOutlined />}
      </div>
      <p className={styles.title}>{title}</p>
      {description && <p className={styles.description}>{description}</p>}
      {action && <div className={styles.action}>{action}</div>}
    </div>
  );
};

export default EmptyState;
