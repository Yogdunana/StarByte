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
      title: '状态',
      dataIndex: 'is_read',
      key: 'is_read',
      width: 70,
      render: (isRead: boolean) =>
        isRead ? <Tag>已读</Tag> : <Badge status="error" text="未读" />,
    },
    {
      title: '标题',
      dataIndex: 'title',
      key: 'title',
      width: 200,
      render: (text: string, record: Notification) => (
        <Text
          strong={!record.is_read}
          style={{ cursor: 'pointer' }}
          onClick={() => handlers.onView(record)}
        >
          {text}
        </Text>
      ),
    },
    {
      title: '分类',
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
      title: '优先级',
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
      title: '发送者',
      key: 'sender',
      width: 100,
      render: (_: unknown, record: Notification) => record.sender?.name || '-',
    },
    {
      title: '时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 160,
      render: (time: string) =>
        new Date(time).toLocaleString('zh-CN', { hour12: false }),
    },
    {
      title: '操作',
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
            查看
          </Button>
          {!record.is_read && (
            <Button
              type="link"
              size="small"
              icon={<CheckOutlined />}
              onClick={() => handlers.onMarkRead(record)}
            >
              已读
            </Button>
          )}
          <Popconfirm
            title="确认删除此通知？"
            onConfirm={() => handlers.onDelete(record)}
          >
            <Button type="link" size="small" danger icon={<DeleteOutlined />}>
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];
}
