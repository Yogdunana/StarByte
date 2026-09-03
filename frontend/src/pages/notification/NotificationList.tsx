import React, { useState, useEffect, useCallback } from 'react';
import {
  Card,
  Table,
  Button,
  Select,
  Space,
  Tag,
  Drawer,
  Typography,
  message,
  Popconfirm,
  Switch,
  Badge,
} from 'antd';
import {
  CheckOutlined,
  DeleteOutlined,
  ReloadOutlined,
  EyeOutlined,
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { useDispatch } from 'react-redux';
import { useNavigate } from 'react-router-dom';
import type { AppDispatch } from '@/store';
import {
  fetchUnreadCount,
  markAllNotificationsAsRead,
} from '@/store/slices/notificationSlice';
import {
  getNotificationList,
  markAsRead,
  markAllAsRead,
  deleteNotification,
} from '@/api/notification';
import type { Notification, NotificationCategory } from '@/types/api';

const { Text, Paragraph, Title } = Typography;

/** 分类选项 */
const categoryOptions = [
  { label: '全部', value: '' },
  { label: '系统', value: 'system' },
  { label: '任务', value: 'task' },
  { label: '会议', value: 'meeting' },
  { label: '审批', value: 'approval' },
  { label: '面试', value: 'interview' },
  { label: '其他', value: 'other' },
];

/** 分类标签颜色映射 */
const categoryColorMap: Record<string, string> = {
  system: 'blue',
  task: 'green',
  meeting: 'purple',
  approval: 'orange',
  interview: 'cyan',
  other: 'default',
};

/** 分类中文映射 */
const categoryLabelMap: Record<string, string> = {
  system: '系统',
  task: '任务',
  meeting: '会议',
  approval: '审批',
  interview: '面试',
  other: '其他',
};

/** 优先级颜色映射 */
const priorityColorMap: Record<string, string> = {
  urgent: 'red',
  high: 'orange',
  normal: 'blue',
  low: 'default',
};

/** 优先级中文映射 */
const priorityLabelMap: Record<string, string> = {
  urgent: '紧急',
  high: '高',
  normal: '普通',
  low: '低',
};

const NotificationList: React.FC = () => {
  const dispatch = useDispatch<AppDispatch>();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [data, setData] = useState<Notification[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [category, setCategory] = useState<string>('');
  const [unreadOnly, setUnreadOnly] = useState(false);
  const [drawerVisible, setDrawerVisible] = useState(false);
  const [currentNotification, setCurrentNotification] = useState<Notification | null>(null);

  const loadData = useCallback(async () => {
    setLoading(true);
    try {
      const res = await getNotificationList({
        page,
        page_size: pageSize,
        category: (category || undefined) as NotificationCategory,
        unread_only: unreadOnly || undefined,
      });
      setData(res.list);
      setTotal(res.total);
    } catch {
      message.error('加载通知列表失败');
    } finally {
      setLoading(false);
    }
  }, [page, pageSize, category, unreadOnly]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  // 标记单条已读
  const handleMarkRead = async (record: Notification) => {
    try {
      await markAsRead(record.id);
      // 更新本地数据
      setData((prev) =>
        prev.map((item) =>
          item.id === record.id ? { ...item, is_read: true } : item,
        ),
      );
      dispatch(fetchUnreadCount());
      message.success('已标记为已读');
    } catch {
      message.error('标记已读失败');
    }
  };

  // 全部已读
  const handleMarkAllRead = async () => {
    try {
      await markAllAsRead(category || undefined);
      dispatch(markAllNotificationsAsRead());
      loadData();
      message.success('已全部标记为已读');
    } catch {
      message.error('操作失败');
    }
  };

  // 删除
  const handleDelete = async (record: Notification) => {
    try {
      await deleteNotification(record.id);
      setData((prev) => prev.filter((item) => item.id !== record.id));
      setTotal((prev) => prev - 1);
      dispatch(fetchUnreadCount());
      message.success('删除成功');
    } catch {
      message.error('删除失败');
    }
  };

  // 查看详情
  const handleViewDetail = (record: Notification) => {
    setCurrentNotification(record);
    setDrawerVisible(true);
    if (!record.is_read) {
      handleMarkRead(record);
    }
  };

  // 点击通知跳转
  const handleAction = (record: Notification) => {
    if (record.action_url) {
      navigate(record.action_url);
    }
  };

  const columns: ColumnsType<Notification> = [
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
          onClick={() => handleViewDetail(record)}
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
      render: (_: unknown, record: Notification) =>
        record.sender?.name || '-',
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
            onClick={() => handleViewDetail(record)}
          >
            查看
          </Button>
          {!record.is_read && (
            <Button
              type="link"
              size="small"
              icon={<CheckOutlined />}
              onClick={() => handleMarkRead(record)}
            >
              已读
            </Button>
          )}
          <Popconfirm
            title="确认删除此通知？"
            onConfirm={() => handleDelete(record)}
          >
            <Button type="link" size="small" danger icon={<DeleteOutlined />}>
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <Card>
      {/* 工具栏 */}
      <div style={{ marginBottom: 16, display: 'flex', gap: 12, alignItems: 'center' }}>
        <Select
          value={category}
          onChange={(val) => {
            setCategory(val);
            setPage(1);
          }}
          style={{ width: 120 }}
          options={categoryOptions}
        />
        <Space>
          <Text>仅未读</Text>
          <Switch
            checked={unreadOnly}
            onChange={(checked) => {
              setUnreadOnly(checked);
              setPage(1);
            }}
          />
        </Space>
        <Button icon={<ReloadOutlined />} onClick={loadData}>
          刷新
        </Button>
        <div style={{ flex: 1 }} />
        <Button
          type="primary"
          icon={<CheckOutlined />}
          onClick={handleMarkAllRead}
          disabled={total === 0}
        >
          全部已读
        </Button>
      </div>

      {/* 表格 */}
      <Table
        columns={columns}
        dataSource={data}
        rowKey="id"
        loading={loading}
        scroll={{ x: 900 }}
        pagination={{
          current: page,
          pageSize,
          total,
          showSizeChanger: true,
          showQuickJumper: true,
          showTotal: (t) => `共 ${t} 条`,
          onChange: (p, ps) => {
            setPage(p);
            setPageSize(ps);
          },
        }}
      />

      {/* 详情抽屉 */}
      <Drawer
        title="通知详情"
        open={drawerVisible}
        onClose={() => setDrawerVisible(false)}
        width={480}
      >
        {currentNotification && (
          <div>
            <div style={{ marginBottom: 16, display: 'flex', gap: 8 }}>
              <Tag color={categoryColorMap[currentNotification.category] || 'default'}>
                {categoryLabelMap[currentNotification.category as NotificationCategory] || currentNotification.category}
              </Tag>
              <Tag color={priorityColorMap[currentNotification.priority] || 'default'}>
                {priorityLabelMap[currentNotification.priority] || currentNotification.priority}
              </Tag>
              {currentNotification.is_read ? (
                <Tag>已读</Tag>
              ) : (
                <Badge status="error" text="未读" />
              )}
            </div>

            <Title level={5}>{currentNotification.title}</Title>

            <div style={{ marginBottom: 16 }}>
              <Text type="secondary">
                发送者：{currentNotification.sender?.name || '系统'}
              </Text>
              <br />
              <Text type="secondary">
                时间：{new Date(currentNotification.created_at).toLocaleString('zh-CN', { hour12: false })}
              </Text>
            </div>

            <Paragraph style={{ whiteSpace: 'pre-wrap' }}>
              {currentNotification.content}
            </Paragraph>

            {currentNotification.action_url && (
              <Button
                type="primary"
                onClick={() => handleAction(currentNotification)}
              >
                查看详情
              </Button>
            )}
          </div>
        )}
      </Drawer>
    </Card>
  );
};

export default NotificationList;
