import React from 'react';
import { Drawer, Tag, Badge, Typography, Button } from 'antd';
import type { Notification, NotificationCategory } from '@/types/api';
import {
  categoryColorMap,
  categoryLabelMap,
  priorityColorMap,
  priorityLabelMap,
} from './notificationMeta';

const { Text, Paragraph, Title } = Typography;

export interface NotificationDetailDrawerProps {
  open: boolean;
  notification: Notification | null;
  onClose: () => void;
  onAction: (record: Notification) => void;
}

const NotificationDetailDrawer: React.FC<NotificationDetailDrawerProps> = ({
  open,
  notification,
  onClose,
  onAction,
}) => (
  <Drawer title="通知详情" open={open} onClose={onClose} width={480}>
    {notification && (
      <div>
        <div style={{ marginBottom: 16, display: 'flex', gap: 8 }}>
          <Tag color={categoryColorMap[notification.category] || 'default'}>
            {categoryLabelMap[notification.category as NotificationCategory] ||
              notification.category}
          </Tag>
          <Tag color={priorityColorMap[notification.priority] || 'default'}>
            {priorityLabelMap[notification.priority] || notification.priority}
          </Tag>
          {notification.is_read ? (
            <Tag>已读</Tag>
          ) : (
            <Badge status="error" text="未读" />
          )}
        </div>

        <Title level={5}>{notification.title}</Title>

        <div style={{ marginBottom: 16 }}>
          <Text type="secondary">发送者：{notification.sender?.name || '系统'}</Text>
          <br />
          <Text type="secondary">
            时间：
            {new Date(notification.created_at).toLocaleString('zh-CN', { hour12: false })}
          </Text>
        </div>

        <Paragraph style={{ whiteSpace: 'pre-wrap' }}>
          {notification.content}
        </Paragraph>

        {notification.action_url && (
          <Button type="primary" onClick={() => onAction(notification)}>
            查看详情
          </Button>
        )}
      </div>
    )}
  </Drawer>
);

export default NotificationDetailDrawer;
