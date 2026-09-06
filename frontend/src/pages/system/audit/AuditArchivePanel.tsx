import React, { useCallback, useEffect, useState } from 'react';
import { Button, Drawer, Input, Table, Tag, message } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import dayjs from 'dayjs';
import {
  getAuditArchives,
  pullAuditArchive,
  type AuditArchiveItem,
  type AuditLogItem,
} from '@/api/audit';
import { actionColorMap } from './auditColumns';

const AuditArchivePanel: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [list, setList] = useState<AuditArchiveItem[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [open, setOpen] = useState(false);
  const [pulling, setPulling] = useState(false);
  const [keyword, setKeyword] = useState('');
  const [current, setCurrent] = useState<AuditArchiveItem | null>(null);
  const [records, setRecords] = useState<AuditLogItem[]>([]);
  const [recordTotal, setRecordTotal] = useState(0);
  const [recordPage, setRecordPage] = useState(1);

  const load = useCallback(async (p: number, ps: number) => {
    setLoading(true);
    try {
      const res = await getAuditArchives({ page: p, page_size: ps });
      setList(res.list);
      setTotal(res.total);
    } catch {
      message.error('加载归档列表失败');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    load(page, pageSize);
  }, [page, pageSize, load]);

  const pull = async (row: AuditArchiveItem, p = 1, kw = keyword) => {
    setCurrent(row);
    setOpen(true);
    setPulling(true);
    try {
      const res = await pullAuditArchive({ id: row.id, page: p, page_size: 20, keyword: kw || undefined });
      setRecords(res.list);
      setRecordTotal(res.total);
      setRecordPage(p);
      if (res.truncated) {
        message.warning('归档对象过大，结果可能被截断');
      }
    } catch {
      message.error('拉取归档对象失败');
    } finally {
      setPulling(false);
    }
  };

  const columns: ColumnsType<AuditArchiveItem> = [
    { title: '归档日期', dataIndex: 'archive_date', width: 120 },
    { title: '记录数', dataIndex: 'record_count', width: 90 },
    { title: 'MinIO 对象', dataIndex: 'minio_object', ellipsis: true },
    {
      title: '时间',
      dataIndex: 'created_at',
      width: 170,
      render: (t: string) => (t ? dayjs(t).format('YYYY-MM-DD HH:mm:ss') : '-'),
    },
    {
      title: '操作',
      width: 90,
      render: (_, row) => (
        <Button type="link" size="small" onClick={() => pull(row, 1, '')}>
          拉取
        </Button>
      ),
    },
  ];

  const recordColumns: ColumnsType<AuditLogItem> = [
    {
      title: '时间',
      dataIndex: 'timestamp',
      width: 170,
      render: (t: string) => (t ? dayjs(t).format('YYYY-MM-DD HH:mm:ss') : '-'),
    },
    { title: '用户', render: (_, r) => r.user?.username || '-' },
    {
      title: '动作',
      dataIndex: 'action',
      render: (a: string) => <Tag color={actionColorMap[a] || 'default'}>{a}</Tag>,
    },
    { title: '路径', dataIndex: 'path', ellipsis: true },
  ];

  return (
    <>
      <Table
        rowKey="id"
        loading={loading}
        columns={columns}
        dataSource={list}
        size="middle"
        pagination={{
          current: page,
          pageSize,
          total,
          onChange: (p, ps) => {
            setPage(p);
            setPageSize(ps);
          },
        }}
      />
      <Drawer
        title={current ? `归档 ${current.archive_date}` : '归档内容'}
        open={open}
        width={720}
        onClose={() => setOpen(false)}
      >
        <Input.Search
          placeholder="按路径/用户/模块筛选"
          allowClear
          value={keyword}
          onChange={(e) => setKeyword(e.target.value)}
          onSearch={(v) => current && pull(current, 1, v)}
          style={{ marginBottom: 12 }}
        />
        <Table
          rowKey="id"
          loading={pulling}
          columns={recordColumns}
          dataSource={records}
          size="small"
          pagination={{
            current: recordPage,
            pageSize: 20,
            total: recordTotal,
            onChange: (p) => current && pull(current, p, keyword),
          }}
        />
      </Drawer>
    </>
  );
};

export default AuditArchivePanel;
