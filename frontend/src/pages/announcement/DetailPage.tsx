import React, { useCallback, useEffect, useRef, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { Badge, Button, Card, Descriptions, Space, Table, Tabs, Tag, Typography, message } from 'antd';
import { PaperClipOutlined } from '@ant-design/icons';
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
import type { Announcement, AnnouncementAttachment, AnnouncementReadStatus } from '@/api/announcement';
import { getFileDetail } from '@/api/file';
import { formatDateTime } from '@/utils/format';
import { announcementStatusMap, sanitizeAnnouncementHTML } from './meta';
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
  const openedAt = useRef(Date.now());
  const reportForID = useRef('');

  const load = useCallback(async () => {
    if (!id) return;
    const a = await getAnnouncementDetail(id);
    if (a.status === 1 || a.status === 2) {
      reportForID.current = id;
    }
    setItem(a);
    if (a.status === 1 || a.status === 2) {
      if (!a.is_read) {
        await markAnnouncementRead(id);
        setItem({ ...a, is_read: true });
        await refresh();
      }
      if (canManage) {
        try {
          setReads(await getAnnouncementReadStatus(id));
        } catch {
          setReads(null);
        }
      } else {
        setReads(null);
      }
    } else {
      setReads(null);
    }
  }, [id, refresh, canManage]);

  useEffect(() => { void load(); }, [load]);

  useEffect(() => {
    openedAt.current = Date.now();
    const watching = id;
    return () => {
      if (!watching || reportForID.current !== watching) return;
      const seconds = Math.round((Date.now() - openedAt.current) / 1000);
      if (seconds > 0) void markAnnouncementRead(watching, seconds);
    };
  }, [id]);

  const handleDownload = async (file: AnnouncementAttachment) => {
    try {
      // 鉴权取预签名地址后走顶层导航，避免 fetch blob 撞对象存储 CORS。
      const detail = await getFileDetail(file.file_id);
      if (!detail.url) {
        message.error(t('announcement.downloadFailed'));
        return;
      }
      window.open(detail.url, '_blank', 'noopener,noreferrer');
    } catch {
      message.error(t('announcement.downloadFailed'));
    }
  };

  if (!item) return <Card loading />;

  const content = item.content || t('announcement.emptyContent');

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
          <Descriptions.Item label={t('announcement.expiresAt')}>
            {item.expires_at ? formatDateTime(item.expires_at, 'YYYY-MM-DD HH:mm') : '-'}
          </Descriptions.Item>
          <Descriptions.Item label={t('announcement.audience')}>
            {t(`announcement.audienceType.${item.audience_type || 'all'}`, { defaultValue: item.audience_type || 'all' })}
          </Descriptions.Item>
        </Descriptions>
        {item.content_type === 'html' ? (
          <div
            className="announcement-html"
            // HTML 仅走白名单消毒后渲染
            dangerouslySetInnerHTML={{ __html: sanitizeAnnouncementHTML(item.content) }}
          />
        ) : (
          <Typography.Paragraph className="announcement-content">{content}</Typography.Paragraph>
        )}
        {item.attachments?.length > 0 && (
          <div className="announcement-attachments">
            <Typography.Text type="secondary">{t('announcement.attachments')}</Typography.Text>
            <ul>
              {item.attachments.map((file) => (
                <li key={file.file_id}>
                  <Button type="link" icon={<PaperClipOutlined />} onClick={() => void handleDownload(file)}>
                    {file.name}
                  </Button>
                </li>
              ))}
            </ul>
          </div>
        )}
      </Card>
      {reads && (
        <Card
          title={t('announcement.readStatus', { read: reads.read_count, unread: reads.unread_count })}
          extra={t('announcement.readStats', {
            rate: Math.round((reads.read_rate || 0) * 100),
            avg: Math.round(reads.avg_duration_seconds || 0),
          })}
        >
          <Tabs
            items={[
              {
                key: 'read',
                label: t('announcement.readersTab'),
                children: (
                  <Table
                    rowKey={(r) => r.user.id}
                    size="small"
                    pagination={false}
                    dataSource={reads.readers}
                    columns={[
                      { title: t('announcement.reader'), dataIndex: ['user', 'name'] },
                      { title: t('announcement.readAt'), dataIndex: 'read_at', render: (v: string) => formatDateTime(v, 'YYYY-MM-DD HH:mm') },
                      { title: t('announcement.duration'), dataIndex: 'duration_seconds', render: (v: number) => `${v || 0}s` },
                    ]}
                  />
                ),
              },
              {
                key: 'unread',
                label: t('announcement.unreadTab'),
                children: (
                  <Table
                    rowKey="id"
                    size="small"
                    pagination={false}
                    dataSource={reads.unread_users || []}
                    columns={[{ title: t('announcement.reader'), dataIndex: 'name' }]}
                  />
                ),
              },
            ]}
          />
        </Card>
      )}
    </Space>
  );
};

export default DetailPage;
