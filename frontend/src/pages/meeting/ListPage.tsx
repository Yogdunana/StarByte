import React, { useCallback, useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Button, Card, Input, Select, Space, Table, message } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import StatusTag from '@/components/StatusTag/StatusTag';
import { usePermission } from '@/hooks/usePermission';
import {
  cancelMeeting, createMeeting, deleteMeeting, endMeeting, getMeetingList, startMeeting, updateMeeting,
} from '@/api/meeting';
import type { Meeting, MeetingStatus } from '@/types/api';
import { MeetingStatusMap, MeetingTypeMap } from './meta';
import FormModal from './FormModal';

const ListPage: React.FC = () => {
  const nav = useNavigate();
  const canCreate = usePermission('meeting:create');
  const canManage = usePermission('meeting:manage');
  const [list, setList] = useState<Meeting[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [status, setStatus] = useState<MeetingStatus | undefined>();
  const [keyword, setKeyword] = useState('');
  const [loading, setLoading] = useState(false);
  const [open, setOpen] = useState(false);
  const [editing, setEditing] = useState<Meeting | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const res = await getMeetingList({ page, page_size: 10, status, keyword });
      setList(res.list);
      setTotal(res.total);
    } finally {
      setLoading(false);
    }
  }, [page, status, keyword]);

  useEffect(() => { void load(); }, [load]);

  const columns: ColumnsType<Meeting> = [
    { title: '标题', dataIndex: 'title' },
    { title: '类型', dataIndex: 'meeting_type', width: 100, render: (v: number) => MeetingTypeMap[v] || '-' },
    { title: '地点', dataIndex: 'location', render: (v?: string) => v || '-' },
    { title: '开始', dataIndex: 'start_time', width: 180, render: (v: string) => v?.replace('T', ' ').slice(0, 16) },
    { title: '组织者', key: 'org', width: 100, render: (_, r) => r.organizer?.name || '-' },
    { title: '签到', key: 'n', width: 90, render: (_, r) => `${r.checked_in_count}/${r.attendee_count}` },
    { title: '状态', dataIndex: 'status', width: 90, render: (v: number) => <StatusTag status={v} mapping={MeetingStatusMap} /> },
    {
      title: '操作',
      width: 280,
      render: (_, record) => (
        <Space wrap>
          <Button type="link" size="small" onClick={() => nav(`/meeting/${record.id}`)}>详情</Button>
          {canManage && record.status === 0 && (
            <Button type="link" size="small" onClick={() => { setEditing(record); setOpen(true); }}>编辑</Button>
          )}
          {canManage && record.status === 0 && (
            <Button type="link" size="small" onClick={() => startMeeting(record.id).then(load)}>开始</Button>
          )}
          {canManage && record.status === 1 && (
            <Button type="link" size="small" onClick={() => endMeeting(record.id).then(load)}>结束</Button>
          )}
          {canManage && (record.status === 0 || record.status === 1) && (
            <Button type="link" size="small" onClick={() => cancelMeeting(record.id).then(load)}>取消</Button>
          )}
          {canManage && (record.status === 0 || record.status === 3) && (
            <Button type="link" size="small" danger onClick={() => deleteMeeting(record.id).then(load)}>删除</Button>
          )}
        </Space>
      ),
    },
  ];

  return (
    <Card
      title="会议列表"
      extra={canCreate && (
        <Button type="primary" icon={<PlusOutlined />} onClick={() => { setEditing(null); setOpen(true); }}>
          新建会议
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
          options={Object.entries(MeetingStatusMap).map(([k, v]) => ({ value: Number(k), label: v.text }))}
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
            await updateMeeting(editing.id, values);
            message.success('已更新');
          } else {
            await createMeeting({
              title: String(values.title),
              description: values.description ? String(values.description) : undefined,
              meeting_type: Number(values.meeting_type),
              start_time: String(values.start_time),
              end_time: String(values.end_time),
              location: values.location ? String(values.location) : undefined,
              online_link: values.online_link ? String(values.online_link) : undefined,
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
