import React, { useCallback, useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { Badge, Button, Card, Descriptions, Space, Table, Tag, Typography, message } from 'antd';
import { useTranslation } from 'react-i18next';
import StatusTag from '@/components/StatusTag/StatusTag';
import { usePermission } from '@/hooks/usePermission';
import { useAnnouncementUnread } from '@/hooks/useAnnouncementUnread';
import {
  archiveAnnouncement,
  deleteAnnouncement,
  getAnnouncementDetail,
  getAnnouncementReadStatus,
  markAnnouncementRead,
  pinAnnouncement,
  publishAnnouncement,
} from '@/api/announcement';
import type { Announcement, AnnouncementReadStatus } from '@/api/announcement';
import { formatDateTime } from '@/utils/format';
import { announcementStatusMap } from './meta';
import './announcement.css';

const DetailPage: React.FC = () => {
  const { id = '' } = useParams();
  const nav = useNavigate();
  const { t } = useTranslation();
  const canPublish = usePermission('announcement:publish');
  const canManage = usePermission('announcement:manage');
  const canDelete = usePermission('announcement:delete');
  const { refresh } = useAnnouncementUnread(0);
  const [item, setItem] = useState<Announcement | null>(null);
  const [reads, setReads] = useState<AnnouncementReadStatus | null>(null);

  const load = useCallback(async () => {
    if (!id) return;
    const a = await getAnnouncementDetail(id);
    setItem(a);
    if (a.status === 1 || a.status === 2) {
      if (!a.is_read) {
        await markAnnouncementRead(id);
        setItem({ ...a, is_read: true });
        await refresh();
      }
      try {
        setReads(await getAnnouncementReadStatus(id));
      } catch {
        setReads(null);
      }
    } else {
      setReads(null);
    }
  }, [id, refresh]);

  useEffect(() => { void load(); }, [load]);

  if (!item) return <Card loading />;

  return (
    <Space direction="vertical" size={16} style={{ width: '100%' }}>
      <Card
        title={
          <Space>
            {item.pinned && <Tag color="gold">{t('announcement.pinned')}</Tag>}
            {item.required && <Tag color="red">{t('announcement.required')}</Tag>}
            {!item.is_read && item.status === 1 && <Badge status="processing" text={t('announcement.unread')} />}
            {item.title}
          </Space>
        }
        extra={
          <Space wrap>
            <Button onClick={() => nav('/announcement/list')}>{t('announcement.back')}</Button>
            {canPublish && item.status === 0 && (
              <Button type="primary" onClick={() => publishAnnouncement(item.id).then(() => { message.success(t('announcement.published')); void load(); })}>
                {t('announcement.publish')}
              </Button>
            )}
            {canManage && (
              <Button onClick={() => pinAnnouncement(item.id, !item.pinned).then(() => { void load(); })}>
                {item.pinned ? t('announcement.unpin') : t('announcement.pin')}
              </Button>
            )}
            {canManage && item.status === 1 && (
              <Button onClick={() => archiveAnnouncement(item.id).then(() => { void load(); })}>
                {t('announcement.archive')}
              </Button>
            )}
            {canDelete && (
              <Button danger onClick={() => deleteAnnouncement(item.id).then(() => { message.success(t('common.deleted')); nav('/announcement/list'); })}>
                {t('common.delete')}
              </Button>
            )}
          </Space>
        }
      >
        <Descriptions column={2} size="small">
          <Descriptions.Item label={t('announcement.categoryLabel')}>
            {t(`announcement.category.${item.category}`, { defaultValue: item.category })}
          </Descriptions.Item>
          <Descriptions.Item label={t('announcement.statusLabel')}>
            <StatusTag status={item.status} mapping={announcementStatusMap(t)} />
          </Descriptions.Item>
          <Descriptions.Item label={t('announcement.author')}>{item.author?.name || '-'}</Descriptions.Item>
          <Descriptions.Item label={t('announcement.publishedAt')}>
            {item.published_at ? formatDateTime(item.published_at, 'YYYY-MM-DD HH:mm') : '-'}
          </Descriptions.Item>
        </Descriptions>
        <Typography.Paragraph className="announcement-content">
          {item.content || t('announcement.emptyContent')}
        </Typography.Paragraph>
      </Card>
      {reads && (
        <Card title={t('announcement.readStatus', { read: reads.read_count, unread: reads.unread_count })}>
          <Table
            rowKey={(r) => r.user.id}
            size="small"
            pagination={false}
            dataSource={reads.readers}
            columns={[
              { title: t('announcement.reader'), dataIndex: ['user', 'name'] },
              { title: t('announcement.readAt'), dataIndex: 'read_at', render: (v: string) => formatDateTime(v, 'YYYY-MM-DD HH:mm') },
            ]}
          />
        </Card>
      )}
    </Space>
  );
};

export default DetailPage;
