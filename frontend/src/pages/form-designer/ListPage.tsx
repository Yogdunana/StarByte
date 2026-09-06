import React, { useCallback, useEffect, useState } from 'react';
import { Button, Card, Input, Select, Space, Table, Tag, message } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { PlusOutlined, ReloadOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { listForms, updateForm, type FormListItem } from '@/api/forms';
import { usePermission } from '@/hooks/usePermission';

const statusMeta: Record<number, { color: string; text: string }> = {
  0: { color: 'default', text: '草稿' },
  1: { color: 'green', text: '已发布' },
  2: { color: 'red', text: '已停用' },
};

const ListPage: React.FC = () => {
  const nav = useNavigate();
  const canWrite = usePermission('form:write');
  const canRead = usePermission('form:read');
  const canSubmit = usePermission('form:submit');
  const [list, setList] = useState<FormListItem[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [keyword, setKeyword] = useState('');
  const [status, setStatus] = useState<number | undefined>();
  const [loading, setLoading] = useState(false);

  const load = useCallback(async (p = page) => {
    setLoading(true);
    try {
      const res = await listForms({ page: p, page_size: 10, keyword: keyword || undefined, status });
      setList(res?.list || []);
      setTotal(res?.total || 0);
    } finally {
      setLoading(false);
    }
  }, [page, keyword, status]);

  useEffect(() => { void load(); }, [load]);

  const setStatusOf = async (row: FormListItem, next: 0 | 1 | 2) => {
    await updateForm(row.id, { status: next });
    message.success('状态已更新');
    void load();
  };

  const columns: ColumnsType<FormListItem> = [
    { title: '名称', dataIndex: 'name' },
    { title: '说明', dataIndex: 'description', ellipsis: true },
    {
      title: '状态',
      dataIndex: 'status',
      width: 100,
      render: (s: number) => <Tag color={statusMeta[s]?.color}>{statusMeta[s]?.text || s}</Tag>,
    },
    { title: '提交数', dataIndex: 'submission_count', width: 90 },
    {
      title: '操作',
      width: 280,
      render: (_, row) => (
        <Space wrap>
          {canWrite ? <Button type="link" size="small" onClick={() => nav(`/forms/designer/${row.id}`)}>设计</Button> : null}
          {canSubmit && row.status === 1 ? <Button type="link" size="small" onClick={() => nav(`/forms/${row.id}/fill`)}>填写</Button> : null}
          {canRead ? <Button type="link" size="small" onClick={() => nav(`/forms/${row.id}/submissions`)}>记录</Button> : null}
          {canWrite && row.status !== 1 ? <Button type="link" size="small" onClick={() => { void setStatusOf(row, 1); }}>发布</Button> : null}
          {canWrite && row.status === 1 ? <Button type="link" size="small" onClick={() => { void setStatusOf(row, 2); }}>停用</Button> : null}
        </Space>
      ),
    },
  ];

  return (
    <Card
      title="动态表单"
      extra={(
        <Space>
          <Input.Search allowClear placeholder="搜索名称" onSearch={(v) => { setKeyword(v); setPage(1); }} style={{ width: 200 }} />
          <Select allowClear placeholder="状态" style={{ width: 120 }} value={status} onChange={(v) => { setStatus(v); setPage(1); }}
            options={[{ value: 0, label: '草稿' }, { value: 1, label: '已发布' }, { value: 2, label: '已停用' }]} />
          <Button icon={<ReloadOutlined />} onClick={() => { void load(); }}>刷新</Button>
          {canWrite ? <Button type="primary" icon={<PlusOutlined />} onClick={() => nav('/forms/designer')}>新建表单</Button> : null}
        </Space>
      )}
    >
      <Table
        rowKey="id"
        loading={loading}
        columns={columns}
        dataSource={list}
        pagination={{ current: page, total, pageSize: 10, onChange: (p) => setPage(p) }}
      />
    </Card>
  );
};

export default ListPage;
