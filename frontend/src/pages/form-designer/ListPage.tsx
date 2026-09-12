import { tx, useLocale } from '@/i18n/text';
import React, { useCallback, useEffect, useState } from 'react';
import { Button, Card, Input, Select, Space, Table, Tag, message } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { PlusOutlined, ReloadOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { listForms, updateForm, type FormListItem } from '@/api/forms';
import { usePermission } from '@/hooks/usePermission';

const statusMeta: Record<number, { color: string; text: string }> = {
  0: {
    color: 'default',
    get text() {
      return tx('草稿');
    },
  },
  1: {
    color: 'green',
    get text() {
      return tx('已发布');
    },
  },
  2: {
    color: 'red',
    get text() {
      return tx('已停用');
    },
  },
};

const ListPage: React.FC = () => {
  useLocale();
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

  const load = useCallback(
    async (p = page) => {
      setLoading(true);
      try {
        const res = await listForms({
          page: p,
          page_size: 10,
          keyword: keyword || undefined,
          status,
        });
        setList(res?.list || []);
        setTotal(res?.total || 0);
      } finally {
        setLoading(false);
      }
    },
    [page, keyword, status],
  );

  useEffect(() => {
    void load();
  }, [load]);

  const setStatusOf = async (row: FormListItem, next: 0 | 1 | 2) => {
    await updateForm(row.id, { status: next });
    message.success(tx('状态已更新'));
    void load();
  };

  const columns: ColumnsType<FormListItem> = [
    { title: tx('名称'), dataIndex: 'name' },
    { title: tx('说明'), dataIndex: 'description', ellipsis: true },
    {
      title: tx('状态'),
      dataIndex: 'status',
      width: 100,
      render: (s: number) => <Tag color={statusMeta[s]?.color}>{statusMeta[s]?.text || s}</Tag>,
    },
    { title: tx('提交数'), dataIndex: 'submission_count', width: 90 },
    {
      title: tx('操作'),
      width: 280,
      render: (_, row) => (
        <Space wrap>
          {canWrite ? (
            <Button type="link" size="small" onClick={() => nav(`/forms/designer/${row.id}`)}>
              {tx('设计')}
            </Button>
          ) : null}
          {canSubmit && row.status === 1 ? (
            <Button type="link" size="small" onClick={() => nav(`/forms/${row.id}/fill`)}>
              {tx('填写')}
            </Button>
          ) : null}
          {canRead ? (
            <Button type="link" size="small" onClick={() => nav(`/forms/${row.id}/submissions`)}>
              {tx('记录')}
            </Button>
          ) : null}
          {canWrite && row.status !== 1 ? (
            <Button
              type="link"
              size="small"
              onClick={() => {
                void setStatusOf(row, 1);
              }}
            >
              {tx('发布')}
            </Button>
          ) : null}
          {canWrite && row.status === 1 ? (
            <Button
              type="link"
              size="small"
              onClick={() => {
                void setStatusOf(row, 2);
              }}
            >
              {tx('停用')}
            </Button>
          ) : null}
        </Space>
      ),
    },
  ];

  return (
    <Card
      title={tx('动态表单')}
      extra={
        <Space>
          <Input.Search
            allowClear
            placeholder={tx('搜索名称')}
            onSearch={(v) => {
              setKeyword(v);
              setPage(1);
            }}
            style={{ width: 200 }}
          />
          <Select
            allowClear
            placeholder={tx('状态')}
            style={{ width: 120 }}
            value={status}
            onChange={(v) => {
              setStatus(v);
              setPage(1);
            }}
            options={[
              { value: 0, label: tx('草稿') },
              { value: 1, label: tx('已发布') },
              { value: 2, label: tx('已停用') },
            ]}
          />
          <Button
            icon={<ReloadOutlined />}
            onClick={() => {
              void load();
            }}
          >
            {tx('刷新')}
          </Button>
          {canWrite ? (
            <Button type="primary" icon={<PlusOutlined />} onClick={() => nav('/forms/designer')}>
              {tx('新建表单')}
            </Button>
          ) : null}
        </Space>
      }
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
