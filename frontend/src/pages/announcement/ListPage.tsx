import React, { useCallback, useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Badge, Button, Card, Input, Select, Space, Table, Tag, message } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { useTranslation } from 'react-i18next';
import StatusTag from '@/components/StatusTag/StatusTag';
import { usePermission } from '@/hooks/usePermission';
import { useAnnouncementUnread } from '@/hooks/useAnnouncementUnread';
import {
  archiveAnnouncement,
  createAnnouncement,
  deleteAnnouncement,
  getAnnouncementList,
  pinAnnouncement,
  publishAnnouncement,
  updateAnnouncement,
} from '@/api/announcement';
import type { Announcement, AnnouncementStatus, CreateAnnouncementParams } from '@/api/announcement';
import { formatDateTime } from '@/utils/format';
import FeatureEnabled from '@/components/FeatureEnabled/FeatureEnabled';
import { useFeature } from '@/hooks/useFeature';
import FormModal from './FormModal';
import { AnnouncementCategories, announcementStatusMap } from './meta';

const ListPage: React.FC = () => {
  const { t } = useTranslation();
  const nav = useNavigate();
  const canCreate = usePermission('announcement:create');
  const canUpdate = usePermission('announcement:update');
  const canDelete = usePermission('announcement:delete');
  const canPublish = usePermission('announcement:publish');
  const canManage = usePermission('announcement:manage');
  const { count: unread, refresh: refreshUnread } = useAnnouncementUnread(0);
  const [list, setList] = useState<Announcement[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [status, setStatus] = useState<AnnouncementStatus | undefined>();
  const [category, setCategory] = useState<string | undefined>();
  const [keyword, setKeyword] = useState('');
  const [unreadOnly, setUnreadOnly] = useState(false);
  const [loading, setLoading] = useState(false);
  const [open, setOpen] = useState(false);
  const [editing, setEditing] = useState<Announcement | null>(null);
  const statusMap = announcementStatusMap(t);

  const feed = useFeature('announcement.feed');
  const load = useCallback(async () => {
    if (!feed.enabled) return;
    setLoading(true);
    try {
      const res = await getAnnouncementList({
        page,
        page_size: 10,
        status,
        category,
        keyword,
        unread_only: unreadOnly || undefined,
      });
      setList(res.list);
      setTotal(res.total);
      await refreshUnread();
    } finally {
      setLoading(false);
    }
  }, [page, status, category, keyword, unreadOnly, refreshUnread, feed.enabled]);

  useEffect(() => { void load(); }, [load]);

  const columns: ColumnsType<Announcement> = [
    {
      title: t('announcement.title'),
      dataIndex: 'title',
      render: (v: string, r) => (
        <Space>
          {r.pinned && <Tag color="gold">{t('announcement.pinned')}</Tag>}
          {r.required && <Tag color="red">{t('announcement.required')}</Tag>}
          {!r.is_read && r.status === 1 && <Badge status="processing" />}
          <Button type="link" style={{ padding: 0 }} onClick={() => nav(`/announcement/${r.id}`)}>{v}</Button>
        </Space>
      ),
    },
    {
      title: t('announcement.categoryLabel'),
      dataIndex: 'category',
      width: 120,
      render: (v: string) => t(`announcement.category.${v}`, { defaultValue: v }),
    },
    {
      title: t('announcement.author'),
      key: 'author',
      width: 100,
      render: (_, r) => r.author?.name || '-',
    },
    {
      title: t('announcement.publishedAt'),
      dataIndex: 'published_at',
      width: 170,
      render: (v?: string) => (v ? formatDateTime(v, 'YYYY-MM-DD HH:mm') : '-'),
    },
    {
      title: t('announcement.statusLabel'),
      dataIndex: 'status',
      width: 90,
      render: (v: number) => <StatusTag status={v} mapping={statusMap} />,
    },
    {
      title: t('common.actions'),
      width: 280,
      render: (_, record) => (
        <Space wrap>
          <Button type="link" size="small" onClick={() => nav(`/announcement/${record.id}`)}>
            {t('announcement.detail')}
          </Button>
          {canUpdate && record.status !== 2 && (
            <Button type="link" size="small" onClick={() => { setEditing(record); setOpen(true); }}>
              {t('common.edit')}
            </Button>
          )}
          {canPublish && record.status === 0 && (
            <Button type="link" size="small" onClick={() => publishAnnouncement(record.id).then(() => { message.success(t('announcement.published')); void load(); })}>
              {t('announcement.publish')}
            </Button>
          )}
          {canManage && (
            <Button type="link" size="small" onClick={() => pinAnnouncement(record.id, !record.pinned).then(() => { void load(); })}>
              {record.pinned ? t('announcement.unpin') : t('announcement.pin')}
            </Button>
          )}
          {canManage && record.status === 1 && (
            <Button type="link" size="small" onClick={() => archiveAnnouncement(record.id).then(() => { void load(); })}>
              {t('announcement.archive')}
            </Button>
          )}
          {canDelete && (
            <Button type="link" size="small" danger onClick={() => deleteAnnouncement(record.id).then(() => { message.success(t('common.deleted')); void load(); })}>
              {t('common.delete')}
            </Button>
          )}
        </Space>
      ),
    },
  ];

  return (
    <FeatureEnabled flag="announcement.feed">
    <Card
      title={
        <Space>
          {t('announcement.list')}
          {unread > 0 && <Badge count={unread} />}
        </Space>
      }
      extra={canCreate && (
        <Button type="primary" icon={<PlusOutlined />} onClick={() => { setEditing(null); setOpen(true); }}>
          {t('announcement.create')}
        </Button>
      )}
    >
      <Space style={{ marginBottom: 16 }} wrap>
        <Input.Search
          allowClear
          placeholder={t('announcement.search')}
          onSearch={(v) => { setKeyword(v); setPage(1); }}
        />
        <Select
          allowClear
          placeholder={t('announcement.categoryLabel')}
          style={{ width: 160 }}
          value={category}
          onChange={(v) => { setCategory(v); setPage(1); }}
          options={AnnouncementCategories.map((item) => ({ value: item.value, label: t(item.labelKey) }))}
        />
        <Select
          allowClear
          placeholder={t('announcement.statusLabel')}
          style={{ width: 140 }}
          value={status}
          onChange={(v) => { setStatus(v); setPage(1); }}
          options={Object.entries(statusMap).map(([k, v]) => ({ value: Number(k), label: v.text }))}
        />
        {canCreate && (
          <Button onClick={() => { setStatus(0); setPage(1); }}>{t('announcement.draftBox')}</Button>
        )}
        <Select
          allowClear
          placeholder={t('announcement.unreadFilter')}
          style={{ width: 140 }}
          value={unreadOnly ? true : undefined}
          onChange={(v) => { setUnreadOnly(Boolean(v)); setPage(1); }}
          options={[{ value: true, label: t('announcement.unreadOnly') }]}
        />
      </Space>
      <Table
        rowKey="id"
        loading={loading}
        columns={columns}
        dataSource={list}
        pagination={{ current: page, total, pageSize: 10, onChange: setPage }}
      />
      <FormModal
        open={open}
        editing={editing}
        canSchedule={canPublish}
        canManage={canManage}
        onCancel={() => setOpen(false)}
        onSubmit={async (values) => {
          const payload = {
            title: values.title as string,
            content: (values.content as string) || '',
            content_type: values.content_type as CreateAnnouncementParams['content_type'],
            category: values.category as CreateAnnouncementParams['category'],
            required: values.required as boolean,
            sort_order: values.sort_order as number | undefined,
            scheduled_at: values.scheduled_at as string | undefined,
            expires_at: values.expires_at as string | undefined,
            audience_type: values.audience_type as CreateAnnouncementParams['audience_type'],
            audience_ids: values.audience_ids as string[] | undefined,
            attachments: values.attachments as CreateAnnouncementParams['attachments'],
          };
          if (editing) {
            await updateAnnouncement(editing.id, {
              ...payload,
              clear_scheduled_at: values.clear_scheduled_at as boolean,
              clear_expires_at: values.clear_expires_at as boolean,
            });
            message.success(t('common.saved'));
          } else {
            await createAnnouncement(payload);
            message.success(t('announcement.created'));
          }
          setOpen(false);
          await load();
        }}
      />
    </Card>
    </FeatureEnabled>
  );
};

export default ListPage;
