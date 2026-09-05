import React, { useCallback, useEffect, useState } from 'react';
import { Button, Card, Input, Select, Space, Switch, Table, Tag, message } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import StatusTag from '@/components/StatusTag/StatusTag';
import { usePermission } from '@/hooks/usePermission';
import {
  createInternship,
  deleteInternship,
  getInternshipConfig,
  getInternshipList,
  updateInternship,
  updateInternshipConfig,
} from '@/api/internship';
import type { Internship, InternshipConfig, InternshipStatus, InternshipType } from '@/types/api';
import DetailDrawer from './DetailDrawer';
import FormDrawer from './FormDrawer';
import { InternshipStatusMap, InternshipTypeMap } from './meta';
import './internship.css';

const ListPage: React.FC = () => {
  const canCreate = usePermission('internship:create');
  const canUpdate = usePermission('internship:update');
  const canDelete = usePermission('internship:delete');
  const canConfig = usePermission('system:config');
  const [list, setList] = useState<Internship[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [keyword, setKeyword] = useState('');
  const [status, setStatus] = useState<InternshipStatus | undefined>();
  const [type, setType] = useState<InternshipType | undefined>();
  const [loading, setLoading] = useState(false);
  const [open, setOpen] = useState(false);
  const [editing, setEditing] = useState<Internship | null>(null);
  const [detailId, setDetailId] = useState<string | null>(null);
  const [config, setConfig] = useState<InternshipConfig | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const res = await getInternshipList({ page, page_size: 10, keyword, status, type });
      setList(res.list || []);
      setTotal(res.total);
    } finally {
      setLoading(false);
    }
  }, [page, keyword, status, type]);

  useEffect(() => { void load(); }, [load]);
  useEffect(() => {
    void getInternshipConfig().then(setConfig).catch(() => undefined);
  }, []);

  const columns: ColumnsType<Internship> = [
    {
      title: '项目',
      dataIndex: 'title',
      render: (v: string, r) => <Button type="link" onClick={() => setDetailId(r.id)}>{v}</Button>,
    },
    { title: '成员', key: 'user', width: 100, render: (_, r) => r.user?.name || '-' },
    { title: '单位', dataIndex: 'organization' },
    { title: '类型', dataIndex: 'type', width: 100, render: (v: number) => <StatusTag status={v} mapping={InternshipTypeMap} /> },
    { title: '时长', dataIndex: 'duration_days', width: 80, render: (v: number) => `${v} 天` },
    {
      title: '技能',
      dataIndex: 'skills',
      width: 180,
      render: (skills: string[]) => (skills || []).map((s) => <Tag key={s}>{s}</Tag>),
    },
    { title: '状态', dataIndex: 'status', width: 90, render: (v: number) => <StatusTag status={v} mapping={InternshipStatusMap} /> },
    {
      title: '操作',
      width: 150,
      render: (_, record) => (
        <Space>
          {canUpdate && record.status === 0 && (
            <Button type="link" size="small" onClick={() => { setEditing(record); setOpen(true); }}>编辑</Button>
          )}
          {canDelete && (
            <Button type="link" size="small" danger onClick={() => {
              void deleteInternship(record.id).then(() => { message.success('已删除'); void load(); });
            }}>删除</Button>
          )}
        </Space>
      ),
    },
  ];

  return (
    <div>
      <div className="intern-hero">
        <div>
          <h2>IT 实习档案</h2>
          <p>登记校内社团与校外实习，按时长统计、排名，方便招新后的干事培养。</p>
        </div>
        {canCreate && (
          <Button type="primary" size="large" icon={<PlusOutlined />} onClick={() => { setEditing(null); setOpen(true); }}>
            登记实习
          </Button>
        )}
      </div>
      <Card className="page-shell">
        {canConfig && config && (
          <Space style={{ marginBottom: 16 }} wrap>
            <span>学生可改</span>
            <Switch checked={config.allow_student_edit} onChange={(v) => {
              void updateInternshipConfig({ allow_student_edit: v }).then(setConfig);
            }} />
            <span>部长可改</span>
            <Switch checked={config.allow_minister_edit} onChange={(v) => {
              void updateInternshipConfig({ allow_minister_edit: v }).then(setConfig);
            }} />
            <span>排行榜公开</span>
            <Switch checked={config.ranking_visible} onChange={(v) => {
              void updateInternshipConfig({ ranking_visible: v }).then(setConfig);
            }} />
          </Space>
        )}
        <Space style={{ marginBottom: 16 }} wrap>
          <Input.Search allowClear placeholder="搜索项目/单位/姓名" onSearch={(v) => { setKeyword(v); setPage(1); }} />
          <Select
            allowClear placeholder="状态" style={{ width: 120 }} value={status}
            onChange={(v) => { setStatus(v); setPage(1); }}
            options={Object.entries(InternshipStatusMap).map(([k, v]) => ({ value: Number(k), label: v.text }))}
          />
          <Select
            allowClear placeholder="类型" style={{ width: 120 }} value={type}
            onChange={(v) => { setType(v); setPage(1); }}
            options={Object.entries(InternshipTypeMap).map(([k, v]) => ({ value: Number(k), label: v.text }))}
          />
        </Space>
        <Table
          rowKey="id"
          loading={loading}
          columns={columns}
          dataSource={list}
          pagination={{ current: page, pageSize: 10, total, onChange: setPage }}
        />
      </Card>
      <FormDrawer
        open={open}
        editing={editing}
        onClose={() => setOpen(false)}
        onSubmit={async (values) => {
          if (editing) {
            await updateInternship(editing.id, values);
            message.success('已更新');
          } else {
            await createInternship(values);
            message.success('已创建');
          }
          setOpen(false);
          void load();
        }}
      />
      <DetailDrawer id={detailId} onClose={() => setDetailId(null)} onChanged={() => void load()} />
    </div>
  );
};

export default ListPage;
