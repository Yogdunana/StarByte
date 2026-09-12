import { notificationText, notificationSender } from './localizedText';
import i18n from '@/i18n';
import { tx, useLocale } from '@/i18n/text';
import { notificationActionURL } from './actionURL';
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
}) => {
  useLocale();
  return (
    <Drawer title={tx('通知详情')} open={open} onClose={onClose} width={480}>
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
              <Tag>{tx('已读')}</Tag>
            ) : (
              <Badge status="error" text={tx('未读')} />
            )}
          </div>

          <Title level={5}>{notificationText(notification, 'title')}</Title>

          <div style={{ marginBottom: 16 }}>
            <Text type="secondary">
              {tx('发送者：')}
              {notificationSender(notification) || tx('系统')}
            </Text>
            <br />
            <Text type="secondary">
              {tx('时间：')}
              {new Date(notification.created_at).toLocaleString(i18n.language, { hour12: false })}
            </Text>
          </div>

          <Paragraph style={{ whiteSpace: 'pre-wrap' }}>
            {notificationText(notification, 'content')}
          </Paragraph>

          {notificationActionURL(notification) && (
            <Button type="primary" onClick={() => onAction(notification)}>
              {tx('查看详情')}
            </Button>
          )}
        </div>
      )}
    </Drawer>
  );
};

export default NotificationDetailDrawer;
