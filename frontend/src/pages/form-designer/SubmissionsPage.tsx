import { tx, useLocale } from '@/i18n/text';
import React, { useCallback, useEffect, useState } from 'react';
import { Button, Card, Table } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { useNavigate, useParams } from 'react-router-dom';
import { getForm, listFormSubmissions, type FormSubmissionItem } from '@/api/forms';

const SubmissionsPage: React.FC = () => {
  useLocale();
  const { id } = useParams();
  const nav = useNavigate();
  const [title, setTitle] = useState(tx('提交记录'));
  const [list, setList] = useState<FormSubmissionItem[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(false);

  const load = useCallback(
    async (p = page) => {
      if (!id) return;
      setLoading(true);
      try {
        const res = await listFormSubmissions(id, { page: p, page_size: 10 });
        setList(res?.list || []);
        setTotal(res?.total || 0);
      } finally {
        setLoading(false);
      }
    },
    [id, page],
  );

  useEffect(() => {
    if (!id) return;
    void getForm(id).then((f) => setTitle(tx('{{value0}} · 提交记录', { value0: f.name })));
  }, [id]);
  useEffect(() => {
    void load();
  }, [load]);

  const columns: ColumnsType<FormSubmissionItem> = [
    {
      title: tx('提交人'),
      dataIndex: ['submitted_by', 'name'],
      render: (_, r) => r.submitted_by?.name || r.submitted_by?.id || '-',
    },
    { title: tx('时间'), dataIndex: 'submitted_at', width: 200 },
    {
      title: tx('数据'),
      dataIndex: 'data',
      render: (data: Record<string, unknown>) => (
        <pre style={{ margin: 0, maxHeight: 120, overflow: 'auto' }}>
          {JSON.stringify(data, null, 2)}
        </pre>
      ),
    },
  ];

  return (
    <Card title={title} extra={<Button onClick={() => nav('/forms')}>{tx('返回列表')}</Button>}>
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

export default SubmissionsPage;
