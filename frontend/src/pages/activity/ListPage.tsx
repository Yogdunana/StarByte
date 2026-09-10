import React, { useCallback, useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Button, Card, Input, Select, Space, Table, message } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import StatusTag from '@/components/StatusTag/StatusTag';
import { usePermission } from '@/hooks/usePermission';
import {
  cancelActivity, createActivity, deleteActivity, endActivity, getActivityList,
  startActivity, updateActivity,
} from '@/api/activity';
import type { Activity, ActivityStatus } from '@/api/activity';
import { ActivityStatusMap } from './meta';
import FormModal from './FormModal';

const ListPage: React.FC = () => {
  const nav = useNavigate();
  const canCreate = usePermission('activity:create');
  const canManage = usePermission('activity:manage');
  const [list, setList] = useState<Activity[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [status, setStatus] = useState<ActivityStatus | undefined>();
  const [keyword, setKeyword] = useState('');
  const [loading, setLoading] = useState(false);
  const [open, setOpen] = useState(false);
  const [editing, setEditing] = useState<Activity | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const res = await getActivityList({ page, page_size: 10, status, keyword });
      setList(res.list);
      setTotal(res.total);
    } finally {
      setLoading(false);
    }
  }, [page, status, keyword]);

  useEffect(() => { void load(); }, [load]);

  const columns: ColumnsType<Activity> = [
    { title: '标题', dataIndex: 'title', render: (v: string, r) => <Button type="link" style={{ padding: 0 }} onClick={() => nav(`/activity/${r.id}`)}>{v}</Button> },
    { title: '分类', dataIndex: 'category', width: 100, render: (v?: string) => v || '-' },
    { title: '地点', dataIndex: 'location', width: 140, render: (v?: string) => v || '-' },
    { title: '开始', dataIndex: 'start_time', width: 160, render: (v: string) => v?.replace('T', ' ').slice(0, 16) },
    { title: '组织者', key: 'org', width: 100, render: (_, r) => r.organizer?.name || '-' },
    {
      title: '报名/上限',
      key: 'n',
      width: 90,
      render: (_, r) => `${r.registered_count}/${r.max_participants === 0 ? '不限' : r.max_participants}`,
    },
    { title: '状态', dataIndex: 'status', width: 90, render: (v: number) => <StatusTag status={v} mapping={ActivityStatusMap} /> },
    {
      title: '操作',
      width: 280,
      render: (_, record) => (
        <Space wrap>
          <Button type="link" size="small" onClick={() => nav(`/activity/${record.id}`)}>详情</Button>
          {canManage && (record.status === 0 || record.status === 1) && (
            <Button type="link" size="small" onClick={() => { setEditing(record); setOpen(true); }}>编辑</Button>
          )}
          {canManage && record.status === 1 && (
            <Button type="link" size="small" onClick={() => startActivity(record.id).then(load)}>开始</Button>
          )}
          {canManage && record.status === 2 && (
            <Button type="link" size="small" onClick={() => endActivity(record.id).then(load)}>结束</Button>
          )}
          {canManage && (record.status === 0 || record.status === 1 || record.status === 2) && (
            <Button type="link" size="small" onClick={() => cancelActivity(record.id).then(load)}>取消</Button>
          )}
          {canManage && (record.status === 0 || record.status === 3 || record.status === 4) && (
            <Button type="link" size="small" danger onClick={() => deleteActivity(record.id).then(load)}>删除</Button>
          )}
        </Space>
      ),
    },
  ];

  return (
    <Card
      title="活动列表"
      extra={canCreate && (
        <Button type="primary" icon={<PlusOutlined />} onClick={() => { setEditing(null); setOpen(true); }}>
          新建活动
        </Button>
      )}
    >
      <Space style={{ marginBottom: 16 }}>
        <Input.Search allowClear placeholder="搜索标题/地点" onSearch={(v) => { setKeyword(v); setPage(1); }} />
        <Select
          allowClear
          placeholder="状态"
          style={{ width: 140 }}
          value={status}
          onChange={(v) => { setStatus(v); setPage(1); }}
          options={Object.entries(ActivityStatusMap).map(([k, v]) => ({ value: Number(k), label: v.text }))}
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
        onCancel={() => setOpen(false)}
        onSubmit={async (values) => {
          if (editing) {
            await updateActivity(editing.id, values);
            message.success('已更新');
          } else {
            await createActivity({
              title: String(values.title),
              description: values.description ? String(values.description) : undefined,
              category: values.category ? String(values.category) : undefined,
              tags: Array.isArray(values.tags) ? (values.tags as string[]) : undefined,
              location: values.location ? String(values.location) : undefined,
              max_participants: values.max_participants != null ? Number(values.max_participants) : undefined,
              start_time: String(values.start_time),
              end_time: String(values.end_time),
            });
            message.success('已创建');
          }
          setOpen(false);
          await load();
        }}
      />
    </Card>
  );
};

export default ListPage;
