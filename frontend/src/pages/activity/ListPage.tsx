import React, { useCallback, useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Button, Card, Input, Select, Space, Table, Tag, message } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { usePermission } from '@/hooks/usePermission';
import {
  getActivityList, deleteActivity, startActivity, endActivity, cancelActivity,
} from '@/api/activity';
import type { Activity, ActivityStatus } from '@/api/activity';
import { ActivityStatusMap, ActivityCategoryOptions } from './meta';
import FormModal from './FormModal';

const ListPage: React.FC = () => {
  const nav = useNavigate();
  const canCreate = usePermission('activity:create');
  const canUpdate = usePermission('activity:update');
  const canDelete = usePermission('activity:delete');
  const canManage = usePermission('activity:manage');
  const [list, setList] = useState<Activity[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize] = useState(10);
  const [status, setStatus] = useState<ActivityStatus | undefined>();
  const [category, setCategory] = useState<string>();
  const [keyword, setKeyword] = useState('');
  const [loading, setLoading] = useState(false);
  const [open, setOpen] = useState(false);
  const [editing, setEditing] = useState<Activity | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const res = await getActivityList({ page, page_size: pageSize, status, category, keyword });
      setList(res.list);
      setTotal(res.total);
    } finally {
      setLoading(false);
    }
  }, [page, pageSize, status, category, keyword]);

  useEffect(() => { void load(); }, [load]);

  const onEdit = (a: Activity) => { setEditing(a); setOpen(true); };
  const onCreate = () => { setEditing(null); setOpen(true); };

  const onDelete = async (a: Activity) => {
    await deleteActivity(a.id);
    message.success('删除成功');
    load();
  };

  const onStart = async (a: Activity) => {
    await startActivity(a.id);
    message.success('活动已开始');
    load();
  };

  const onEnd = async (a: Activity) => {
    await endActivity(a.id);
    message.success('活动已结束');
    load();
  };

  const onCancel = async (a: Activity) => {
    await cancelActivity(a.id);
    message.success('活动已取消');
    load();
  };

  const columns: ColumnsType<Activity> = [
    {
      title: '活动标题', dataIndex: 'title', key: 'title',
      render: (v, r) => <a onClick={() => nav(`/activity/${r.id}`)}>{v}</a>,
    },
    { title: '分类', dataIndex: 'category', key: 'category', width: 100 },
    {
      title: '状态', dataIndex: 'status', key: 'status', width: 90,
      render: (s: ActivityStatus) => {
        const m = ActivityStatusMap[s];
        return <Tag color={m.color}>{m.text}</Tag>;
      },
    },
    { title: '时间', dataIndex: 'start_time', key: 'start_time', width: 160 },
    { title: '地点', dataIndex: 'location', key: 'location', width: 120 },
    {
      title: '报名/上限', key: 'participants', width: 110,
      render: (_, r) => (
        <span>{r.registered_count}/{r.max_participants > 0 ? r.max_participants : '不限'}</span>
      ),
    },
    {
      title: '操作', key: 'action', width: 240, fixed: 'right',
      render: (_, r) => (
        <Space size={4}>
          {canUpdate && r.status === 1 && (
            <Button size="small" type="link" onClick={() => onStart(r)}>开始</Button>
          )}
          {canUpdate && r.status === 2 && (
            <Button size="small" type="link" onClick={() => onEnd(r)}>结束</Button>
          )}
          {canUpdate && (r.status === 1 || r.status === 2) && (
            <Button size="small" type="link" danger onClick={() => onCancel(r)}>取消</Button>
          )}
          {canUpdate && (r.status === 0 || r.status === 1) && (
            <Button size="small" type="link" onClick={() => onEdit(r)}>编辑</Button>
          )}
          {canDelete && r.status !== 2 && (
            <Button size="small" type="link" danger onClick={() => onDelete(r)}>删除</Button>
          )}
          {canManage && (
            <Button size="small" type="link" onClick={() => nav(`/activity/${r.id}`)}>管理</Button>
          )}
        </Space>
      ),
    },
  ];

  return (
    <Card>
      <Space style={{ marginBottom: 16 }} wrap>
        {canCreate && (
          <Button type="primary" icon={<PlusOutlined />} onClick={onCreate}>新建活动</Button>
        )}
        <Select
          placeholder="状态" allowClear style={{ width: 120 }}
          value={status} onChange={(v) => { setStatus(v); setPage(1); }}
          options={Object.entries(ActivityStatusMap).map(([k, v]) => ({ value: Number(k), label: v.text }))}
        />
        <Select
          placeholder="分类" allowClear style={{ width: 120 }}
          value={category} onChange={(v) => { setCategory(v); setPage(1); }}
          options={ActivityCategoryOptions}
        />
        <Input.Search
          placeholder="搜索标题/地点" allowClear style={{ width: 220 }}
          onSearch={(v) => { setKeyword(v); setPage(1); }}
        />
      </Space>
      <Table
        rowKey="id" columns={columns} dataSource={list} loading={loading}
        pagination={{ current: page, pageSize, total, showSizeChanger: false }}
        onChange={(p) => setPage(p.current || 1)}
      />
      <FormModal
        open={open} editing={editing}
        onClose={() => setOpen(false)}
        onSaved={() => { setOpen(false); load(); }}
      />
    </Card>
  );
};

export default ListPage;
