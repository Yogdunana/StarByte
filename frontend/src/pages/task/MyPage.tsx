import React, { useCallback, useEffect, useState } from 'react';
import { Card, Input, Table, Tabs } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import StatusTag from '@/components/StatusTag/StatusTag';
import { getMyCreated, getMyDone, getMyOverdue, getMyTodo } from '@/api/task';
import type { PageResponse, Task } from '@/types/api';
import { TaskPriorityMap, TaskStatusMap } from './meta';
import DetailDrawer from './DetailDrawer';

type TabKey = 'todo' | 'done' | 'created' | 'overdue';

const fetchers: Record<TabKey, (p: { page: number; page_size: number; keyword?: string }) => Promise<PageResponse<Task>>> = {
  todo: getMyTodo,
  done: getMyDone,
  created: getMyCreated,
  overdue: getMyOverdue,
};

const MyPage: React.FC = () => {
  const [tab, setTab] = useState<TabKey>('todo');
  const [list, setList] = useState<Task[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [keyword, setKeyword] = useState('');
  const [loading, setLoading] = useState(false);
  const [detailId, setDetailId] = useState<string | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const res = await fetchers[tab]({ page, page_size: 10, keyword });
      setList(res.list || []);
      setTotal(res.total);
    } finally {
      setLoading(false);
    }
  }, [tab, page, keyword]);

  useEffect(() => { void load(); }, [load]);

  const columns: ColumnsType<Task> = [
    { title: '标题', dataIndex: 'title', render: (v: string, r) => <a onClick={() => setDetailId(r.id)}>{v}</a> },
    { title: '优先级', dataIndex: 'priority', width: 80, render: (v: number) => <StatusTag status={v} mapping={TaskPriorityMap} /> },
    { title: '状态', dataIndex: 'status', width: 90, render: (v: number) => <StatusTag status={v} mapping={TaskStatusMap} /> },
    { title: '负责人', key: 'a', width: 100, render: (_, r) => r.assignee?.name || '-' },
    { title: '截止', dataIndex: 'due_date', width: 160, render: (v?: string) => v?.replace('T', ' ').slice(0, 16) || '-' },
  ];

  return (
    <Card title="我的任务">
      <Input.Search
        allowClear
        placeholder="搜索标题"
        style={{ width: 260, marginBottom: 12 }}
        onSearch={(v) => { setKeyword(v); setPage(1); }}
      />
      <Tabs
        activeKey={tab}
        onChange={(k) => { setTab(k as TabKey); setPage(1); }}
        items={[
          { key: 'todo', label: '待办' },
          { key: 'done', label: '已办' },
          { key: 'created', label: '我创建的' },
          { key: 'overdue', label: '超期' },
        ]}
      />
      <Table
        rowKey="id"
        loading={loading}
        columns={columns}
        dataSource={list}
        pagination={{ current: page, total, pageSize: 10, onChange: setPage }}
      />
      <DetailDrawer taskId={detailId} open={!!detailId} onClose={() => setDetailId(null)} onChanged={() => { void load(); }} />
    </Card>
  );
};

export default MyPage;
