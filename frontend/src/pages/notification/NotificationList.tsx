import React, { useState, useEffect, useCallback } from 'react';
import { Card, Table, Button, Select, Space, Typography, message, Switch } from 'antd';
import { CheckOutlined, ReloadOutlined } from '@ant-design/icons';
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
import { categoryOptions } from './notificationMeta';
import { getNotificationColumns } from './notificationColumns';
import NotificationDetailDrawer from './NotificationDetailDrawer';

const { Text } = Typography;

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

  const handleMarkRead = async (record: Notification) => {
    try {
      await markAsRead(record.id);
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

  const handleViewDetail = (record: Notification) => {
    setCurrentNotification(record);
    setDrawerVisible(true);
    if (!record.is_read) {
      handleMarkRead(record);
    }
  };

  const handleAction = (record: Notification) => {
    if (record.action_url) {
      navigate(record.action_url);
    }
  };

  const columns = getNotificationColumns({
    onView: handleViewDetail,
    onMarkRead: handleMarkRead,
    onDelete: handleDelete,
  });

  return (
    <Card>
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

      <NotificationDetailDrawer
        open={drawerVisible}
        notification={currentNotification}
        onClose={() => setDrawerVisible(false)}
        onAction={handleAction}
      />
    </Card>
  );
};

export default NotificationList;
