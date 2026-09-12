import { notificationText } from '@/pages/notification/localizedText';
import i18n from '@/i18n';
import { tx, useLocale } from '@/i18n/text';
import { notificationActionURL } from '@/pages/notification/actionURL';
import React, { useCallback } from 'react';
import { Badge, Popover, List, Typography, Button, Empty, Tag, Tooltip } from 'antd';
import { BellOutlined, CheckOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useDispatch, useSelector } from 'react-redux';
import type { AppDispatch } from '@/store';
import {
  selectUnreadCount,
  selectRecentNotifications,
  selectWSConnected,
  fetchRecentNotifications,
  markAllNotificationsAsRead,
  markNotificationAsRead,
} from '@/store/slices/notificationSlice';
import type { Notification, NotificationPriority, NotificationCategory } from '@/types/api';

const { Text } = Typography;

/** 优先级标签颜色映射 */
const priorityColorMap: Record<string, string> = {
  urgent: 'red',
  high: 'orange',
  normal: 'blue',
  low: 'default',
};

/** 分类中文映射 */
const categoryLabelMap: Record<string, string> = {
  get system() {
    return tx('系统');
  },
  get task() {
    return tx('任务');
  },
  get meeting() {
    return tx('会议');
  },
  get approval() {
    return tx('审批');
  },
  get member() {
    return tx('入会');
  },
  get interview() {
    return tx('面试');
  },
  get announcement() {
    return tx('公告');
  },
  get other() {
    return tx('其他');
  },
};

/** 格式化时间为相对时间 */
function formatRelativeTime(dateStr: string): string {
  if (!dateStr) return '';
  const date = new Date(dateStr);
  const now = Date.now();
  const diff = now - date.getTime();
  const minutes = Math.floor(diff / 60000);
  const hours = Math.floor(diff / 3600000);
  const days = Math.floor(diff / 86400000);

  if (minutes < 1) return tx('刚刚');
  if (minutes < 60) return tx('{{value0}} 分钟前', { value0: minutes });
  if (hours < 24) return tx('{{value0}} 小时前', { value0: hours });
  if (days < 7) return tx('{{value0}} 天前', { value0: days });
  return date.toLocaleDateString(i18n.language);
}

const NotificationBell: React.FC = () => {
  useLocale();
  const dispatch = useDispatch<AppDispatch>();
  const navigate = useNavigate();
  const unreadCount = useSelector(selectUnreadCount);
  const recentNotifications = useSelector(selectRecentNotifications);
  const wsConnected = useSelector(selectWSConnected);

  const handleMarkRead = useCallback(
    (id: string, e: React.MouseEvent) => {
      e.stopPropagation();
      dispatch(markNotificationAsRead(id));
    },
    [dispatch],
  );

  const handleMarkAllRead = useCallback(() => {
    dispatch(markAllNotificationsAsRead());
  }, [dispatch]);

  const handleViewAll = useCallback(() => {
    navigate('/notification/list');
  }, [navigate]);

  const handleNotificationClick = useCallback(
    (notification: Notification) => {
      navigate(notificationActionURL(notification) || '/notification/list');
      // 标记已读
      if (!notification.is_read) {
        dispatch(markNotificationAsRead(notification.id));
      }
    },
    [dispatch, navigate],
  );

  const renderNotificationItem = (item: Notification) => (
    <List.Item
      style={{
        cursor: 'pointer',
        padding: '8px 12px',
        background: item.is_read ? 'transparent' : 'var(--sb-brand-soft)',
      }}
      onClick={() => handleNotificationClick(item)}
    >
      <div style={{ width: '100%' }}>
        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            marginBottom: 4,
          }}
        >
          <Button
            type="text"
            style={{ height: 'auto', whiteSpace: 'normal', textAlign: 'left', padding: 0 }}
            onClick={(event) => {
              event.stopPropagation();
              handleNotificationClick(item);
            }}
          >
            <Text strong={!item.is_read}>{notificationText(item, 'title')}</Text>
          </Button>
          {!item.is_read && (
            <Tooltip title={tx('标记已读')}>
              <Button
                type="text"
                size="small"
                aria-label={tx('标记已读')}
                icon={<CheckOutlined />}
                onClick={(e) => handleMarkRead(item.id, e)}
              />
            </Tooltip>
          )}
        </div>
        <Text type="secondary" style={{ fontSize: 12, display: 'block' }} ellipsis>
          {notificationText(item, 'content')}
        </Text>
        <div style={{ display: 'flex', gap: 6, marginTop: 4, alignItems: 'center' }}>
          <Tag
            color={priorityColorMap[item.priority as NotificationPriority] || 'default'}
            style={{ fontSize: 11 }}
          >
            {categoryLabelMap[item.category as NotificationCategory] || item.category}
          </Tag>
          <Text type="secondary" style={{ fontSize: 11 }}>
            {formatRelativeTime(item.created_at)}
          </Text>
        </div>
      </div>
    </List.Item>
  );

  const content = (
    <div style={{ width: 'min(360px, calc(100vw - 48px))' }}>
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          padding: '4px 4px 8px',
        }}
      >
        <Text strong>{tx('消息通知')}</Text>
        <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
          <Tooltip title={wsConnected ? tx('已连接') : tx('未连接')}>
            <Badge status={wsConnected ? 'success' : 'default'} />
          </Tooltip>
          {unreadCount > 0 && (
            <Button type="link" size="small" icon={<CheckOutlined />} onClick={handleMarkAllRead}>
              {tx('全部已读')}
            </Button>
          )}
        </div>
      </div>
      {recentNotifications.length === 0 ? (
        <Empty
          description={tx('暂无通知')}
          image={Empty.PRESENTED_IMAGE_SIMPLE}
          style={{ padding: 20 }}
        />
      ) : (
        <List
          dataSource={recentNotifications.slice(0, 10)}
          renderItem={renderNotificationItem}
          style={{ maxHeight: 400, overflowY: 'auto' }}
        />
      )}
      <div style={{ textAlign: 'center', borderTop: '1px solid #f0f0f0', padding: '8px 0' }}>
        <Button type="link" onClick={handleViewAll} style={{ fontSize: 13 }}>
          {tx('查看全部通知')}
        </Button>
      </div>
    </div>
  );

  return (
    <Popover
      content={content}
      trigger="click"
      placement="bottomRight"
      arrow={false}
      onOpenChange={(open) => {
        if (open) {
          dispatch(fetchRecentNotifications());
        }
      }}
    >
      <Tooltip title={tx('消息通知')}>
        <Badge count={unreadCount} size="small" offset={[-2, 2]}>
          <Button
            type="text"
            aria-label={tx('消息通知')}
            icon={<BellOutlined style={{ fontSize: 18 }} />}
          />
        </Badge>
      </Tooltip>
    </Popover>
  );
};

export default NotificationBell;
