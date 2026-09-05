import React, { useCallback, useEffect, useState } from 'react';
import { Button, Card, Input, Select, Space, Table, Tag, message } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import StatusTag from '@/components/StatusTag/StatusTag';
import { usePermission } from '@/hooks/usePermission';
import { createTask, deleteTask, getTaskList, getTaskStats, updateTask } from '@/api/task';
import type { Task, TaskPriority, TaskStats, TaskStatus } from '@/types/api';
import { TaskPriorityMap, TaskStatusMap } from './meta';
import FormModal from './FormModal';
import DetailDrawer from './DetailDrawer';

const ListPage: React.FC = () => {
  const canCreate = usePermission('task:create');
  const canUpdate = usePermission('task:update');
  const canDelete = usePermission('task:delete');
  const [list, setList] = useState<Task[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [status, setStatus] = useState<TaskStatus | undefined>();
  const [priority, setPriority] = useState<TaskPriority | undefined>();
  const [keyword, setKeyword] = useState('');
  const [sortBy, setSortBy] = useState('created_at');
  const [loading, setLoading] = useState(false);
  const [open, setOpen] = useState(false);
  const [editing, setEditing] = useState<Task | null>(null);
  const [detailId, setDetailId] = useState<string | null>(null);
  const [stats, setStats] = useState<TaskStats | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const res = await getTaskList({
        page, page_size: 10, status, priority, keyword, sort_by: sortBy, sort_order: 'desc',
      });
      setList(res.list || []);
      setTotal(res.total);
      setStats(await getTaskStats());
    } finally {
      setLoading(false);
    }
  }, [page, status, priority, keyword, sortBy]);

  useEffect(() => { void load(); }, [load]);

  const columns: ColumnsType<Task> = [
    { title: '标题', dataIndex: 'title', render: (v: string, r) => <Button type="link" onClick={() => setDetailId(r.id)}>{v}</Button> },
    { title: '优先级', dataIndex: 'priority', width: 80, render: (v: number) => <StatusTag status={v} mapping={TaskPriorityMap} /> },
    { title: '负责人', key: 'a', width: 100, render: (_, r) => r.assignee?.name || '-' },
    { title: '创建人', key: 'c', width: 100, render: (_, r) => r.creator?.name || '-' },
    { title: '截止', dataIndex: 'due_date', width: 160, render: (v?: string) => v?.replace('T', ' ').slice(0, 16) || '-' },
    { title: '标签', dataIndex: 'tags', width: 160, render: (tags: string[]) => (tags || []).map((t) => <Tag key={t}>{t}</Tag>) },
    { title: '状态', dataIndex: 'status', width: 90, render: (v: number) => <StatusTag status={v} mapping={TaskStatusMap} /> },
    {
      title: '操作',
      width: 160,
      render: (_, record) => (
        <Space>
          {canUpdate && record.status !== 2 && record.status !== 3 && (
            <Button type="link" size="small" onClick={() => { setEditing(record); setOpen(true); }}>编辑</Button>
          )}
          {canDelete && record.status !== 1 && (
            <Button type="link" size="small" danger onClick={() => deleteTask(record.id).then(() => { message.success('已删除'); void load(); })}>删除</Button>
          )}
        </Space>
      ),
    },
  ];

  return (
    <Card
      title="任务列表"
      extra={canCreate && (
        <Button type="primary" icon={<PlusOutlined />} onClick={() => { setEditing(null); setOpen(true); }}>
          新建任务
        </Button>
      )}
    >
      {stats && (
        <Space style={{ marginBottom: 12 }}>
          <Tag>全部 {stats.total}</Tag>
          <Tag color="red">超期 {stats.overdue}</Tag>
          <Tag color="blue">进行中 {stats.by_status?.doing || 0}</Tag>
          <Tag color="green">已完成 {stats.by_status?.done || 0}</Tag>
        </Space>
      )}
      <Space style={{ marginBottom: 16 }} wrap>
        <Input.Search allowClear placeholder="搜索标题" onSearch={(v) => { setKeyword(v); setPage(1); }} />
        <Select
          allowClear placeholder="状态" style={{ width: 120 }} value={status}
          onChange={(v) => { setStatus(v); setPage(1); }}
          options={Object.entries(TaskStatusMap).map(([k, v]) => ({ value: Number(k), label: v.text }))}
        />
        <Select
          allowClear placeholder="优先级" style={{ width: 120 }} value={priority}
          onChange={(v) => { setPriority(v); setPage(1); }}
          options={Object.entries(TaskPriorityMap).map(([k, v]) => ({ value: Number(k), label: v.text }))}
        />
        <Select
          value={sortBy} style={{ width: 140 }} onChange={setSortBy}
          options={[
            { value: 'created_at', label: '按创建时间' },
            { value: 'due_date', label: '按截止日期' },
            { value: 'priority', label: '按优先级' },
            { value: 'status', label: '按状态' },
          ]}
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
            await updateTask(editing.id, {
              title: String(values.title),
              description: values.description ? String(values.description) : undefined,
              priority: Number(values.priority) as TaskPriority,
              due_date: values.due_date ? String(values.due_date) : undefined,
              tags: Array.isArray(values.tags) ? values.tags.map(String) : undefined,
            });
            message.success('已更新');
          } else {
            await createTask({
              title: String(values.title),
              description: values.description ? String(values.description) : undefined,
              priority: Number(values.priority) as TaskPriority,
              assignee_id: values.assignee_id ? String(values.assignee_id) : undefined,
              department_id: values.department_id ? String(values.department_id) : undefined,
              due_date: values.due_date ? String(values.due_date) : undefined,
              tags: Array.isArray(values.tags) ? values.tags.map(String) : undefined,
            });
            message.success('已创建');
          }
          setOpen(false);
          await load();
        }}
      />
      <DetailDrawer taskId={detailId} open={!!detailId} onClose={() => setDetailId(null)} onChanged={() => { void load(); }} />
    </Card>
  );
};

export default ListPage;
