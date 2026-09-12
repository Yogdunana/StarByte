import { notificationText, notificationSender } from './localizedText';
import i18n from '@/i18n';
import { tx } from '@/i18n/text';
import { Button, Space, Tag, Badge, Typography, Popconfirm } from 'antd';
import { CheckOutlined, DeleteOutlined, EyeOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import type { Notification, NotificationCategory } from '@/types/api';
import {
  categoryColorMap,
  categoryLabelMap,
  priorityColorMap,
  priorityLabelMap,
} from './notificationMeta';

const { Text } = Typography;

export interface NotificationColumnHandlers {
  onView: (record: Notification) => void;
  onMarkRead: (record: Notification) => void;
  onDelete: (record: Notification) => void;
}

export function getNotificationColumns(
  handlers: NotificationColumnHandlers,
): ColumnsType<Notification> {
  return [
    {
      title: tx('状态'),
      dataIndex: 'is_read',
      key: 'is_read',
      width: 70,
      render: (isRead: boolean) =>
        isRead ? <Tag>{tx('已读')}</Tag> : <Badge status="error" text={tx('未读')} />,
    },
    {
      title: tx('标题'),
      dataIndex: 'title',
      key: 'title',
      width: 200,
      render: (_text: string, record: Notification) => (
        <Text
          strong={!record.is_read}
          style={{ cursor: 'pointer' }}
          onClick={() => handlers.onView(record)}
        >
          {notificationText(record, 'title')}
        </Text>
      ),
    },
    {
      title: tx('分类'),
      dataIndex: 'category',
      key: 'category',
      width: 90,
      render: (cat: string) => (
        <Tag color={categoryColorMap[cat] || 'default'}>
          {categoryLabelMap[cat as NotificationCategory] || cat}
        </Tag>
      ),
    },
    {
      title: tx('优先级'),
      dataIndex: 'priority',
      key: 'priority',
      width: 80,
      render: (priority: string) => (
        <Tag color={priorityColorMap[priority] || 'default'}>
          {priorityLabelMap[priority] || priority}
        </Tag>
      ),
    },
    {
      title: tx('发送者'),
      key: 'sender',
      width: 100,
      render: (_: unknown, record: Notification) => notificationSender(record) || '-',
    },
    {
      title: tx('时间'),
      dataIndex: 'created_at',
      key: 'created_at',
      width: 160,
      render: (time: string) => new Date(time).toLocaleString(i18n.language, { hour12: false }),
    },
    {
      title: tx('操作'),
      key: 'action',
      width: 150,
      fixed: 'right',
      render: (_: unknown, record: Notification) => (
        <Space>
          <Button
            type="link"
            size="small"
            icon={<EyeOutlined />}
            onClick={() => handlers.onView(record)}
          >
            {tx('查看')}
          </Button>
          {!record.is_read && (
            <Button
              type="link"
              size="small"
              icon={<CheckOutlined />}
              onClick={() => handlers.onMarkRead(record)}
            >
              {tx('已读')}
            </Button>
          )}
          <Popconfirm title={tx('确认删除此通知？')} onConfirm={() => handlers.onDelete(record)}>
            <Button type="link" size="small" danger icon={<DeleteOutlined />}>
              {tx('删除')}
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];
}
